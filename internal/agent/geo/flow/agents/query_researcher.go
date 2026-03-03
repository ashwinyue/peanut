/*
 * Copyright 2025 Peanut Authors
 *
 * Query Researcher Agent - 查询发散研究
 * 参考: eino-examples/flow/agent/deer-go/biz/eino/researcher.go
 */

package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"

	"github.com/solariswu/peanut/internal/agent/geo/llm"
	"github.com/solariswu/peanut/internal/agent/geo/models"
)

// QueryResearcherResult 研究结果
type QueryResearcherResult struct {
	OriginalQuery  string                `json:"original_query"`
	RelatedQueries []string              `json:"related_queries"`
	SearchResults  []models.SearchResult `json:"search_results"`
}

// loadQueryResearcherPrompt 加载 prompt
func loadQueryResearcherPrompt(ctx context.Context, state *models.FlowState) ([]*schema.Message, error) {
	sysPrompt, err := GetPromptTemplate("query_researcher")
	if err != nil {
		sysPrompt = defaultQueryResearcherPrompt
	}

	promptTemp := prompt.FromMessages(schema.Jinja2,
		schema.SystemMessage(sysPrompt),
		schema.UserMessage("开始研究"),
	)

	variables := map[string]any{
		"title":   state.Title,
		"content": state.Content,
	}

	return promptTemp.Format(ctx, variables)
}

// defaultQueryResearcherPrompt 默认 prompt
const defaultQueryResearcherPrompt = `你是搜索研究专家。

## 任务
1. 基于网页标题和内容，分析核心主题
2. 使用 search_queries 工具进行搜索
3. 收集相关查询列表（至少5个）

## 输出格式
请以 JSON 格式返回：
{
  "original_query": "原始查询词",
  "related_queries": ["查询1", "查询2", ...],
  "search_results": [{"title": "...", "url": "...", "snippet": "..."}]
}`

// routerQueryResearcher 路由函数
func routerQueryResearcher(ctx context.Context, input *schema.Message, state *models.FlowState) (string, error) {
	result := parseQueryResearcherResult(input.Content)

	// 保存到 State
	state.QueryFanout = result.RelatedQueries
	state.SearchResults = result.SearchResults
	state.Step = 2

	// 发送进度回调
	if state.OnProgress != nil {
		state.OnProgress(2, state.TotalSteps, "查询发散", "处理完成")
	}

	state.Goto = AgentMainQueryExtractor
	return state.Goto, nil
}

// parseQueryResearcherResult 解析结果
func parseQueryResearcherResult(content string) *QueryResearcherResult {
	result := &QueryResearcherResult{}

	// 尝试解析 JSON
	if err := json.Unmarshal([]byte(content), &result); err == nil {
		return result
	}

	// 降级处理：从文本中提取
	result.OriginalQuery = extractJSONField(content, "original_query")
	result.RelatedQueries = extractStringArray(content, "related_queries")

	return result
}

// extractJSONField 提取 JSON 字段
func extractJSONField(content, field string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err == nil {
		if val, ok := data[field].(string); ok {
			return val
		}
	}
	return ""
}

// extractStringArray 提取字符串数组
func extractStringArray(content, field string) []string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err == nil {
		if arr, ok := data[field].([]interface{}); ok {
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return []string{}
}

// NewQueryResearcherAgent 创建 Query Researcher Agent
func NewQueryResearcherAgent[I, O any](ctx context.Context, searchTool tool.InvokableTool) *compose.Graph[I, O] {
	cag := compose.NewGraph[I, O]()

	// 创建 LLM 模型
	llmModel, err := llm.NewChatModel(ctx)
	if err != nil {
		panic(fmt.Sprintf("创建 LLM 模型失败: %v", err))
	}

	// 获取工具信息
	toolInfo, err := searchTool.Info(ctx)
	if err != nil {
		panic(fmt.Sprintf("获取工具信息失败: %v", err))
	}

	// 为模型添加工具
	modelWithTools, err := llmModel.WithTools([]*schema.ToolInfo{toolInfo})
	if err != nil {
		panic(fmt.Sprintf("添加工具失败: %v", err))
	}

	// 创建 ReAct Agent
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		MaxStep:          15,
		ToolCallingModel: modelWithTools,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{searchTool},
		},
	})
	if err != nil {
		panic(fmt.Sprintf("创建 ReAct Agent 失败: %v", err))
	}

	// 包装为 Lambda
	agentLambda, err := compose.AnyLambda(agent.Generate, agent.Stream, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("包装 Agent 失败: %v", err))
	}

	// 添加 load 节点
	_ = cag.AddLambdaNode("load", compose.InvokableLambdaWithOption(func(ctx context.Context, input string, opts ...any) ([]*schema.Message, error) {
		var state *models.FlowState
		if err := compose.ProcessState[*models.FlowState](ctx, func(_ context.Context, s *models.FlowState) error {
			state = s
			return nil
		}); err != nil {
			return nil, err
		}
		return loadQueryResearcherPrompt(ctx, state)
	}))

	// 添加 agent 节点
	_ = cag.AddLambdaNode("agent", agentLambda)

	// 添加 router 节点
	_ = cag.AddLambdaNode("router", compose.InvokableLambdaWithOption(func(ctx context.Context, input *schema.Message, opts ...any) (string, error) {
		var next string
		err := compose.ProcessState[*models.FlowState](ctx, func(_ context.Context, state *models.FlowState) error {
			var err error
			next, err = routerQueryResearcher(ctx, input, state)
			return err
		})
		return next, err
	}))

	// 添加边
	_ = cag.AddEdge(compose.START, "load")
	_ = cag.AddEdge("load", "agent")
	_ = cag.AddEdge("agent", "router")
	_ = cag.AddEdge("router", compose.END)

	return cag
}
