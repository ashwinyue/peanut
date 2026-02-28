// Package model 定义数据模型
package model

import (
	"time"
)

// BaseModel 基础模型，包含公共字段
type BaseModel struct {
	ID        int64     `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// User 用户模型
type User struct {
	BaseModel
	Username string `json:"username" db:"username"`
	Email    string `json:"email" db:"email"`
	Password string `json:"-" db:"password"` // 密码不返回给前端
	Status   int    `json:"status" db:"status"`
}

// 用户状态常量
const (
	UserStatusInactive = 0 // 未激活
	UserStatusActive   = 1 // 已激活
	UserStatusBanned   = 2 // 已封禁
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=32"`
	Email    string `json:"email" binding:"omitempty,email"`
	Status   *int   `json:"status" binding:"omitempty,oneof=0 1 2"`
}

// UserQueryParams 用户查询参数
type UserQueryParams struct {
	Username string `form:"username"`
	Email    string `form:"email"`
	Status   *int   `form:"status"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// TableName 返回表名
func (User) TableName() string {
	return "users"
}
