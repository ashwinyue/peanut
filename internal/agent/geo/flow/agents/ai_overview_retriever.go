/*
 * Copyright 2025 Peanut Authors
 *
 * AI Overview Retriever Agent - AI 摘要获取
 */

package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"

	"github.com/solariswu/peanut/internal/agent/geo/models"
	"github.com/solariswu/peanut/internal/agent/geo/llm"
)

// AIOverviewResult AI 摘要结果
type AIOverviewResult struct {
	Query            string   `json:"query"`
	Summary          string   `json:"summary"`
	Sources          []string `json:"sources"`
	KeyPoints        []string `json:"key_points"`
	ContentStructure string   `json:"content_structure"`
}

// loadAIOOverviewRetrieverPrompt 加载 prompt
func loadAIOOverviewRetrieverPrompt(ctx context.Context, state *models.FlowState) ([]*schema.Message, error) {
	sysPrompt, err := GetPromptTemplate("ai_overview_retriever")
	if err != nil {
		sysPrompt = defaultAIOOverviewRetrieverPrompt
	}

	promptTemp := prompt.FromMessages(schema.Jinja2,
		schema.SystemMessage(sysPrompt),
	)

	variables := map[string]any{
		"platform":   state.PlatformType,
		"main_query": state.MainQuery,
		"keywords":   strings.Join(state.Keywords, ", "),
	}

	return promptTemp.Format(ctx, variables)
}

const defaultAIOOverviewRetrieverPrompt = `你是 AI 搜索引擎摘要专家。

## 任务
1. 基于主查询和关键词，使用 get_ai_overview 工具获取 AI 摘要
2. 分析摘要的内容结构和特点
3. 提取关键信息点

## 输出格式
{
  "query": "查询词",
  "summary": "AI摘要内容",
  "sources": ["来源1", "来源2"],
  "key_points": ["要点1", "要点2"],
  "content_structure": "摘要结构特点"
}`

// routerAIOverviewRetriever 路由函数
func routerAIOverviewRetriever(ctx context.Context, input *schema.Message, state *models.FlowState) (string, error) {
	result := parseAIOverviewResult(input.Content)

	state.AIOverview = result.Summary
	state.Sources = result.Sources
	state.Step = 4

	state.Goto = AgentQuerySummarizer
	return state.Goto, nil
}

func parseAIOverviewResult(content string) *AIOverviewResult {
	result := &AIOverviewResult{}
	if err := json.Unmarshal([]byte(content), &result); err == nil {
		return result
	}
	result.Summary = content
	return result
}

// NewAIOverviewRetrieverAgent 创建 AI Overview Retriever Agent
func NewAIOverviewRetrieverAgent(ctx context.Context, overviewTool tool.InvokableTool) *compose.Graph[string, string] {
	cag := compose.NewGraph[string, string]()

	llmModel, err := llm.NewChatModel(ctx)
	if err != nil {
		panic(fmt.Sprintf("创建 LLM 模型失败: %v", err))
	}

	toolInfo, err := overviewTool.Info(ctx)
	if err != nil {
		panic(fmt.Sprintf("获取工具信息失败: %v", err))
	}

	modelWithTools, err := llmModel.WithTools([]*schema.ToolInfo{toolInfo})
	if err != nil {
		panic(fmt.Sprintf("添加工具失败: %v", err))
	}

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		MaxStep:          10,
		ToolCallingModel: modelWithTools,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{overviewTool},
		},
	})
	if err != nil {
		panic(fmt.Sprintf("创建 ReAct Agent 失败: %v", err))
	}

	agentLambda, err := compose.AnyLambda(agent.Generate, agent.Stream, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("包装 Agent 失败: %v", err))
	}

	_ = cag.AddLambdaNode("load", compose.InvokableLambdaWithOption(func(ctx context.Context, input string, opts ...any) ([]*schema.Message, error) {
		var state *models.FlowState
		if err := compose.ProcessState[*models.FlowState](ctx, func(_ context.Context, s *models.FlowState) error {
			state = s
			return nil
		}); err != nil {
			return nil, err
		}
		return loadAIOOverviewRetrieverPrompt(ctx, state)
	}))

	_ = cag.AddLambdaNode("agent", agentLambda)

	_ = cag.AddLambdaNode("router", compose.InvokableLambdaWithOption(func(ctx context.Context, input *schema.Message, opts ...any) (string, error) {
		var next string
		err := compose.ProcessState[*models.FlowState](ctx, func(_ context.Context, state *models.FlowState) error {
			var err error
			next, err = routerAIOverviewRetriever(ctx, input, state)
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
