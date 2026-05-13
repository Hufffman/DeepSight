package dto

// CreateKBRequest 创建知识库请求
type CreateKBRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateKBRequest 更新知识库请求
type UpdateKBRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// KBResponse 知识库响应
type KBResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UserID      uint   `json:"user_id"`
	FileCount   int64  `json:"file_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// KBDetailResponse 知识库详情响应（包含文件列表）
type KBDetailResponse struct {
	ID          uint          `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	UserID      uint          `json:"user_id"`
	Files       []FileSummary `json:"files"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
}

// FileSummary 文件摘要
type FileSummary struct {
	ID        uint   `json:"id"`
	FileName  string `json:"file_name"`
	FileType  string `json:"file_type"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// KBListResponse 知识库列表响应
type KBListResponse struct {
	KnowledgeBases []KBResponse `json:"knowledge_bases"`
	Total          int64        `json:"total"`
	Page           int          `json:"page"`
	PageSize       int          `json:"page_size"`
}
