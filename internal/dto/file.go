package dto

// FileResponse 文件响应
type FileResponse struct {
	ID              uint   `json:"id"`
	KnowledgeBaseID uint   `json:"knowledge_base_id"`
	FileName        string `json:"file_name"`
	FileHash        string `json:"file_hash"`
	FileSize        int64  `json:"file_size"`
	FileType        string `json:"file_type"`
	Status          string `json:"status"`
	ParsedText      string `json:"parsed_text,omitempty"`
	StorageKey      string `json:"storage_key"`
	StorageURL      string `json:"storage_url"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// FileDetailResponse 文件详情响应（包含chunks）
type FileDetailResponse struct {
	ID              uint   `json:"id"`
	KnowledgeBaseID uint   `json:"knowledge_base_id"`
	FileName        string `json:"file_name"`
	FileHash        string `json:"file_hash"`
	FileSize        int64  `json:"file_size"`
	FileType        string `json:"file_type"`
	Status          string `json:"status"`
	ParsedText      string `json:"parsed_text,omitempty"`
	StorageKey      string `json:"storage_key"`
	StorageURL      string `json:"storage_url"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// ChunkInfo 文档块信息
type ChunkInfo struct {
	ID          uint   `json:"id"`
	ChunkIndex  int    `json:"chunk_index"`
	Content     string `json:"content"`
	StartOffset int    `json:"start_offset"`
	EndOffset   int    `json:"end_offset"`
}

// FileListResponse 文件列表响应
type FileListResponse struct {
	Files    []FileResponse `json:"files"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}
