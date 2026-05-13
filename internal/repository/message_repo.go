package repository

import (
	"DeepSight/internal/model"

	"gorm.io/gorm"
)

// MessageRepository 消息数据访问
type MessageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建消息数据访问实例
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create 创建消息
func (r *MessageRepository) Create(msg *model.Message) error {
	return r.db.Create(msg).Error
}

// GetRecentByConversationID 获取会话最近的消息（用于构建对话历史）
func (r *MessageRepository) GetRecentByConversationID(convID uint, limit int) ([]model.Message, error) {
	var messages []model.Message
	err := r.db.Where("conversation_id = ?", convID).Order("created_at desc").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	// 反转顺序，使最新的消息在最后
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

