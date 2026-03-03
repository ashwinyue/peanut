/*
 * Copyright 2025 Peanut Authors
 *
 * Content Rewriter Agent - 内容重写
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

// loadContentRewriterPrompt 加载 prompt
func loadContentRewriterPrompt(ctx context.Context, state *models.FlowState) ([]*schema.Message, error) {
	sysPrompt, err := GetPromptTemplate("content_rewriter")
	if err != nil {
		sysPrompt = defaultContentRewriterPrompt
	}

	promptTemp := prompt.FromMessages(schema.Jinja2,
		schema.SystemMessage(sysPrompt),
	)

	// 构建优化报告摘要
	reportSummary := buildReportSummary(state.Report)

	variables := map[string]any{
		"platform":            state.PlatformType,
		"platform_type":       getPlatformTypeName(state.PlatformType),
		"title":               state.Title,
		"content":             state.Content,
		"main_query":          state.MainQuery,
		"ai_overview":         state.AIOverview,
		"query_summary":       state.QuerySummary,
		"optimization_report": reportSummary,
	}

	return promptTemp.Format(ctx, variables)
}

func getPlatformTypeName(platform string) string {
	switch platform {
	case "google":
		return "Google AI Overview"
	default:
		return "Google AI Overview"
	}
}

func buildReportSummary(report *models.OptimizationReport) string {
	if report == nil {
		return "暂无优化报告"
	}
	return report.OptimizationReport
}

const defaultContentRewriterPrompt = `你是 GEO 内容重写专家。

## 任务
1. 分析优化报告中的建议
2. 基于原始网页内容和 AI Overview 洞察
3. 重写优化后的完整文章
4. 增强可信度和可读性

## 优化原则
- 权威性优先: 开篇明确引用权威来源
- 时效性强化: 标注时间、使用最新数据
- 结构化呈现: 使用标题、列表、表格
- 覆盖 AI Overview 中的关键信息点
- 保持原文核心观点和信息

## 输出格式
请直接返回优化后的完整文章（Markdown格式）。`

// routerContentRewriter 路由函数
func routerContentRewriter(ctx context.Context, input *schema.Message, state *models.FlowState) (string, error) {
	state.OptimizedArticle = input.Content
	state.Step = 7

	// 发送进度回调
	if state.OnProgress != nil {
		state.OnProgress(7, state.TotalSteps, "文章重写", "处理完成")
	}

	// 更新报告
	if state.Report != nil {
		state.Report.OptimizedArticle = input.Content
	}

	state.Goto = compose.END
	return state.Goto, nil
}

// NewContentRewriterAgent 创建 Content Rewriter Agent
func NewContentRewriterAgent[I, O any](ctx context.Context) *compose.Graph[I, O] {
	cag := compose.NewGraph[I, O]()

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
		return loadContentRewriterPrompt(ctx, state)
	}))

	_ = cag.AddChatModelNode("agent", llmModel)

	_ = cag.AddLambdaNode("router", compose.InvokableLambdaWithOption(func(ctx context.Context, input *schema.Message, opts ...any) (string, error) {
		var next string
		err := compose.ProcessState[*models.FlowState](ctx, func(_ context.Context, state *models.FlowState) error {
			var err error
			next, err = routerContentRewriter(ctx, input, state)
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
