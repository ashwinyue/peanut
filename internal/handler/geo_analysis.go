package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/solariswu/peanut/internal/agent/geo/models"
	"github.com/solariswu/peanut/internal/model"
	"github.com/solariswu/peanut/internal/pkg/progress"
	"github.com/solariswu/peanut/internal/pkg/response"
	"github.com/solariswu/peanut/internal/service"
)

// GEOAnalysisHandler GEO 分析处理器
type GEOAnalysisHandler struct {
	service     *service.GEOAnalysisService
	progressMgr *progress.Manager
}

// NewGEOAnalysisHandler 创建处理器
func NewGEOAnalysisHandler(service *service.GEOAnalysisService, progressMgr *progress.Manager) *GEOAnalysisHandler {
	return &GEOAnalysisHandler{
		service:     service,
		progressMgr: progressMgr,
	}
}

// RegisterRoutes 注册路由
func (h *GEOAnalysisHandler) RegisterRoutes(r *gin.RouterGroup) {
	analysis := r.Group("/geo/analysis")
	{
		analysis.POST("", h.Create)
		analysis.GET("", h.List)
		analysis.GET("/platforms", h.GetPlatforms) // 获取支持的平台列表
		analysis.GET("/:id", h.GetByID)
		analysis.DELETE("/:id", h.Delete)
		analysis.GET("/:id/progress", h.GetProgress)
	}
}

// Create 创建分析任务
// @Summary 创建 GEO 分析任务
// @Description 创建一个新的 URL 分析任务
// @Tags GEO 分析
// @Accept json
// @Produce json
// @Param request body model.GEOAnalysisCreateRequest true "创建请求"
// @Success 200 {object} response.Response{data=model.GEOAnalysisResponse}
// @Router /api/v1/geo/analysis [post]
func (h *GEOAnalysisHandler) Create(c *gin.Context) {
	var req model.GEOAnalysisCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// TODO: 从 JWT 获取 userID
	var userID *int64

	analysis, err := h.service.Create(c.Request.Context(), &req, userID)
	if err != nil {
		response.ServerError(c, "创建分析任务失败: "+err.Error())
		return
	}

	response.Success(c, h.service.ToResponse(analysis))
}

// GetByID 获取分析详情
// @Summary 获取 GEO 分析详情
// @Description 根据 ID 获取分析详情
// @Tags GEO 分析
// @Accept json
// @Produce json
// @Param id path int true "分析 ID"
// @Success 200 {object} response.Response{data=model.GEOAnalysisResponse}
// @Router /api/v1/geo/analysis/{id} [get]
func (h *GEOAnalysisHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}

	analysis, err := h.service.GetByID(id)
	if err != nil {
		response.NotFound(c, "分析记录不存在")
		return
	}

	response.Success(c, analysis)
}

// List 查询分析列表
// @Summary 获取 GEO 分析列表
// @Description 分页查询分析记录
// @Tags GEO 分析
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param status query string false "状态筛选"
// @Success 200 {object} response.PageResponse
// @Router /api/v1/geo/analysis [get]
func (h *GEOAnalysisHandler) List(c *gin.Context) {
	var req model.GEOAnalysisListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// TODO: 从 JWT 获取 userID
	// if req.UserID == nil {
	//     req.UserID = &userID
	// }

	list, total, err := h.service.List(&req)
	if err != nil {
		response.ServerError(c, "查询失败: "+err.Error())
		return
	}

	response.SuccessPage(c, list, total, req.Page, req.PageSize)
}

// Delete 删除分析
// @Summary 删除 GEO 分析
// @Description 删除指定的分析记录
// @Tags GEO 分析
// @Accept json
// @Produce json
// @Param id path int true "分析 ID"
// @Success 200 {object} response.Response
// @Router /api/v1/geo/analysis/{id} [delete]
func (h *GEOAnalysisHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		response.ServerError(c, "删除失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// GetProgress 获取分析进度（SSE）
// @Summary 获取分析进度
// @Description 通过 Server-Sent Events 获取实时进度
// @Tags GEO 分析
// @Produce text/event-stream
// @Param id path int true "分析 ID"
// @Router /api/v1/geo/analysis/{id}/progress [get]
func (h *GEOAnalysisHandler) GetProgress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// 立即刷新头部
	c.Writer.Flush()

	// 检查进度管理器
	if h.progressMgr == nil {
		h.sendSSEError(c, "进度服务不可用")
		return
	}

	// 订阅进度更新
	progressCh := h.progressMgr.Subscribe(id)
	defer h.progressMgr.Unsubscribe(id, progressCh)

	// 获取请求上下文
	ctx := c.Request.Context()

	// 发送初始连接事件
	h.sendSSEEvent(c, "connected", map[string]int64{"analysis_id": id})

	// 监听进度更新
	for {
		select {
		case <-ctx.Done():
			// 客户端断开连接
			return
		case p, ok := <-progressCh:
			if !ok {
				// channel 已关闭，分析完成或失败
				return
			}
			// 发送进度事件
			h.sendSSEEvent(c, "progress", p)

			// 如果状态是终态（完成或失败），结束流
			if p.Status == "completed" || p.Status == "failed" {
				return
			}
		}
	}
}

// sendSSEEvent 发送 SSE 事件
func (h *GEOAnalysisHandler) sendSSEEvent(c *gin.Context, event string, data interface{}) {
	c.SSEvent(event, data)
	c.Writer.Flush()
}

// sendSSEError 发送 SSE 错误事件
func (h *GEOAnalysisHandler) sendSSEError(c *gin.Context, message string) {
	c.SSEvent("error", map[string]string{"message": message})
	c.Writer.Flush()
}

// GetPlatforms 获取支持的平台列表
// @Summary 获取支持的 GEO 优化平台
// @Description 获取所有支持的 AI 搜索平台列表及其权重配置
// @Tags GEO 分析
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]models.PlatformConfig}
// @Router /api/v1/geo/analysis/platforms [get]
func (h *GEOAnalysisHandler) GetPlatforms(c *gin.Context) {
	platforms := models.AllPlatforms()
	response.Success(c, platforms)
}
