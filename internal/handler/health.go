// Package handler 提供健康检查处理器
package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/solariswu/peanut/internal/pkg/response"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health 健康检查
// @Summary 健康检查
// @Description 检查服务是否正常运行
// @Tags 系统
// @Produce json
// @Success 200 {object} response.Response
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	response.SuccessWithMessage(c, "ok", gin.H{
		"status": "healthy",
	})
}

// Ready 就绪检查
// @Summary 就绪检查
// @Description 检查服务是否就绪
// @Tags 系统
// @Produce json
// @Success 200 {object} response.Response
// @Router /ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	// 可以在这里添加数据库、Redis 等的检查
	response.SuccessWithMessage(c, "ready", gin.H{
		"status": "ready",
	})
}

// RegisterRoutes 注册路由
func (h *HealthHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", h.Health)
	r.GET("/ready", h.Ready)
}
