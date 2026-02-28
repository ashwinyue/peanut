// Package database 提供 PostgreSQL 数据库连接管理
package database

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/solariswu/peanut/internal/config"
)

// PostgreSQL 数据库连接
type PostgreSQL struct {
	db *gorm.DB
}

// NewPostgreSQL 创建新的 GORM 数据库连接
func NewPostgreSQL(cfg *config.DatabaseConfig) (*PostgreSQL, error) {
	// 打开数据库连接，使用默认 logger
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层 SQL DB 以配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(int(cfg.MaxConns))
	sqlDB.SetMaxIdleConns(int(cfg.MinConns))
	sqlDB.SetConnMaxLifetime(cfg.MaxConnLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.MaxConnIdleTime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	return &PostgreSQL{db: db}, nil
}

// DB 获取 GORM DB 实例
func (p *PostgreSQL) DB() *gorm.DB {
	return p.db
}

// Close 关闭数据库连接
func (p *PostgreSQL) Close() {
	sqlDB, err := p.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// Ping 测试连接
func (p *PostgreSQL) Ping(ctx context.Context) error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
