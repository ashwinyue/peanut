// Package database 提供 PostgreSQL 数据库连接管理
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/solariswu/peanut/internal/config"
)

// PostgreSQL 数据库连接池
type PostgreSQL struct {
	pool *pgxpool.Pool
}

// NewPostgreSQL 创建新的 PostgreSQL 连接池
func NewPostgreSQL(cfg *config.DatabaseConfig) (*PostgreSQL, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("解析数据库配置失败: %w", err)
	}

	// 设置连接池参数
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	// 创建连接池
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接池失败: %w", err)
	}

	// 测试连接
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	return &PostgreSQL{pool: pool}, nil
}

// Pool 获取连接池
func (p *PostgreSQL) Pool() *pgxpool.Pool {
	return p.pool
}

// Close 关闭连接池
func (p *PostgreSQL) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}

// Ping 测试连接
func (p *PostgreSQL) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}
