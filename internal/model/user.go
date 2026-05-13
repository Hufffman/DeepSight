package model

import "gorm.io/gorm"

// User 用户模型
type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password string `gorm:"size:255;not null" json:"-"` // 不暴露给 JSON
	Email    string `gorm:"uniqueIndex;size:100;not null" json:"email"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
