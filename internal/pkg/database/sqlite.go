// Package database 提供 SQLite 数据库连接管理
package database

import (
	"context"
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SQLite 数据库连接
type SQLite struct {
	db *gorm.DB
}

// NewSQLite 创建新的 GORM SQLite 连接
func NewSQLite(dsn string) (*SQLite, error) {
	// 打开数据库连接
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层 SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 设置连接池参数（SQLite 单连接）
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	return &SQLite{db: db}, nil
}

// DB 获取 GORM DB 实例
func (s *SQLite) DB() *gorm.DB {
	return s.db
}

// Close 关闭数据库连接
func (s *SQLite) Close() {
	sqlDB, err := s.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// Ping 测试连接
func (s *SQLite) Ping(ctx context.Context) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
