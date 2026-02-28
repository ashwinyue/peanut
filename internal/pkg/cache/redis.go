// Package cache 提供 Redis 缓存管理
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/solariswu/peanut/internal/config"
)

// Redis 缓存客户端
type Redis struct {
	client *redis.Client
}

// NewRedis 创建新的 Redis 客户端
func NewRedis(cfg *config.RedisConfig) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis 连接测试失败: %w", err)
	}

	return &Redis{client: client}, nil
}

// Client 获取原始 Redis 客户端
func (r *Redis) Client() *redis.Client {
	return r.client
}

// Close 关闭连接
func (r *Redis) Close() error {
	return r.client.Close()
}

// Set 设置缓存
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化值失败: %w", err)
	}
	return r.client.Set(ctx, key, data, expiration).Err()
}

// Get 获取缓存
func (r *Redis) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Delete 删除缓存
func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// SetNX 仅当键不存在时设置
func (r *Redis) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("序列化值失败: %w", err)
	}
	return r.client.SetNX(ctx, key, data, expiration).Result()
}
