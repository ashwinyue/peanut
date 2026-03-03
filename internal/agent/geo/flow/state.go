/*
 * Copyright 2025 Peanut Authors
 *
 * GEO Agent State - 导出到 models 包
 */

package flow

import (
	"context"
	"fmt"

	"github.com/solariswu/peanut/internal/agent/geo/models"
)

// State 导出 models.FlowState
type State = models.FlowState

// SearchResult 导出 models.SearchResult
type SearchResult = models.SearchResult

// progressCallbackKey 是存储进度回调的上下文键
type progressCallbackKey struct{}

// WithProgressCallback 将进度回调添加到上下文
func WithProgressCallback(ctx context.Context, callback func(step int, total int, agentName string, message string)) context.Context {
	return context.WithValue(ctx, progressCallbackKey{}, callback)
}

// GetProgressCallback 从上下文获取进度回调
func GetProgressCallback(ctx context.Context) func(step int, total int, agentName string, message string) {
	if cb, ok := ctx.Value(progressCallbackKey{}).(func(step int, total int, agentName string, message string)); ok {
		return cb
	}
	return nil
}

// GenLocalState 生成 Local State 的工厂函数
func GenLocalState(ctx context.Context) *State {
	fmt.Println("[GEO] GenLocalState 被调用")
	state := models.GenFlowState(ctx)

	// 从上下文中获取进度回调并设置到 state
	if callback := GetProgressCallback(ctx); callback != nil {
		state.OnProgress = callback
		state.TotalSteps = 7
		fmt.Println("[GEO] GenLocalState: 进度回调已设置")
	}

	return state
}
