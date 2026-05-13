package model

import (
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type Chunk struct {
	gorm.Model
	FileHash    string          `gorm:"size:64;not null" json:"file_hash"`
	ChunkIndex  int             `gorm:"not null" json:"chunk_index"`
	Content     string          `gorm:"type:text;not null" json:"content"`
	Vector      pgvector.Vector `gorm:"type:vector(1536)" json:"-"`
	StartOffset int             `json:"start_offset"`
	EndOffset   int             `json:"end_offset"`
	Metadata    string          `gorm:"type:text" json:"metadata"`
}

func (Chunk) TableName() string {
	return "chunks"
}
