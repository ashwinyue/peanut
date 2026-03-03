/*
 * Copyright 2025 Peanut Authors
 *
 * GEO Agent State - 导出到 models 包
 */

package flow

import (
	"context"

	"github.com/solariswu/peanut/internal/agent/geo/models"
)

// State 导出 models.FlowState
type State = models.FlowState

// SearchResult 导出 models.SearchResult
type SearchResult = models.SearchResult

// GenLocalState 生成 Local State 的工厂函数
func GenLocalState(ctx context.Context) *State {
	return models.GenFlowState(ctx)
}
