package model

import "gorm.io/gorm"

// KnowledgeBase 知识库模型
type KnowledgeBase struct {
	gorm.Model
	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"size:500" json:"description"`
	UserID      uint   `gorm:"not null;index" json:"user_id"` // 绑定用户（1对多关系，用户是1）
	User        User   `gorm:"foreignKey:UserID"`
	Files       []File `gorm:"foreignKey:KnowledgeBaseID"` // 包含的文件列表
}

// TableName 指定表名
func (KnowledgeBase) TableName() string {
	return "knowledge_bases"
}