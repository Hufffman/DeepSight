package worker

import (
	"DeepSight/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"DeepSight/internal/database"
	"DeepSight/internal/model"
	"DeepSight/internal/repository"
	"DeepSight/internal/util/chunker"
	"DeepSight/internal/util/parser"

	"github.com/pgvector/pgvector-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

type FileProcessMessage struct {
	FileID     uint   `json:"file_id"`
	FileHash   string `json:"file_hash"`
	StorageKey string `json:"storage_key"`
	FileType   string `json:"file_type"`
	FileName   string `json:"file_name"`
	KbID       uint   `json:"kb_id"`
}

type FileWorker struct {
	fileRepo  *repository.FileRepository
	chunkRepo *repository.ChunkRepository
}

func NewFileWorker(fileRepo *repository.FileRepository, chunkRepo *repository.ChunkRepository) *FileWorker {
	return &FileWorker{
		fileRepo:  fileRepo,
		chunkRepo: chunkRepo,
	}
}

func (w *FileWorker) Start() error {
	rmq := database.GetRabbitMQ()
	if rmq == nil {
		return fmt.Errorf("rabbitmq is not initialized")
	}

	msgs, err := rmq.ReceiveRouting()
	if err != nil {
		return fmt.Errorf("failed to receive from rabbitmq: %w", err)
	}

	go w.consumeMessages(msgs)
	return nil
}

func (w *FileWorker) consumeMessages(msgs <-chan amqp.Delivery) {
	for msg := range msgs {
		if err := w.handleMessage(msg); err != nil {
			log.Printf("failed to process file: %v", err)
			msg.Nack(false, false) // 不重新入队，避免无限循环
		} else {
			msg.Ack(false)
		}
	}
}

func (w *FileWorker) handleMessage(msg amqp.Delivery) error {
	var data FileProcessMessage
	if err := json.Unmarshal(msg.Body, &data); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return w.processFile(data)
}

func (w *FileWorker) processFile(data FileProcessMessage) error {
	// 更新状态为 processing
	file, err := w.fileRepo.GetByID(data.FileID)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}
	file.Status = model.FileStatusParsing
	if err := w.fileRepo.Update(file); err != nil {
		return fmt.Errorf("failed to update file status: %w", err)
	}

	// 从 rustfs 下载文件
	rustfs := database.GetRustFS()
	if rustfs == nil {
		w.updateFileError(data.FileID, "rustfs is not initialized")
		return fmt.Errorf("rustfs is not initialized")
	}

	fileData, err := rustfs.DownloadObject(context.Background(), data.StorageKey)
	if err != nil {
		w.updateFileError(data.FileID, "failed to download file: "+err.Error())
		return fmt.Errorf("failed to download file: %w", err)
	}

	// 解析文本
	var parsedText string
	switch data.FileType {
	case "pdf":
		parsedText, err = parser.ExtractPDFText(fileData)
		if err != nil {
			w.updateFileError(data.FileID, "PDF解析失败: "+err.Error())
			return fmt.Errorf("failed to parse pdf: %w", err)
		}
	case "docx":
		parsedText, err = parser.ExtractDocxText(fileData)
		if err != nil {
			w.updateFileError(data.FileID, "Word解析失败: "+err.Error())
			return fmt.Errorf("failed to parse docx: %w", err)
		}
	case "md", "txt":
		parsedText = string(fileData)
	default:
		w.updateFileError(data.FileID, "不支持的文件类型: "+data.FileType)
		return fmt.Errorf("unsupported file type: %s", data.FileType)
	}

	// 分块
	textChunker := chunker.DefaultChunker()
	textChunks := textChunker.Chunk(parsedText)

	// 生成 embedding
	total := len(textChunks)
	chunkTexts := make([]string, total)
	for i, chunk := range textChunks {
		chunkTexts[i] = chunk.Content
	}

	chunkEmbeddings := make([][]float32, total)
	waited := len(textChunks)
	for {
		if waited > service.LLM.EmbeddingBatchSize {
			ce, err := service.LLM.EmbeddingBatch(chunkTexts[total-waited : total-waited+service.LLM.EmbeddingBatchSize])
			if err != nil {
				w.updateFileError(data.FileID, "Embedding生成失败: "+err.Error())
				return fmt.Errorf("failed to generate embeddings: %w", err)
			}
			copy(chunkEmbeddings[total-waited:total-waited+service.LLM.EmbeddingBatchSize], ce)
			waited -= service.LLM.EmbeddingBatchSize
		} else {
			ce, err := service.LLM.EmbeddingBatch(chunkTexts[total-waited : total])
			if err != nil {
				w.updateFileError(data.FileID, "Embedding生成失败: "+err.Error())
				return fmt.Errorf("failed to generate embeddings: %w", err)
			}
			copy(chunkEmbeddings[total-waited:total], ce)
			break
		}
	}

	// 创建 Chunk 记录
	for i, tc := range textChunks {
		dbChunk := &model.Chunk{
			FileHash:    data.FileHash,
			ChunkIndex:  tc.Index,
			Content:     tc.Content,
			StartOffset: tc.StartOffset,
			EndOffset:   tc.EndOffset,
			Vector:      pgvector.NewVector(chunkEmbeddings[i]),
		}
		if err := w.chunkRepo.Create(dbChunk); err != nil {
			log.Printf("failed to create chunk %d: %v", tc.Index, err)
			continue
		}
	}

	// 更新 File 状态为 parsed
	file.Status = model.FileStatusParsed
	file.ParsedText = parsedText
	if err := w.fileRepo.Update(file); err != nil {
		return fmt.Errorf("failed to update file status: %w", err)
	}

	// 清除知识库文件hash缓存（确保缓存与数据库同步）
	_ = w.fileRepo.InvalidateKBFileHashesCache(data.KbID)

	log.Printf("file %d processed successfully", data.FileID)
	return nil
}

func (w *FileWorker) updateFileError(fileID uint, errMsg string) {
	file, err := w.fileRepo.GetByID(fileID)
	if err != nil {
		log.Printf("failed to get file for error update: %v", err)
		return
	}
	file.Status = model.FileStatusError
	file.ParseError = errMsg
	if err := w.fileRepo.Update(file); err != nil {
		log.Printf("failed to update file error status: %v", err)
	}
}
