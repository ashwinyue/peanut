package geo

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/solariswu/peanut/internal/agent/geo/models"
)

// AgentService Agent 服务接口
type AgentService interface {
	ExecuteWithStreaming(ctx context.Context, url string, platform string, callback func(step int, agentName string, message string)) (*models.OptimizationReport, error)
}

// Service GEO 服务
type Service struct {
	agent adk.Agent
}

// NewService 创建新的 GEO 服务（使用指定平台配置）
func NewService(platform string) (*Service, error) {
	agent, err := NewChain(platform)
	if err != nil {
		return nil, fmt.Errorf("创建 Chain 失败: %w", err)
	}

	return &Service{
		agent: agent,
	}, nil
}

// NewDefaultService 创建默认 GEO 服务（豆包元宝平台）
func NewDefaultService() (*Service, error) {
	return NewService("doubao")
}

// AnalyzeURL 分析 URL（完整流程，使用指定平台）
func (s *Service) AnalyzeURL(ctx context.Context, url string, platform string) (*models.OptimizationReport, error) {
	return Execute(ctx, s.agent, url, platform)
}

// AnalyzeURLWithProgress 分析 URL（带进度回调，使用指定平台）
func (s *Service) AnalyzeURLWithProgress(ctx context.Context, url string, platform string, progress func(step int, total int, agentName string, message string)) (*models.OptimizationReport, error) {
	totalSteps := 7
	currentStep := 0

	return ExecuteWithStreaming(ctx, s.agent, url, platform, func(step int, agentName string, message string) {
		currentStep = step
		if progress != nil {
			progress(currentStep, totalSteps, agentName, message)
		}
	})
}

// ExecuteWithStreaming 实现 AgentService 接口
func (s *Service) ExecuteWithStreaming(ctx context.Context, url string, platform string, callback func(step int, agentName string, message string)) (*models.OptimizationReport, error) {
	return ExecuteWithStreaming(ctx, s.agent, url, platform, callback)
}
