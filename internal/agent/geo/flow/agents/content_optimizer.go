/*
 * Copyright 2025 Peanut Authors
 *
 * Content Optimizer Agent - 内容优化
 */

package agents

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/solariswu/peanut/internal/agent/geo/llm"
	"github.com/solariswu/peanut/internal/agent/geo/models"
)

// loadContentOptimizerPrompt 加载 prompt
func loadContentOptimizerPrompt(ctx context.Context, state *models.FlowState) ([]*schema.Message, error) {
	sysPrompt, err := GetPromptTemplate("content_optimizer")
	if err != nil {
		sysPrompt = defaultContentOptimizerPrompt
	}

	promptTemp := prompt.FromMessages(schema.Jinja2,
		schema.SystemMessage(sysPrompt),
	)

	variables := map[string]any{
		"platform":      state.PlatformType,
		"main_query":    state.MainQuery,
		"ai_overview":   state.AIOverview,
		"query_summary": state.QuerySummary,
	}

	return promptTemp.Format(ctx, variables)
}

const defaultContentOptimizerPrompt = `你是 GEO 内容优化专家。

## 任务
1. 对比分析 Query Summary 与 Google AI Overview 的差距
2. 识别两者的共性和差异
3. 生成可操作的优化建议（action items）
4. 输出 Markdown 格式的对比报告

## 输出格式要求
请生成一份 Markdown 格式的优化报告，必须包含以下内容：

1. 执行摘要 - 总体对比结论
2. 对比表格（必须包含以下列）：
   | Aspect | Query Summary | Google AI Overview | Similarities/Patterns | Differences |
3. 关键发现 - 主要共性和差异点
4. Action Items - 基于对比的具体优化建议（按优先级排序）
5. 结论

注意：表格必须完整呈现，包含所有对比维度。`

// routerContentOptimizer 路由函数
func routerContentOptimizer(ctx context.Context, input *schema.Message, state *models.FlowState) (string, error) {
	// 直接保存 Markdown 报告
	state.Report = &models.OptimizationReport{
		URL:                state.URL,
		Title:              state.Title,
		OverallScore:       0, // Markdown 报告不需要数字评分
		OptimizationReport: input.Content,
	}
	state.Step = 6

	state.Goto = AgentContentRewriter
	return state.Goto, nil
}

// NewContentOptimizerAgent 创建 Content Optimizer Agent
func NewContentOptimizerAgent(ctx context.Context) *compose.Graph[string, string] {
	cag := compose.NewGraph[string, string]()

	llmModel, err := llm.NewChatModel(ctx)
	if err != nil {
		panic(fmt.Sprintf("创建 LLM 模型失败: %v", err))
	}

	_ = cag.AddLambdaNode("load", compose.InvokableLambdaWithOption(func(ctx context.Context, input string, opts ...any) ([]*schema.Message, error) {
		var state *models.FlowState
		if err := compose.ProcessState[*models.FlowState](ctx, func(_ context.Context, s *models.FlowState) error {
			state = s
			return nil
		}); err != nil {
			return nil, err
		}
		return loadContentOptimizerPrompt(ctx, state)
	}))

	_ = cag.AddChatModelNode("agent", llmModel)

	_ = cag.AddLambdaNode("router", compose.InvokableLambdaWithOption(func(ctx context.Context, input *schema.Message, opts ...any) (string, error) {
		var next string
		err := compose.ProcessState[*models.FlowState](ctx, func(_ context.Context, state *models.FlowState) error {
			var err error
			next, err = routerContentOptimizer(ctx, input, state)
			return err
		})
		return next, err
	}))

	_ = cag.AddEdge(compose.START, "load")
	_ = cag.AddEdge("load", "agent")
	_ = cag.AddEdge("agent", "router")
	_ = cag.AddEdge("router", compose.END)

	return cag
}
