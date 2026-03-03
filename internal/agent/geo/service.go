/*
 * Copyright 2025 Peanut Authors
 *
 * GEO Service - Flow 模式实现
 */

package geo

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/compose"

	"github.com/solariswu/peanut/internal/agent/geo/flow"
	"github.com/solariswu/peanut/internal/agent/geo/models"
)

// AgentService Agent 服务接口
type AgentService interface {
	ExecuteWithStreaming(ctx context.Context, url string, platform string, callback func(step int, agentName string, message string)) (*models.OptimizationReport, error)
}

// Service GEO 服务
type Service struct {
	runnable compose.Runnable[string, string]
}

// NewService 创建新的 GEO 服务（使用 flow 模式）
func NewService(platform string) (*Service, error) {
	ctx := context.Background()

	runnable, err := flow.BuildGraph(ctx)
	if err != nil {
		return nil, fmt.Errorf("构建 Flow Graph 失败: %w", err)
	}

	return &Service{
		runnable: runnable,
	}, nil
}

// NewDefaultService 创建默认 GEO 服务
func NewDefaultService() (*Service, error) {
	return NewService("google")
}

// Analyze 分析 URL（实现 flow.AgentService 接口）
func (s *Service) Analyze(ctx context.Context, url, platform string) (*models.OptimizationReport, error) {
	// 执行 Graph，使用 StateModifier 初始化状态
	_, err := s.runnable.Invoke(ctx, url,
		compose.WithStateModifier(func(ctx context.Context, path compose.NodePath, state any) error {
			s := state.(*flow.State)
			s.URL = url
			s.PlatformType = platform
			s.Goto = flow.AgentTitleScraper
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("执行 GEO 分析失败: %w", err)
	}

	// 获取最终状态
	var finalState *flow.State
	err = compose.ProcessState[*flow.State](ctx, func(_ context.Context, state *flow.State) error {
		finalState = state
		return nil
	})
	if err != nil {
		return nil, err
	}

	return convertStateToReport(finalState), nil
}

// AnalyzeWithProgress 分析 URL（带进度回调，实现 flow.AgentService 接口）
func (s *Service) AnalyzeWithProgress(ctx context.Context, url, platform string, progress func(step int, total int, agentName string, message string)) (*models.OptimizationReport, error) {
	// Flow 模式暂不支持细粒度进度回调
	return s.Analyze(ctx, url, platform)
}

// AnalyzeURL 分析 URL（完整流程，使用指定平台）
// Deprecated: 使用 Analyze 代替
func (s *Service) AnalyzeURL(ctx context.Context, url string, platform string) (*models.OptimizationReport, error) {
	return s.Analyze(ctx, url, platform)
}

// AnalyzeURLWithProgress 分析 URL（带进度回调，使用指定平台）
// Deprecated: 使用 AnalyzeWithProgress 代替
func (s *Service) AnalyzeURLWithProgress(ctx context.Context, url string, platform string, progress func(step int, total int, agentName string, message string)) (*models.OptimizationReport, error) {
	return s.AnalyzeWithProgress(ctx, url, platform, progress)
}

// ExecuteWithStreaming 实现 AgentService 接口
func (s *Service) ExecuteWithStreaming(ctx context.Context, url string, platform string, callback func(step int, agentName string, message string)) (*models.OptimizationReport, error) {
	return s.AnalyzeWithProgress(ctx, url, platform, func(step int, total int, agentName string, message string) {
		if callback != nil {
			callback(step, agentName, message)
		}
	})
}

// convertStateToReport 将 State 转换为 OptimizationReport
func convertStateToReport(state *flow.State) *models.OptimizationReport {
	if state == nil {
		return &models.OptimizationReport{
			OverallScore: 0,
		}
	}

	report := &models.OptimizationReport{
		URL:              state.URL,
		Title:            state.Title,
		MainQuery:        state.MainQuery,
		QueryFanout:      strings.Join(state.QueryFanout, ", "),
		QueryFanoutSummary: state.QuerySummary,
		AIOverview:       state.AIOverview,
		OptimizedArticle: state.OptimizedArticle,
	}

	if state.Report != nil {
		report.OverallScore = state.Report.OverallScore
		report.OptimizationSuggestions = state.Report.OptimizationSuggestions
		report.OptimizationReport = state.Report.OptimizationReport
	}

	return report
}
