package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"DeepSight/internal/database"
	"DeepSight/internal/model"

	"gorm.io/gorm"
)

const (
	// 缓存过期时间
	fileHashesCacheExpire = 10 * time.Minute
	fileByHashCacheExpire = 30 * time.Minute
)

// FileRepository 文件数据访问
type FileRepository struct {
	db *gorm.DB
}

// NewFileRepository 创建文件数据访问实例
func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

// Create 创建文件
func (r *FileRepository) Create(file *model.File) error {
	return r.db.Create(file).Error
}

// GetByID 根据ID获取文件
func (r *FileRepository) GetByID(id uint) (*model.File, error) {
	var file model.File
	err := r.db.First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetByIDWithChunks 根据ID获取文件（包含文档块列表）
func (r *FileRepository) GetByIDWithChunks(id uint) (*model.File, error) {
	var file model.File
	err := r.db.Preload("Chunks", func(db *gorm.DB) *gorm.DB {
		return db.Order("chunk_index")
	}).First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// CountByKBIDs 批量获取知识库的文件数量
func (r *FileRepository) CountByKBIDs(kbIDs []uint) (map[uint]int64, error) {
	if len(kbIDs) == 0 {
		return map[uint]int64{}, nil
	}

	type row struct {
		KnowledgeBaseID uint
		Count           int64
	}
	var rows []row
	err := r.db.Model(&model.File{}).
		Select("knowledge_base_id, count(*) as count").
		Where("knowledge_base_id IN ?", kbIDs).
		Group("knowledge_base_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint]int64, len(rows))
	for _, r := range rows {
		result[r.KnowledgeBaseID] = r.Count
	}
	return result, nil
}

// GetByKnowledgeBaseID 获取知识库的文件列表（分页）
func (r *FileRepository) GetByKnowledgeBaseID(kbID uint, page, pageSize int) ([]model.File, int64, error) {
	var files []model.File
	var total int64

	if err := r.db.Model(&model.File{}).Where("knowledge_base_id = ?", kbID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.Where("knowledge_base_id = ?", kbID).Offset(offset).Limit(pageSize).Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

// GetByHash 根据文件哈希获取文件列表（用于复用判断）
func (r *FileRepository) GetByHash(hash string) ([]model.File, error) {
	ctx := context.Background()
	rdb := database.GetRedis()

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("file:hash:%s", hash)
	if rdb != nil {
		cached, err := rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			var files []model.File
			if err := json.Unmarshal([]byte(cached), &files); err == nil {
				return files, nil
			}
		}
	}

	// 从数据库获取
	var files []model.File
	err := r.db.Where("file_hash = ?", hash).Find(&files).Error
	if err != nil {
		return files, err
	}

	// 写入缓存
	if rdb != nil && len(files) > 0 {
		data, _ := json.Marshal(files)
		rdb.Set(ctx, cacheKey, data, fileByHashCacheExpire)
	}

	return files, nil
}

// Update 更新文件
func (r *FileRepository) Update(file *model.File) error {
	return r.db.Save(file).Error
}

// Delete 删除文件（级联删除chunks）
func (r *FileRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var file model.File
		if err := tx.First(&file, id).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.File{}, id).Error; err != nil {
			return err
		}

		var remaining int64
		if err := tx.Model(&model.File{}).Where("file_hash = ?", file.FileHash).Count(&remaining).Error; err != nil {
			return err
		}
		if remaining == 0 {
			if err := tx.Where("file_hash = ?", file.FileHash).Delete(&model.Chunk{}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetHashesByKnowledgeBaseID 获取知识库所有文件的hash列表
func (r *FileRepository) GetHashesByKnowledgeBaseID(kbID uint) ([]string, error) {
	ctx := context.Background()
	rdb := database.GetRedis()

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("kb:%d:file_hashes", kbID)
	if rdb != nil {
		cached, err := rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			var hashes []string
			if err := json.Unmarshal([]byte(cached), &hashes); err == nil {
				return hashes, nil
			}
		}
	}

	// 从数据库获取
	var hashes []string
	err := r.db.Model(&model.File{}).Where("knowledge_base_id = ?", kbID).Pluck("file_hash", &hashes).Error
	if err != nil {
		return nil, err
	}

	// 写入缓存
	if rdb != nil && len(hashes) > 0 {
		data, _ := json.Marshal(hashes)
		rdb.Set(ctx, cacheKey, data, fileHashesCacheExpire)
	}

	return hashes, nil
}

// InvalidateKBFileHashesCache 清除知识库文件hash缓存
func (r *FileRepository) InvalidateKBFileHashesCache(kbID uint) error {
	ctx := context.Background()
	rdb := database.GetRedis()
	if rdb == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("kb:%d:file_hashes", kbID)
	return rdb.Del(ctx, cacheKey).Err()
}

