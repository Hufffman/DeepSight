package model

import "gorm.io/gorm"

// FileStatus 文件状态常量
const (
	FileStatusPending = "pending" // 等待上传
	FileStatusParsing = "parsing" // 正在解析
	FileStatusParsed  = "parsed"  // 解析完成
	FileStatusError   = "error"   // 处理出错
)

// File 文件模型
type File struct {
	gorm.Model
	KnowledgeBaseID uint          `gorm:"not null;index" json:"knowledge_base_id"` // 所属知识库（1对多关系，知识库是1）
	KnowledgeBase   KnowledgeBase `gorm:"foreignKey:KnowledgeBaseID"`
	FileName        string        `gorm:"size:255;not null" json:"file_name"`
	FileHash        string        `gorm:"size:64;not null" json:"file_hash"` // SHA256哈希，用于复用
	FileSize        int64         `gorm:"not null" json:"file_size"`
	StorageKey      string        `gorm:"size:255" json:"storage_key"`             // rustfs存储对象ID
	StorageURL      string        `gorm:"size:500" json:"storage_url"`             // 访问URL
	FileType        string        `gorm:"size:50" json:"file_type"`                // pdf/docx/txt等
	Status          string        `gorm:"size:20;default:'pending'" json:"status"` // 处理状态
	ParsedText      string        `gorm:"type:text" json:"parsed_text"`            // 解析后的文本内容
	ParseError      string        `gorm:"size:500" json:"parse_error"`             // 解析错误信息
}

func (File) TableName() string {
	return "files"
}
