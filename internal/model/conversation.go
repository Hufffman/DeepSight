package model

import (
	"time"

	"gorm.io/gorm"
)

// Conversation 会话模型
type Conversation struct {
	gorm.Model
	UserID          uint      `gorm:"not null;index" json:"user_id"`
	KnowledgeBaseID uint      `gorm:"not null;index" json:"knowledge_base_id"`
	Title           string    `gorm:"size:200" json:"title"`
	LastMessageAt   time.Time `json:"last_message_at"`
}

func (Conversation) TableName() string {
	return "conversations"
}