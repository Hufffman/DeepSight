package repository

import (
	"DeepSight/internal/model"

	"gorm.io/gorm"
)

// KnowledgeBaseRepository 知识库数据访问
type KnowledgeBaseRepository struct {
	db *gorm.DB
}

// NewKnowledgeBaseRepository 创建知识库数据访问实例
func NewKnowledgeBaseRepository(db *gorm.DB) *KnowledgeBaseRepository {
	return &KnowledgeBaseRepository{db: db}
}

// Create 创建知识库
func (r *KnowledgeBaseRepository) Create(kb *model.KnowledgeBase) error {
	return r.db.Create(kb).Error
}

// GetByID 根据ID获取知识库
func (r *KnowledgeBaseRepository) GetByID(id uint) (*model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	err := r.db.First(&kb, id).Error
	if err != nil {
		return nil, err
	}
	return &kb, nil
}

// GetByIDWithFiles 根据ID获取知识库（包含文件列表）
func (r *KnowledgeBaseRepository) GetByIDWithFiles(id uint) (*model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	err := r.db.Preload("Files").First(&kb, id).Error
	if err != nil {
		return nil, err
	}
	return &kb, nil
}

// GetByUserID 获取用户的知识库列表（分页）
func (r *KnowledgeBaseRepository) GetByUserID(userID uint, page, pageSize int) ([]model.KnowledgeBase, int64, error) {
	var kbs []model.KnowledgeBase
	var total int64

	if err := r.db.Model(&model.KnowledgeBase{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ?", userID).Offset(offset).Limit(pageSize).Find(&kbs).Error; err != nil {
		return nil, 0, err
	}

	return kbs, total, nil
}

// Update 更新知识库
func (r *KnowledgeBaseRepository) Update(kb *model.KnowledgeBase) error {
	return r.db.Save(kb).Error
}

// Delete 删除知识库（级联删除文件）
func (r *KnowledgeBaseRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("knowledge_base_id = ?", id).Delete(&model.File{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&model.KnowledgeBase{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}