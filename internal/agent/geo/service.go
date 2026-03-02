package geo

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/solariswu/peanut/internal/agent/geo/models"
)

// AgentService Agent 服务接口
type AgentService interface {
	ExecuteWithStreaming(ctx context.Context, url string, callback func(step int, agentName string, message string)) (*models.OptimizationReport, error)
}

// Service GEO 服务
type Service struct {
	agent adk.Agent
}

// NewService 创建新的 GEO 服务（使用默认配置）
func NewService() (*Service, error) {
	agent, err := NewDefaultChain()
	if err != nil {
		return nil, fmt.Errorf("创建 Chain 失败: %w", err)
	}

	return &Service{
		agent: agent,
	}, nil
}

// AnalyzeURL 分析 URL（完整流程）
func (s *Service) AnalyzeURL(ctx context.Context, url string) (*models.OptimizationReport, error) {
	return Execute(ctx, s.agent, url)
}

// AnalyzeURLWithProgress 分析 URL（带进度回调）
func (s *Service) AnalyzeURLWithProgress(ctx context.Context, url string, progress func(step int, total int, agentName string, message string)) (*models.OptimizationReport, error) {
	totalSteps := 7
	currentStep := 0

	return ExecuteWithStreaming(ctx, s.agent, url, func(step int, agentName string, message string) {
		currentStep = step
		if progress != nil {
			progress(currentStep, totalSteps, agentName, message)
		}
	})
}

// ExecuteWithStreaming 实现 AgentService 接口
func (s *Service) ExecuteWithStreaming(ctx context.Context, url string, callback func(step int, agentName string, message string)) (*models.OptimizationReport, error) {
	return ExecuteWithStreaming(ctx, s.agent, url, callback)
}
