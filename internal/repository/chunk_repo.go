package repository

import (
	"DeepSight/internal/model"
	"fmt"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// ChunkRepository 文档块数据访问
type ChunkRepository struct {
	db *gorm.DB
}

// NewChunkRepository 创建文档块数据访问实例
func NewChunkRepository(db *gorm.DB) *ChunkRepository {
	return &ChunkRepository{db: db}
}

// Create 创建文档块
func (r *ChunkRepository) Create(chunk *model.Chunk) error {
	return r.db.Create(chunk).Error
}

// SearchSimilar 根据向量相似度搜索chunk（cosine distance）
func (r *ChunkRepository) SearchSimilar(vector []float32, fileHashes []string, limit int) ([]model.Chunk, error) {
	if len(fileHashes) == 0 {
		return nil, fmt.Errorf("no file hashes provided")
	}

	var chunks []model.Chunk
	vectorStr := pgvector.NewVector(vector).String()

	// 使用 pgvector 的 cosine distance 操作符 <=> 进行相似度搜索
	// <=> 返回 cosine distance，值越小表示越相似
	err := r.db.Raw(`
		SELECT id, file_hash, chunk_index, content, start_offset, end_offset, created_at, updated_at
		FROM chunks
		WHERE file_hash IN ?
		ORDER BY vector <=> ?
		LIMIT ?
	`, fileHashes, vectorStr, limit).Scan(&chunks).Error

	if err != nil {
		return nil, fmt.Errorf("failed to search similar chunks: %w", err)
	}

	return chunks, nil
}

