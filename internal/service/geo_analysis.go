package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/solariswu/peanut/internal/agent/geo"
	"github.com/solariswu/peanut/internal/model"
	"github.com/solariswu/peanut/internal/pkg/progress"
	"github.com/solariswu/peanut/internal/repository"
)

// GEOAnalysisService GEO 分析服务
type GEOAnalysisService struct {
	repo        *repository.GEOAnalysisRepository
	agent       geo.AgentService
	progressMgr *progress.Manager
	totalSteps  int
}

// NewGEOAnalysisService 创建服务
func NewGEOAnalysisService(repo *repository.GEOAnalysisRepository, agent geo.AgentService, progressMgr *progress.Manager) *GEOAnalysisService {
	return &GEOAnalysisService{
		repo:        repo,
		agent:       agent,
		progressMgr: progressMgr,
		totalSteps:  8, // GEO 分析的总步骤数（7步+验证）
	}
}

// Create 创建分析任务
func (s *GEOAnalysisService) Create(ctx context.Context, req *model.GEOAnalysisCreateRequest, userID *int64) (*model.GEOAnalysis, error) {
	// 检查是否已有正在运行的分析
	if existing, _ := s.repo.GetByURL(req.URL); existing != nil && (existing.Status == "pending" || existing.Status == "processing") {
		return nil, fmt.Errorf("该 URL 正在分析中")
	}

	// 验证并设置默认平台
	platform := req.Platform
	if platform == "" {
		platform = "doubao"
	}

	analysis := &model.GEOAnalysis{
		URL:      req.URL,
		Platform: platform,
		Status:   "pending",
		UserID:   userID,
	}

	if err := s.repo.Create(analysis); err != nil {
		return nil, err
	}

	// 异步执行分析
	go s.executeAnalysis(context.Background(), analysis.ID, req.URL, platform)

	return analysis, nil
}

// executeAnalysis 执行分析
func (s *GEOAnalysisService) executeAnalysis(ctx context.Context, analysisID int64, url string, platform string) {
	// 更新状态为处理中
	s.repo.UpdateFields(analysisID, map[string]interface{}{
		"status": "processing",
	})

	// 发布初始进度
	if s.progressMgr != nil {
		s.progressMgr.Update(analysisID, 0, s.totalSteps, "初始化", fmt.Sprintf("开始 %s GEO 分析", platform))
	}

	// 执行 GEO 分析（传入平台参数）
	report, err := s.agent.ExecuteWithStreaming(ctx, url, platform, func(step int, agentName string, message string) {
		// 进度回调 - 通过进度管理器广播
		if s.progressMgr != nil {
			s.progressMgr.Update(analysisID, step, s.totalSteps, agentName, message)
		}
	})

	if err != nil {
		// 标记失败
		if s.progressMgr != nil {
			s.progressMgr.Fail(analysisID, err.Error())
		}
		s.repo.MarkFailed(analysisID, err.Error())
		return
	}

	// 更新最终结果
	now := time.Now()
	updates := map[string]interface{}{
		"title":         report.Title,
		"main_query":    report.MainQuery,
		"overall_score": report.OverallScore,
		"status":        "completed",
		"completed_at":  &now,
	}

	// 保存中间结果字段
	if report.QueryFanout != "" {
		updates["query_fanout"] = report.QueryFanout
	}
	if report.AIOverview != "" {
		updates["ai_overview"] = report.AIOverview
	}
	if report.QueryFanoutSummary != "" {
		updates["query_fanout_summary"] = report.QueryFanoutSummary
	}
	if report.OptimizationReport != "" {
		updates["optimization_report"] = report.OptimizationReport
	}
	if report.OptimizedArticle != "" {
		updates["optimized_article"] = report.OptimizedArticle
	}

	// 序列化复杂字段
	if len(report.ContentGaps) > 0 {
		gapsJSON, _ := json.Marshal(report.ContentGaps)
		updates["content_gaps"] = string(gapsJSON)
	}

	if len(report.OptimizationSuggestions) > 0 {
		suggestionsJSON, _ := json.Marshal(report.OptimizationSuggestions)
		updates["optimization_suggestions"] = string(suggestionsJSON)
	}

	// 注意：验证结果在第8步生成，需要从 stepOutputs 中解析
	// 这里暂时跳过，后续可以扩展报告模型来包含验证结果

	s.repo.UpdateFields(analysisID, updates)

	// 标记完成
	if s.progressMgr != nil {
		s.progressMgr.Complete(analysisID, report.OverallScore)
	}
}

// GetByID 获取分析详情
func (s *GEOAnalysisService) GetByID(id int64) (*model.GEOAnalysisResponse, error) {
	analysis, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.ToResponse(analysis), nil
}

// List 查询列表
func (s *GEOAnalysisService) List(req *model.GEOAnalysisListRequest) ([]model.GEOAnalysisResponse, int64, error) {
	analyses, total, err := s.repo.List(req)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.GEOAnalysisResponse, len(analyses))
	for i, analysis := range analyses {
		responses[i] = *s.ToResponse(&analysis)
	}

	return responses, total, nil
}

// Delete 删除分析
func (s *GEOAnalysisService) Delete(id int64) error {
	return s.repo.Delete(id)
}

// ToResponse 转换为响应格式（公开方法）
func (s *GEOAnalysisService) ToResponse(analysis *model.GEOAnalysis) *model.GEOAnalysisResponse {
	return &model.GEOAnalysisResponse{
		ID:                      analysis.ID,
		URL:                     analysis.URL,
		Title:                   analysis.Title,
		MainQuery:               analysis.MainQuery,
		Platform:                analysis.Platform,
		OverallScore:            analysis.OverallScore,
		OptimizedScore:          analysis.OptimizedScore,
		Status:                  analysis.Status,
		ErrorMessage:            analysis.ErrorMessage,
		QueryFanout:             analysis.QueryFanout,
		AIOverview:              analysis.AIOverview,
		QueryFanoutSummary:      analysis.QueryFanoutSummary,
		OptimizationReport:      analysis.OptimizationReport,
		OptimizedArticle:        analysis.OptimizedArticle,
		ContentGaps:             analysis.ContentGaps,
		OptimizationSuggestions: analysis.OptimizationSuggestions,
		ValidationResult:        analysis.ValidationResult,
		CreatedAt:               analysis.CreatedAt,
		UpdatedAt:               analysis.UpdatedAt,
		CompletedAt:             analysis.CompletedAt,
	}
}
