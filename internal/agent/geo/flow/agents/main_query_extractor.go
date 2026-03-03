/*
 * Copyright 2025 Peanut Authors
 *
 * Main Query Extractor Agent - 主查询提取
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

	"github.com/solariswu/peanut/internal/agent/geo/llm"
	"github.com/solariswu/peanut/internal/agent/geo/models"
)

// MainQueryResult 主查询结果
type MainQueryResult struct {
	MainQuery    string   `json:"main_query"`
	Keywords     []string `json:"keywords"`
	SearchIntent string   `json:"search_intent"`
	Reasoning    string   `json:"reasoning"`
}

// loadMainQueryExtractorPrompt 加载 prompt
func loadMainQueryExtractorPrompt(ctx context.Context, state *models.FlowState) ([]*schema.Message, error) {
	sysPrompt, err := GetPromptTemplate("main_query_extractor")
	if err != nil {
		sysPrompt = defaultMainQueryExtractorPrompt
	}

	promptTemp := prompt.FromMessages(schema.Jinja2,
		schema.SystemMessage(sysPrompt),
	)

	variables := map[string]any{
		"title":   state.Title,
		"queries": strings.Join(state.QueryFanout, ", "),
		"results": formatSearchResults(state.SearchResults),
	}

	return promptTemp.Format(ctx, variables)
}

func formatSearchResults(results []models.SearchResult) string {
	var sb strings.Builder
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", r.Title, r.Snippet))
	}
	return sb.String()
}

const defaultMainQueryExtractorPrompt = `你是查询分析专家。

## 任务
1. 分析相关查询的共同主题
2. 提取最核心的主查询词（1-3个）
3. 识别搜索关键词
4. 判断搜索意图

## 输出格式
{
  "main_query": "核心查询词",
  "keywords": ["关键词1", "关键词2"],
  "search_intent": "信息查询/交易/导航",
  "reasoning": "分析理由"
}`

// routerMainQueryExtractor 路由函数
func routerMainQueryExtractor(ctx context.Context, input *schema.Message, state *models.FlowState) (string, error) {
	result := parseMainQueryResult(input.Content)

	state.MainQuery = result.MainQuery
	state.Keywords = result.Keywords
	state.SearchIntent = result.SearchIntent
	state.Step = 3

	// 发送进度回调
	if state.OnProgress != nil {
		state.OnProgress(3, state.TotalSteps, "主查询提取", "处理完成")
	}

	state.Goto = AgentAIOverviewRetriever
	return state.Goto, nil
}

func parseMainQueryResult(content string) *MainQueryResult {
	result := &MainQueryResult{}
	if err := json.Unmarshal([]byte(content), &result); err == nil {
		return result
	}
	// 降级处理
	result.MainQuery = extractJSONField(content, "main_query")
	result.Keywords = extractStringArray(content, "keywords")
	result.SearchIntent = extractJSONField(content, "search_intent")
	return result
}

// NewMainQueryExtractorAgent 创建主查询提取 Agent
func NewMainQueryExtractorAgent[I, O any](ctx context.Context) *compose.Graph[I, O] {
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
		return loadMainQueryExtractorPrompt(ctx, state)
	}))

	_ = cag.AddChatModelNode("agent", llmModel)

	_ = cag.AddLambdaNode("router", compose.InvokableLambdaWithOption(func(ctx context.Context, input *schema.Message, opts ...any) (string, error) {
		var next string
		err := compose.ProcessState[*models.FlowState](ctx, func(_ context.Context, state *models.FlowState) error {
			var err error
			next, err = routerMainQueryExtractor(ctx, input, state)
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
