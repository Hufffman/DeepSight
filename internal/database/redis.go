package database

import (
	"context"
	"fmt"
	"time"

	"DeepSight/internal/config"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// InitializeRedis 初始化 Redis 连接
func InitializeRedis(cfg *config.RedisConfig) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	return nil
}

// GetRedis 获取 Redis 客户端实例
func GetRedis() *redis.Client {
	return RedisClient
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

// SetToken 将 token 存入 Redis（用于白名单验证）
func SetToken(ctx context.Context, userID uint, token string, expire time.Duration) error {
	key := fmt.Sprintf("token:%d:%s", userID, token)
	return RedisClient.Set(ctx, key, "1", expire).Err()
}

// DeleteToken 从 Redis 删除 token（登出时使用）
func DeleteToken(ctx context.Context, userID uint, token string) error {
	key := fmt.Sprintf("token:%d:%s", userID, token)
	return RedisClient.Del(ctx, key).Err()
}

// TokenExists 检查 token 是否存在（是否有效）
func TokenExists(ctx context.Context, userID uint, token string) (bool, error) {
	key := fmt.Sprintf("token:%d:%s", userID, token)
	result, err := RedisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}
