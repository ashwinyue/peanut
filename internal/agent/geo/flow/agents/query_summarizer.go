/*
 * Copyright 2025 Peanut Authors
 *
 * Query Summarizer Agent - 查询总结
 */

package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/solariswu/peanut/internal/agent/geo/models"
	"github.com/solariswu/peanut/internal/agent/geo/llm"
)

// QuerySummarizerResult 查询总结结果
type QuerySummarizerResult struct {
	Summary      string   `json:"summary"`
	KeyTopics    []string `json:"key_topics"`
	UserIntents  []string `json:"user_intents"`
	HotKeywords  []string `json:"hot_keywords"`
	Insights     string   `json:"insights"`
}

// loadQuerySummarizerPrompt 加载 prompt
func loadQuerySummarizerPrompt(ctx context.Context, state *models.FlowState) ([]*schema.Message, error) {
	sysPrompt, err := GetPromptTemplate("query_summarizer")
	if err != nil {
		sysPrompt = defaultQuerySummarizerPrompt
	}

	promptTemp := prompt.FromMessages(schema.Jinja2,
		schema.SystemMessage(sysPrompt),
	)

	variables := map[string]any{
		"related_queries": strings.Join(state.QueryFanout, ", "),
	}

	return promptTemp.Format(ctx, variables)
}

const defaultQuerySummarizerPrompt = `你是信息总结专家。

## 任务
1. 分析所有搜索结果的共性和差异
2. 识别热门话题和用户关注点
3. 生成结构化的查询总结
4. 提取关键洞察

## 输出格式
{
  "summary": "查询总结（300-500字）",
  "key_topics": ["主题1", "主题2", "主题3"],
  "user_intents": ["用户意图1", "用户意图2"],
  "hot_keywords": ["热词1", "热词2", "热词3"],
  "insights": "关键洞察"
}`

// routerQuerySummarizer 路由函数
func routerQuerySummarizer(ctx context.Context, input *schema.Message, state *models.FlowState) (string, error) {
	result := parseQuerySummarizerResult(input.Content)

	state.QuerySummary = result.Summary
	state.Step = 5

	state.Goto = AgentContentOptimizer
	return state.Goto, nil
}

func parseQuerySummarizerResult(content string) *QuerySummarizerResult {
	result := &QuerySummarizerResult{}
	if err := json.Unmarshal([]byte(content), &result); err == nil {
		return result
	}
	result.Summary = content
	return result
}

// NewQuerySummarizerAgent 创建 Query Summarizer Agent
func NewQuerySummarizerAgent(ctx context.Context) *compose.Graph[string, string] {
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
		return loadQuerySummarizerPrompt(ctx, state)
	}))

	_ = cag.AddChatModelNode("agent", llmModel)

	_ = cag.AddLambdaNode("router", compose.InvokableLambdaWithOption(func(ctx context.Context, input *schema.Message, opts ...any) (string, error) {
		var next string
		err := compose.ProcessState[*models.FlowState](ctx, func(_ context.Context, state *models.FlowState) error {
			var err error
			next, err = routerQuerySummarizer(ctx, input, state)
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
