package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"DeepSight/internal/dto"
	"DeepSight/internal/model"
	"DeepSight/internal/repository"

	"gorm.io/gorm"
)

const (
	// TopK 检索的 chunk 数量
	TopK = 5
	// MaxHistoryMessages 最大历史消息数量
	MaxHistoryMessages = 10
)

type ChatService struct {
	convRepo  *repository.ConversationRepository
	msgRepo   *repository.MessageRepository
	chunkRepo *repository.ChunkRepository
	fileRepo  *repository.FileRepository
	kbRepo    *repository.KnowledgeBaseRepository
}

func NewChatService(
	convRepo *repository.ConversationRepository,
	msgRepo *repository.MessageRepository,
	chunkRepo *repository.ChunkRepository,
	fileRepo *repository.FileRepository,
	kbRepo *repository.KnowledgeBaseRepository,
) *ChatService {
	return &ChatService{
		convRepo:  convRepo,
		msgRepo:   msgRepo,
		chunkRepo: chunkRepo,
		fileRepo:  fileRepo,
		kbRepo:    kbRepo,
	}
}

// CreateConversation 创建新会话
func (s *ChatService) CreateConversation(userID, kbID uint, title string) (*model.Conversation, error) {
	// 验证知识库存在且属于用户
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

	conv := &model.Conversation{
		UserID:          userID,
		KnowledgeBaseID: kbID,
		Title:           title,
		LastMessageAt:   time.Now(),
	}
	if err := s.convRepo.Create(conv); err != nil {
		return nil, err
	}

	return conv, nil
}

// Chat 用户提问处理（RAG 流程）
func (s *ChatService) Chat(userID, convID uint, question string) (*dto.ChatResponse, error) {
	// 1. 获取会话，验证用户权限
	conv, err := s.convRepo.GetByID(convID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dto.ErrNotFound
		}
		return nil, err
	}
	if conv.UserID != userID {
		return nil, dto.ErrForbidden
	}

	// 2. 获取知识库的文件 hashes
	fileHashes, err := s.fileRepo.GetHashesByKnowledgeBaseID(conv.KnowledgeBaseID)
	if err != nil {
		return nil, err
	}
	if len(fileHashes) == 0 {
		return nil, errors.New("知识库中没有已处理的文件，请先上传文件")
	}

	// 3. 生成问题的 embedding
	questionEmbedding, err := LLM.Embedding(question)
	if err != nil {
		return nil, fmt.Errorf("生成问题向量失败: %w", err)
	}

	// 4. 向量检索：获取知识库相关的 chunks
	chunks, err := s.chunkRepo.SearchSimilar(questionEmbedding, fileHashes, TopK)
	if err != nil {
		return nil, fmt.Errorf("检索相关内容失败: %w", err)
	}

	// 5. 构建检索到的 chunks 信息（用于存储）
	retrievedChunksInfo := make([]map[string]interface{}, len(chunks))
	for i, chunk := range chunks {
		retrievedChunksInfo[i] = map[string]interface{}{
			"id":          chunk.ID,
			"content":     truncateContent(chunk.Content, 100),
			"chunk_index": chunk.ChunkIndex,
			"file_hash":   chunk.FileHash,
		}
	}
	retrievedChunksJSON, _ := json.Marshal(retrievedChunksInfo)

	// 6. 获取历史消息
	historyMessages, err := s.msgRepo.GetRecentByConversationID(convID, MaxHistoryMessages)
	if err != nil {
		return nil, err
	}

	// 7. 构建 prompt
	systemPrompt := buildSystemPrompt(chunks)
	chatMessages := buildChatMessages(historyMessages, question)

	// 8. 调用 LLM Chat
	answer, err := LLM.Chat(systemPrompt, chatMessages)
	if err != nil {
		return nil, fmt.Errorf("调用大模型失败: %w", err)
	}

	// 9. 保存用户消息
	userMsg := &model.Message{
		ConversationID:  convID,
		Role:            "user",
		Content:         question,
		RetrievedChunks: string(retrievedChunksJSON),
	}
	if err := s.msgRepo.Create(userMsg); err != nil {
		return nil, err
	}

	// 10. 保存助手消息
	assistantMsg := &model.Message{
		ConversationID: convID,
		Role:           "assistant",
		Content:        answer,
	}
	if err := s.msgRepo.Create(assistantMsg); err != nil {
		return nil, err
	}

	// 11. 更新会话的 LastMessageAt 和标题（如果首次对话）
	conv.LastMessageAt = time.Now()
	if conv.Title == "" {
		conv.Title = truncateContent(question, 20)
	}
	if err := s.convRepo.Update(conv); err != nil {
		return nil, err
	}

	return &dto.ChatResponse{
		UserMessage:      ToMessageResponse(userMsg),
		AssistantMessage: ToMessageResponse(assistantMsg),
	}, nil
}

// ChatStream 流式聊天，返回 token channel 和结果回调
// resultCallback 在流式完成后调用，用于保存消息到数据库
func (s *ChatService) ChatStream(ctx context.Context, userID, convID uint, question string) (<-chan string, func() (*dto.ChatResponse, error), error) {
	// 1. 获取会话，验证用户权限
	conv, err := s.convRepo.GetByID(convID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, dto.ErrNotFound
		}
		return nil, nil, err
	}
	if conv.UserID != userID {
		return nil, nil, dto.ErrForbidden
	}

	// 2. 获取知识库的文件 hashes
	fileHashes, err := s.fileRepo.GetHashesByKnowledgeBaseID(conv.KnowledgeBaseID)
	if err != nil {
		return nil, nil, err
	}
	if len(fileHashes) == 0 {
		return nil, nil, errors.New("知识库中没有已处理的文件，请先上传文件")
	}

	// 3. 生成问题的 embedding
	questionEmbedding, err := LLM.Embedding(question)
	if err != nil {
		return nil, nil, fmt.Errorf("生成问题向量失败: %w", err)
	}

	// 4. 向量检索：获取知识库相关的 chunks
	chunks, err := s.chunkRepo.SearchSimilar(questionEmbedding, fileHashes, TopK)
	if err != nil {
		return nil, nil, fmt.Errorf("检索相关内容失败: %w", err)
	}

	// 5. 构建检索到的 chunks 信息（用于存储）
	retrievedChunksInfo := make([]map[string]interface{}, len(chunks))
	for i, chunk := range chunks {
		retrievedChunksInfo[i] = map[string]interface{}{
			"id":          chunk.ID,
			"content":     truncateContent(chunk.Content, 100),
			"chunk_index": chunk.ChunkIndex,
			"file_hash":   chunk.FileHash,
		}
	}
	retrievedChunksJSON, _ := json.Marshal(retrievedChunksInfo)

	// 6. 获取历史消息
	historyMessages, err := s.msgRepo.GetRecentByConversationID(convID, MaxHistoryMessages)
	if err != nil {
		return nil, nil, err
	}

	// 7. 构建 prompt
	systemPrompt := buildSystemPrompt(chunks)
	chatMessages := buildChatMessages(historyMessages, question)

	// 8. 调用 LLM 流式 Chat
	stream, err := LLM.ChatStream(ctx, systemPrompt, chatMessages)
	if err != nil {
		return nil, nil, fmt.Errorf("调用大模型失败: %w", err)
	}

	// 用于收集完整响应
	var fullAnswer strings.Builder
	var once sync.Once
	var finalResponse *dto.ChatResponse
	var finalError error

	// 创建包装后的 stream，同时收集内容
	wrappedStream := make(chan string)

	go func() {
		defer close(wrappedStream)
		for token := range stream {
			// 检查是否是错误消息
			if after, found := strings.CutPrefix(token, "[ERROR]"); found {
				finalError = errors.New(after)
				return
			}
			fullAnswer.WriteString(token)
			wrappedStream <- token
		}
	}()

	// 结果回调函数，在流式完成后调用
	resultCallback := func() (*dto.ChatResponse, error) {
		once.Do(func() {
			if finalError != nil {
				finalResponse = nil
				return
			}

			answer := fullAnswer.String()

			// 9. 保存用户消息
			userMsg := &model.Message{
				ConversationID:  convID,
				Role:            "user",
				Content:         question,
				RetrievedChunks: string(retrievedChunksJSON),
			}
			if err := s.msgRepo.Create(userMsg); err != nil {
				finalError = err
				return
			}

			// 10. 保存助手消息
			assistantMsg := &model.Message{
				ConversationID: convID,
				Role:           "assistant",
				Content:        answer,
			}
			if err := s.msgRepo.Create(assistantMsg); err != nil {
				finalError = err
				return
			}

			// 11. 更新会话的 LastMessageAt 和标题
			conv.LastMessageAt = time.Now()
			if conv.Title == "" {
				conv.Title = truncateContent(question, 20)
			}
			if err := s.convRepo.Update(conv); err != nil {
				finalError = err
				return
			}

			finalResponse = &dto.ChatResponse{
				UserMessage:      ToMessageResponse(userMsg),
				AssistantMessage: ToMessageResponse(assistantMsg),
			}
		})

		return finalResponse, finalError
	}

	return wrappedStream, resultCallback, nil
}

// GetConversation 获取会话详情
func (s *ChatService) GetConversation(userID, convID uint, page, pageSize int) (*dto.ConversationDetailResponse, error) {
	conv, messages, total, err := s.convRepo.GetByIDWithMessages(convID, page, pageSize)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dto.ErrNotFound
		}
		return nil, err
	}
	if conv.UserID != userID {
		return nil, dto.ErrForbidden
	}

	// 反转消息顺序，使最新的消息在最后
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	msgResponses := make([]dto.MessageResponse, len(messages))
	for i, msg := range messages {
		msgResponses[i] = ToMessageResponse(&msg)
	}

	return &dto.ConversationDetailResponse{
		Conversation: ToConversationResponse(conv),
		Messages:     msgResponses,
		Total:        total,
	}, nil
}

// ListConversations 获取用户的会话列表
func (s *ChatService) ListConversations(userID uint, page, pageSize int) (*dto.ConversationListResponse, error) {
	page = normalizePage(page)
	pageSize = normalizePageSize(pageSize)

	convs, total, err := s.convRepo.GetByUserID(userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ConversationResponse, len(convs))
	for i, conv := range convs {
		responses[i] = ToConversationResponse(&conv)
	}

	return &dto.ConversationListResponse{
		Conversations: responses,
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
	}, nil
}

// DeleteConversation 删除会话
func (s *ChatService) DeleteConversation(userID, convID uint) error {
	conv, err := s.convRepo.GetByID(convID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ErrNotFound
		}
		return err
	}
	if conv.UserID != userID {
		return dto.ErrForbidden
	}

	return s.convRepo.Delete(convID)
}

// buildSystemPrompt 构建 system prompt，包含检索到的 chunks
func buildSystemPrompt(chunks []model.Chunk) string {
	contextParts := make([]string, len(chunks))
	for i, chunk := range chunks {
		contextParts[i] = fmt.Sprintf("[%d] %s", i+1, chunk.Content)
	}

	return fmt.Sprintf(`你是一个智能助手，基于提供的知识库内容回答用户问题。
请严格依据知识库内容回答，如果知识库中没有相关信息，请明确告知用户。
回答要准确、简洁、有帮助。

参考资料：
%s`, strings.Join(contextParts, "\n"))
}

// buildChatMessages 构建对话消息列表
func buildChatMessages(history []model.Message, question string) []ChatMessage {
	messages := make([]ChatMessage, 0, len(history)+1)

	// 添加历史消息
	for _, msg := range history {
		messages = append(messages, ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 添加当前问题
	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: question,
	})

	return messages
}

// truncateContent 截断内容，正确处理 UTF-8 编码
func truncateContent(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) <= maxLen {
		return content
	}
	return string(runes[:maxLen]) + "..."
}

// ToConversationResponse 转换会话响应
func ToConversationResponse(conv *model.Conversation) dto.ConversationResponse {
	return dto.ConversationResponse{
		ID:              conv.ID,
		KnowledgeBaseID: conv.KnowledgeBaseID,
		Title:           conv.Title,
		LastMessageAt:   conv.LastMessageAt.Format("2006-01-02 15:04:05"),
		CreatedAt:       conv.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       conv.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ToMessageResponse 转换消息响应
func ToMessageResponse(msg *model.Message) dto.MessageResponse {
	return dto.MessageResponse{
		ID:        msg.ID,
		Role:      msg.Role,
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
