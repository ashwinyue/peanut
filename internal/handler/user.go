// Package handler 提供 HTTP 处理器
package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/solariswu/peanut/internal/model"
	"github.com/solariswu/peanut/internal/pkg/response"
	"github.com/solariswu/peanut/internal/service"
)

// UserHandler 用户处理器
type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Create 创建用户
// @Summary 创建用户
// @Description 创建新用户
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body model.CreateUserRequest true "创建用户请求"
// @Success 200 {object} response.Response{data=model.User}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	user, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			response.BadRequest(c, "用户名已存在")
			return
		}
		if errors.Is(err, service.ErrEmailAlreadyUsed) {
			response.BadRequest(c, "邮箱已被使用")
			return
		}
		response.ServerError(c, "创建用户失败")
		return
	}

	response.Success(c, user)
}

// Get 获取单个用户
// @Summary 获取用户
// @Description 根据ID获取用户信息
// @Tags 用户
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=model.User}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	user, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, "用户不存在")
			return
		}
		response.ServerError(c, "获取用户失败")
		return
	}

	response.Success(c, user)
}

// List 获取用户列表
// @Summary 用户列表
// @Description 获取用户列表（分页）
// @Tags 用户
// @Produce json
// @Param username query string false "用户名"
// @Param email query string false "邮箱"
// @Param status query int false "状态"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=response.PageData}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	var params model.UserQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	users, total, err := h.svc.List(c.Request.Context(), &params)
	if err != nil {
		response.ServerError(c, "获取用户列表失败")
		return
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	response.SuccessPage(c, users, total, page, pageSize)
}

// Update 更新用户
// @Summary 更新用户
// @Description 更新用户信息
// @Tags 用户
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body model.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} response.Response{data=model.User}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	user, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, "用户不存在")
			return
		}
		if errors.Is(err, service.ErrUserAlreadyExists) {
			response.BadRequest(c, "用户名已存在")
			return
		}
		if errors.Is(err, service.ErrEmailAlreadyUsed) {
			response.BadRequest(c, "邮箱已被使用")
			return
		}
		response.ServerError(c, "更新用户失败")
		return
	}

	response.Success(c, user)
}

// Delete 删除用户
// @Summary 删除用户
// @Description 删除用户
// @Tags 用户
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, "用户不存在")
			return
		}
		response.ServerError(c, "删除用户失败")
		return
	}

	response.Success(c, nil)
}

// RegisterRoutes 注册路由
func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
	users := r.Group("/users")
	{
		users.POST("", h.Create)
		users.GET("", h.List)
		users.GET("/:id", h.Get)
		users.PUT("/:id", h.Update)
		users.DELETE("/:id", h.Delete)
	}
}
