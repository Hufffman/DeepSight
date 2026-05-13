package repository

import (
	"DeepSight/internal/model"

	"gorm.io/gorm"
)

// ConversationRepository 会话数据访问
type ConversationRepository struct {
	db *gorm.DB
}

// NewConversationRepository 创建会话数据访问实例
func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// Create 创建会话
func (r *ConversationRepository) Create(conv *model.Conversation) error {
	return r.db.Create(conv).Error
}

// GetByID 根据ID获取会话
func (r *ConversationRepository) GetByID(id uint) (*model.Conversation, error) {
	var conv model.Conversation
	err := r.db.First(&conv, id).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// GetByIDWithMessages 根据ID获取会话（包含消息列表）
func (r *ConversationRepository) GetByIDWithMessages(id uint, page, pageSize int) (*model.Conversation, []model.Message, int64, error) {
	var conv model.Conversation
	err := r.db.First(&conv, id).Error
	if err != nil {
		return nil, nil, 0, err
	}

	var messages []model.Message
	var total int64

	if err := r.db.Model(&model.Message{}).Where("conversation_id = ?", id).Count(&total).Error; err != nil {
		return nil, nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.Where("conversation_id = ?", id).Order("created_at desc").Offset(offset).Limit(pageSize).Find(&messages).Error; err != nil {
		return nil, nil, 0, err
	}

	return &conv, messages, total, nil
}

// GetByUserID 获取用户的会话列表（分页）
func (r *ConversationRepository) GetByUserID(userID uint, page, pageSize int) ([]model.Conversation, int64, error) {
	var convs []model.Conversation
	var total int64

	if err := r.db.Model(&model.Conversation{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ?", userID).Order("last_message_at desc").Offset(offset).Limit(pageSize).Find(&convs).Error; err != nil {
		return nil, 0, err
	}

	return convs, total, nil
}

// Update 更新会话
func (r *ConversationRepository) Update(conv *model.Conversation) error {
	return r.db.Save(conv).Error
}

// Delete 删除会话（级联删除消息）
func (r *ConversationRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("conversation_id = ?", id).Delete(&model.Message{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&model.Conversation{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}