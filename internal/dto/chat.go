package dto

// CreateConversationRequest 创建会话请求
type CreateConversationRequest struct {
	KnowledgeBaseID uint   `json:"knowledge_base_id" binding:"required"`
	Title           string `json:"title"`
}

// ChatRequest 发送消息请求
type ChatRequest struct {
	Question string `json:"question" binding:"required"`
}

// ConversationResponse 会话响应
type ConversationResponse struct {
	ID              uint   `json:"id"`
	KnowledgeBaseID uint   `json:"knowledge_base_id"`
	Title           string `json:"title"`
	LastMessageAt   string `json:"last_message_at"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	ID        uint   `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	UserMessage      MessageResponse `json:"user_message"`
	AssistantMessage MessageResponse `json:"assistant_message"`
}

// ConversationDetailResponse 会话详情响应
type ConversationDetailResponse struct {
	Conversation ConversationResponse `json:"conversation"`
	Messages     []MessageResponse    `json:"messages"`
	Total        int64                `json:"total"`
}

// ConversationListResponse 会话列表响应
type ConversationListResponse struct {
	Conversations []ConversationResponse `json:"conversations"`
	Total         int64                  `json:"total"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
}