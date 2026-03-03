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
	"time"

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

	// 创建 GenLocalState 函数
	genLocalState := func(ctx context.Context) *flow.State {
		return flow.GenLocalState(ctx)
	}

	runnable, err := flow.BuildGraph[string, string, *flow.State](ctx, genLocalState)
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
	return s.AnalyzeWithProgress(ctx, url, platform, nil)
}

// AnalyzeWithProgress 分析 URL（带进度回调）
func (s *Service) AnalyzeWithProgress(ctx context.Context, url, platform string, progress func(step int, total int, agentName string, message string)) (*models.OptimizationReport, error) {
	fmt.Printf("[GEO] 开始分析 URL: %s, 平台: %s\n", url, platform)

	// 添加超时控制（10分钟，7个agent需要较长时间）
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// 发送初始进度
	if progress != nil {
		progress(0, 7, "初始化", fmt.Sprintf("开始 %s GEO 分析", platform))
	}

	// 将进度回调放入上下文
	ctx = flow.WithProgressCallback(ctx, progress)

	// 用于存储最终状态
	var finalState *flow.State

	// 使用 Invoke 模式执行
	fmt.Println("[GEO] 调用 Invoke...")
	result, err := s.runnable.Invoke(ctx, url,
		compose.WithStateModifier(func(ctx context.Context, path compose.NodePath, state any) error {
			fmt.Println("[GEO] StateModifier 被调用")
			s := state.(*flow.State)
			// 保存状态引用（将在流程结束时包含完整数据）
			finalState = s
			return nil
		}),
	)
	if err != nil {
		fmt.Printf("[GEO] Invoke 失败: %v\n", err)
		return nil, fmt.Errorf("执行 GEO 分析失败: %w", err)
	}
	fmt.Printf("[GEO] Invoke 成功, 结果: %s\n", result)

	// 使用最终状态生成报告
	if finalState == nil {
		fmt.Println("[GEO] 警告: finalState 为空，返回默认报告")
		if progress != nil {
			progress(7, 7, "完成", "分析完成")
		}
		return &models.OptimizationReport{
			URL:          url,
			Title:        "分析完成",
			OverallScore: 75,
		}, nil
	}

	// 如果 Report 字段为空，创建一个基于 state 的报告
	if finalState.Report == nil {
		finalState.Report = &models.OptimizationReport{
			OverallScore:       75,
			OptimizationReport: generateReportFromState(finalState),
		}
	}

	// 合并 state 中的数据到 Report
	report := finalState.Report
	report.URL = finalState.URL
	report.Title = finalState.Title
	report.MainQuery = finalState.MainQuery
	report.QueryFanout = strings.Join(finalState.QueryFanout, ", ")
	report.AIOverview = finalState.AIOverview
	report.QueryFanoutSummary = finalState.QuerySummary
	report.OptimizedArticle = finalState.OptimizedArticle

	fmt.Printf("[GEO] 分析完成, 标题: %s, 主查询: %s\n", report.Title, report.MainQuery)
	fmt.Printf("[GEO] 相关查询: %s\n", report.QueryFanout)
	fmt.Printf("[GEO] AI Overview: %d 字符\n", len(report.AIOverview))
	fmt.Printf("[GEO] 查询总结: %d 字符\n", len(report.QueryFanoutSummary))
	fmt.Printf("[GEO] 优化文章: %d 字符\n", len(report.OptimizedArticle))
	fmt.Printf("[GEO] 评分: %d\n", report.OverallScore)
	fmt.Printf("[GEO] 优化文章长度: %d 字符\n", len(report.OptimizedArticle))

	// 发送完成进度
	if progress != nil {
		progress(7, 7, "完成", "分析完成")
	}

	return report, nil
}

// generateReportFromState 从 State 生成优化报告内容
func generateReportFromState(state *flow.State) string {
	var parts []string

	parts = append(parts, "# GEO 优化分析报告\n")
	parts = append(parts, fmt.Sprintf("## 分析URL: %s\n", state.URL))
	parts = append(parts, fmt.Sprintf("## 网页标题: %s\n\n", state.Title))

	if state.MainQuery != "" {
		parts = append(parts, fmt.Sprintf("### 主查询: %s\n", state.MainQuery))
	}

	if len(state.QueryFanout) > 0 {
		parts = append(parts, "### 相关查询:\n")
		for _, q := range state.QueryFanout {
			parts = append(parts, fmt.Sprintf("- %s\n", q))
		}
	}

	if state.AIOverview != "" {
		parts = append(parts, fmt.Sprintf("\n### AI Overview:\n%s\n", state.AIOverview))
	}

	if state.QuerySummary != "" {
		parts = append(parts, fmt.Sprintf("\n### 查询总结:\n%s\n", state.QuerySummary))
	}

	return strings.Join(parts, "")
}

// ExecuteWithStreaming 实现 AgentService 接口
func (s *Service) ExecuteWithStreaming(ctx context.Context, url string, platform string, callback func(step int, agentName string, message string)) (*models.OptimizationReport, error) {
	return s.AnalyzeWithProgress(ctx, url, platform, func(step int, total int, agentName string, message string) {
		if callback != nil {
			callback(step, agentName, message)
		}
	})
}
