package database

import (
	"fmt"

	"DeepSight/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Initialize 初始化数据库连接
func Initialize(cfg *config.DatabaseConfig) error {
	dsn := cfg.DSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 启用 pgvector 扩展
	if err := enablePGVector(db); err != nil {
		return fmt.Errorf("failed to enable pgvector: %w", err)
	}

	DB = db
	return nil
}

// enablePGVector 启用 pgvector 扩展
func enablePGVector(db *gorm.DB) error {
	return db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// Close 关闭数据库连接
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
