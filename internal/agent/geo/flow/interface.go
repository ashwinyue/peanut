/*
 * Copyright 2025 Peanut Authors
 *
 * GEO Agent 接口定义
 */

package flow

import (
	"context"

	"github.com/solariswu/peanut/internal/agent/geo/models"
)

// AgentService GEO Agent 服务接口
type AgentService interface {
	Analyze(ctx context.Context, url string, platformType string) (*models.OptimizationReport, error)
	AnalyzeWithProgress(ctx context.Context, url string, platformType string, progress func(step int, total int, agentName string, message string)) (*models.OptimizationReport, error)
}
