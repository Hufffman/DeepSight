package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"DeepSight/internal/database"
	"DeepSight/internal/dto"
	"DeepSight/internal/model"
	"DeepSight/internal/repository"

	"gorm.io/gorm"
)

type FileService struct {
	fileRepo  *repository.FileRepository
	chunkRepo *repository.ChunkRepository
	kbRepo    *repository.KnowledgeBaseRepository
}

func NewFileService(fileRepo *repository.FileRepository, chunkRepo *repository.ChunkRepository, kbRepo *repository.KnowledgeBaseRepository) *FileService {
	return &FileService{
		fileRepo:  fileRepo,
		chunkRepo: chunkRepo,
		kbRepo:    kbRepo,
	}
}

func (s *FileService) UploadFile(kbID uint, fileName string, fileData io.Reader, fileSize int64) (*model.File, error) {
	// 生成文件hash
	hash := sha256.New()
	data, err := io.ReadAll(fileData)
	if err != nil {
		return nil, err
	}
	hash.Write(data)
	fileHash := hex.EncodeToString(hash.Sum(nil))

	// 如果文件hash已在数据库中，复用数据库中的文件内容再新增一条记录（考虑多个知识库共用一个文件），返回
	existingFiles, err := s.fileRepo.GetByHash(fileHash)
	if err == nil && len(existingFiles) > 0 {
		// 如果本知识库已存在这个文件，则不创建
		for _, f := range existingFiles {
			if f.KnowledgeBaseID == kbID {
				return &f, errors.New("文件已存在")
			}
		}

		existingFile := existingFiles[0]
		file := model.File{
			KnowledgeBaseID: kbID,
			FileName:        fileName,
			FileHash:        fileHash,
			FileSize:        fileSize,
			StorageKey:      existingFile.StorageKey,
			StorageURL:      existingFile.StorageURL,
			FileType:        existingFile.FileType,
			Status:          existingFile.Status,
			ParsedText:      existingFile.ParsedText,
			ParseError:      existingFile.ParseError,
		}
		if err := s.fileRepo.Create(&file); err != nil {
			return nil, err
		}
		// 清除知识库文件hash缓存
		_ = s.fileRepo.InvalidateKBFileHashesCache(kbID)
		return &file, nil
	}

	// 检查文件类型
	fileType := DetectFileType(fileName)
	if fileType == "" {
		return nil, errors.New("不支持的文件类型，必须是pdf、docx、md、txt中的一种")
	}

	// 将文件存入rustfs
	rustfs := database.GetRustFS()
	if rustfs == nil {
		return nil, errors.New("rustfs is not initialized")
	}

	contentType := mime.TypeByExtension(filepath.Ext(fileName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	storageKey := buildStorageKey(fileHash, fileName)
	storageKey, storageURL, err := rustfs.UploadObject(
		context.Background(),
		storageKey,
		bytes.NewReader(data),
		fileSize,
		contentType,
	)
	if err != nil {
		return nil, err
	}

	// 文件写入数据库，状态为pending，等待异步处理
	file := &model.File{
		KnowledgeBaseID: kbID,
		FileName:        fileName,
		FileHash:        fileHash,
		FileSize:        fileSize,
		StorageKey:      storageKey,
		StorageURL:      storageURL,
		FileType:        fileType,
		Status:          model.FileStatusPending,
	}
	if err := s.fileRepo.Create(file); err != nil {
		return nil, err
	}

	// 清除知识库文件hash缓存
	_ = s.fileRepo.InvalidateKBFileHashesCache(kbID)

	// 发送消息到 RabbitMQ 进行异步处理
	rmq := database.GetRabbitMQ()
	if rmq == nil {
		// RabbitMQ 未初始化，标记错误
		file.Status = model.FileStatusError
		file.ParseError = "RabbitMQ未初始化"
		_ = s.fileRepo.Update(file)
	}

	msg := map[string]interface{}{
		"file_id":     file.ID,
		"file_hash":   fileHash,
		"storage_key": storageKey,
		"file_type":   fileType,
		"file_name":   fileName,
		"kb_id":       kbID,
	}
	if err := rmq.PublishJSON(msg); err != nil {
		// 发送失败不影响上传，只记录错误
		file.Status = model.FileStatusError
		file.ParseError = "发送异步处理消息失败: " + err.Error()
		_ = s.fileRepo.Update(file)
	}

	return file, nil
}

func DetectFileType(filename string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
	switch ext {
	case "pdf":
		return "pdf"
	case "docx":
		return "docx"
	case "md":
		return "md"
	case "txt":
		return "txt"
	default:
		return ""
	}
}

func buildStorageKey(fileHash, fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	return "files/" + fileHash + ext
}

func (s *FileService) GetFileByID(id uint) (*model.File, error) {
	return s.fileRepo.GetByID(id)
}

func (s *FileService) GetFileByIDWithChunks(id uint) (*model.File, error) {
	return s.fileRepo.GetByIDWithChunks(id)
}

func (s *FileService) ListFiles(kbID uint, page, pageSize int) ([]model.File, int64, error) {
	page = normalizePage(page)
	pageSize = normalizePageSize(pageSize)
	return s.fileRepo.GetByKnowledgeBaseID(kbID, page, pageSize)
}

func (s *FileService) UpdateFileStatus(id uint, status string) error {
	file, err := s.fileRepo.GetByID(id)
	if err != nil {
		return err
	}

	file.Status = status
	return s.fileRepo.Update(file)
}

func (s *FileService) UpdateParsedText(id uint, parsedText string) error {
	file, err := s.fileRepo.GetByID(id)
	if err != nil {
		return err
	}

	file.ParsedText = parsedText
	file.Status = model.FileStatusParsed
	return s.fileRepo.Update(file)
}

func (s *FileService) DeleteFile(id uint) error {
	return s.fileRepo.Delete(id)
}

func (s *FileService) UploadFileForUser(userID, kbID uint, fileName string, fileData io.Reader, fileSize int64) (*dto.FileResponse, error) {
	if _, err := s.getOwnedKnowledgeBase(userID, kbID); err != nil {
		return nil, err
	}

	file, err := s.UploadFile(kbID, fileName, fileData, fileSize)
	if err != nil {
		return nil, err
	}

	response := ToFileResponse(file)
	return &response, nil
}

func (s *FileService) ListFilesForUser(userID, kbID uint, page, pageSize int) (*dto.FileListResponse, error) {
	if _, err := s.getOwnedKnowledgeBase(userID, kbID); err != nil {
		return nil, err
	}

	files, total, err := s.ListFiles(kbID, page, pageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.FileResponse, len(files))
	for i := range files {
		responses[i] = ToFileResponse(&files[i])
	}

	return &dto.FileListResponse{
		Files:    responses,
		Total:    total,
		Page:     normalizePage(page),
		PageSize: normalizePageSize(pageSize),
	}, nil
}

func (s *FileService) GetFileDetailForUser(userID, kbID, fileID uint) (*dto.FileDetailResponse, error) {
	if _, err := s.getOwnedKnowledgeBase(userID, kbID); err != nil {
		return nil, err
	}

	file, err := s.GetFileByIDWithChunks(fileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dto.ErrNotFound
		}
		return nil, err
	}
	if file.KnowledgeBaseID != kbID {
		return nil, dto.ErrForbidden
	}

	response := ToFileDetailResponse(file)
	return &response, nil
}

func (s *FileService) DeleteFileForUser(userID, kbID, fileID uint) error {
	if _, err := s.getOwnedKnowledgeBase(userID, kbID); err != nil {
		return err
	}

	file, err := s.GetFileByID(fileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ErrNotFound
		}
		return err
	}
	if file.KnowledgeBaseID != kbID {
		return dto.ErrForbidden
	}

	return s.DeleteFile(fileID)
}

func (s *FileService) getOwnedKnowledgeBase(userID, kbID uint) (*model.KnowledgeBase, error) {
	kb, err := s.kbRepo.GetByID(kbID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dto.ErrNotFound
		}
		return nil, err
	}
	if kb.UserID != userID {
		return nil, dto.ErrForbidden
	}

	return kb, nil
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizePageSize(pageSize int) int {
	if pageSize < 1 || pageSize > 100 {
		return 10
	}
	return pageSize
}

func ToFileResponse(file *model.File) dto.FileResponse {
	return dto.FileResponse{
		ID:              file.ID,
		KnowledgeBaseID: file.KnowledgeBaseID,
		FileName:        file.FileName,
		FileHash:        file.FileHash,
		FileSize:        file.FileSize,
		FileType:        file.FileType,
		Status:          file.Status,
		ParsedText:      file.ParsedText,
		StorageKey:      file.StorageKey,
		StorageURL:      file.StorageURL,
		CreatedAt:       file.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       file.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func ToFileDetailResponse(file *model.File) dto.FileDetailResponse {
	return dto.FileDetailResponse{
		ID:              file.ID,
		KnowledgeBaseID: file.KnowledgeBaseID,
		FileName:        file.FileName,
		FileHash:        file.FileHash,
		FileSize:        file.FileSize,
		FileType:        file.FileType,
		Status:          file.Status,
		ParsedText:      file.ParsedText,
		StorageKey:      file.StorageKey,
		StorageURL:      file.StorageURL,
		CreatedAt:       file.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       file.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
