package handler

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/solariswu/peanut/internal/agent/geo"
	"github.com/solariswu/peanut/internal/agent/geo/models"
	"github.com/solariswu/peanut/internal/pkg/response"
)

// GEOHandler GEO 处理器
type GEOHandler struct {
	geoService *geo.Service
}

// NewGEOHandler 创建新的 GEO 处理器
func NewGEOHandler(geoService *geo.Service) *GEOHandler {
	return &GEOHandler{
		geoService: geoService,
	}
}

// AnalyzeContent 分析内容
// @Summary 分析网页内容的 GEO 优化建议
// @Description 分析指定 URL 的网页内容，生成 GEO 优化报告
// @Tags GEO
// @Accept json
// @Produce json
// @Param request body models.AnalysisRequest true "分析请求"
// @Success 200 {object} response.Response{data=models.OptimizationReport}
// @Router /api/v1/geo/analyze [post]
func (h *GEOHandler) AnalyzeContent(c *gin.Context) {
	var req models.AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 执行分析
	report, err := h.geoService.AnalyzeURL(c.Request.Context(), req.URL)
	if err != nil {
		response.ServerError(c, "分析失败: "+err.Error())
		return
	}

	response.Success(c, report)
}

// AnalyzeContentWithProgress 分析内容（带进度）
// @Summary 分析网页内容（带进度）
// @Description 分析指定 URL 的网页内容，返回 SSE 进度流
// @Tags GEO
// @Accept json
// @Produce text/event-stream
// @Param request body models.StreamAnalysisRequest true "分析请求"
// @Router /api/v1/geo/analyze/stream [post]
func (h *GEOHandler) AnalyzeContentWithProgress(c *gin.Context) {
	var req models.StreamAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// 发送开始事件
	c.SSEvent("start", `{"status":"started"}`)
	c.Writer.Flush()

	// 执行分析（带进度）
	report, err := h.geoService.AnalyzeURLWithProgress(c.Request.Context(), req.URL,
		func(step int, total int, agentName string, message string) {
			// 发送进度事件
			event := fmt.Sprintf(`{"step":%d,"total":%d,"agent":"%s","message":"%s"}`,
				step, total, agentName, escapeJSON(message))
			c.SSEvent("progress", event)
			c.Writer.Flush()
		})

	if err != nil {
		c.SSEvent("error", fmt.Sprintf(`{"error":"%s"}`, escapeJSON(err.Error())))
		return
	}

	// 发送完成事件
	c.SSEvent("complete", fmt.Sprintf(`{"status":"success","score":%d}`, report.OverallScore))
}

// escapeJSON 转义 JSON 字符串
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// RegisterRoutes 注册路由
func (h *GEOHandler) RegisterRoutes(router *gin.RouterGroup) {
	geo := router.Group("/geo")
	{
		geo.POST("/analyze", h.AnalyzeContent)
		geo.POST("/analyze/stream", h.AnalyzeContentWithProgress)
	}
}
