// Package service 提供业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/solariswu/peanut/internal/model"
	"github.com/solariswu/peanut/internal/repository"
)

// 用户相关错误
var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserAlreadyExists = errors.New("用户已存在")
	ErrEmailAlreadyUsed  = errors.New("邮箱已被使用")
	ErrInvalidPassword   = errors.New("密码错误")
)

// UserService 用户服务
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Create 创建用户
func (s *UserService) Create(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
	// 检查用户名是否存在
	exists, err := s.repo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if exists {
		return nil, ErrUserAlreadyExists
	}

	// 检查邮箱是否存在
	exists, err = s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if exists {
		return nil, ErrEmailAlreadyUsed
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("加密密码失败: %w", err)
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Status:   model.UserStatusActive,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
}

// GetByID 根据 ID 获取用户
func (s *UserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// List 获取用户列表
func (s *UserService) List(ctx context.Context, params *model.UserQueryParams) ([]model.User, int64, error) {
	return s.repo.List(ctx, params)
}

// Update 更新用户
func (s *UserService) Update(ctx context.Context, id int64, req *model.UpdateUserRequest) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// 更新字段
	if req.Username != "" {
		// 检查用户名是否被其他用户使用
		existingUser, err := s.repo.GetByUsername(ctx, req.Username)
		if err == nil && existingUser.ID != id {
			return nil, ErrUserAlreadyExists
		}
		user.Username = req.Username
	}

	if req.Email != "" {
		// 检查邮箱是否被其他用户使用
		existingUser, err := s.repo.GetByEmail(ctx, req.Email)
		if err == nil && existingUser.ID != id {
			return nil, ErrEmailAlreadyUsed
		}
		user.Email = req.Email
	}

	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	return user, nil
}

// Delete 删除用户
func (s *UserService) Delete(ctx context.Context, id int64) error {
	// 检查用户是否存在
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	return s.repo.Delete(ctx, id)
}
