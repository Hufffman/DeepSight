package model

import "gorm.io/gorm"

// Message 消息模型
type Message struct {
	gorm.Model
	ConversationID  uint   `gorm:"not null;index" json:"conversation_id"`
	Role            string `gorm:"size:20;not null" json:"role"` // "user" 或 "assistant"
	Content         string `gorm:"type:text;not null" json:"content"`
	RetrievedChunks string `gorm:"type:text" json:"retrieved_chunks"` // JSON存储检索到的chunk IDs和内容摘要
}

func (Message) TableName() string {
	return "messages"
}