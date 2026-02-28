// Package repository 提供数据访问层
package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/solariswu/peanut/internal/model"
)

// UserRepository 用户仓储
type UserRepository struct {
	db Querier
}

// Querier 数据库查询接口（便于测试时 mock）
type Querier interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db Querier) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	sql := `
		INSERT INTO users (username, email, password, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, sql,
		user.Username, user.Email, user.Password, user.Status,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	sql := `
		SELECT id, username, email, password, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var user model.User
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	sql := `
		SELECT id, username, email, password, status, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	var user model.User
	err := r.db.QueryRow(ctx, sql, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	sql := `
		SELECT id, username, email, password, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var user model.User
	err := r.db.QueryRow(ctx, sql, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	return &user, nil
}

// List 获取用户列表
func (r *UserRepository) List(ctx context.Context, params *model.UserQueryParams) ([]model.User, int64, error) {
	// 构建基础查询
	baseSQL := `FROM users WHERE 1=1`
	args := make([]interface{}, 0)
	argIndex := 1

	if params.Username != "" {
		baseSQL += fmt.Sprintf(" AND username LIKE $%d", argIndex)
		args = append(args, "%"+params.Username+"%")
		argIndex++
	}
	if params.Email != "" {
		baseSQL += fmt.Sprintf(" AND email LIKE $%d", argIndex)
		args = append(args, "%"+params.Email+"%")
		argIndex++
	}
	if params.Status != nil {
		baseSQL += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *params.Status)
		argIndex++
	}

	// 获取总数
	var total int64
	countSQL := "SELECT COUNT(*) " + baseSQL
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("统计用户数量失败: %w", err)
	}

	// 分页参数
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 获取列表
	listSQL := fmt.Sprintf(`
		SELECT id, username, email, password, status, created_at, updated_at
		%s
		ORDER BY id DESC
		LIMIT $%d OFFSET $%d
	`, baseSQL, argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("获取用户列表失败: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Password,
			&user.Status, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("扫描用户数据失败: %w", err)
		}
		users = append(users, user)
	}

	return users, total, rows.Err()
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	sql := `
		UPDATE users
		SET username = $1, email = $2, status = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, sql,
		user.Username, user.Email, user.Status, user.ID,
	).Scan(&user.UpdatedAt)
}

// Delete 删除用户
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	sql := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

// ExistsByUsername 检查用户名是否存在
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, sql, username).Scan(&exists)
	return exists, err
}

// ExistsByEmail 检查邮箱是否存在
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, sql, email).Scan(&exists)
	return exists, err
}
