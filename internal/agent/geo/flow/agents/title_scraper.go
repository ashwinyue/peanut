/*
 * Copyright 2025 Peanut Authors
 *
 * Title Scraper Agent - 网页标题爬取
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

	"github.com/solariswu/peanut/internal/agent/geo/models"
	"github.com/solariswu/peanut/internal/agent/geo/llm"
)

// TitleScraperResult 爬取结果
type TitleScraperResult struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	H1      string `json:"h1"`
	Content string `json:"content"`
}

// loadTitleScraperPrompt 从 State 加载 prompt
func loadTitleScraperPrompt(ctx context.Context, state *models.FlowState, scraperTool tool.BaseTool) ([]*schema.Message, error) {
	// 读取 prompt 模板
	sysPrompt, err := GetPromptTemplate("title_scraper")
	if err != nil {
		// 使用默认 prompt
		sysPrompt = defaultTitleScraperPrompt
	}

	promptTemp := prompt.FromMessages(schema.Jinja2,
		schema.SystemMessage(sysPrompt),
		schema.UserMessage("请爬取这个网页: {{url}}"),
	)

	variables := map[string]any{
		"url": state.URL,
	}

	return promptTemp.Format(ctx, variables)
}

// defaultTitleScraperPrompt 默认 prompt
const defaultTitleScraperPrompt = `你是网页爬虫专家。使用 scrape_webpage 工具爬取网页并提取标题和主要内容。

请以 JSON 格式返回:
{
  "url": "网页URL",
  "title": "网页标题",
  "h1": "主标题",
  "content": "正文摘要"
}`

// routerTitleScraper 路由函数 - 保存结果并决定下一步
func routerTitleScraper(ctx context.Context, input *schema.Message, state *models.FlowState) (string, error) {
	// 解析结果
	result := parseTitleScraperResult(input.Content)

	// 保存到 State
	state.Title = result.Title
	state.Content = result.Content
	state.Step = 1

	// 决定下一步
	if result.Title == "" {
		state.LastError = "无法提取网页标题"
		return AgentQueryResearcher, nil // 继续尝试
	}

	// 成功，进入下一步
	state.Goto = AgentQueryResearcher
	return state.Goto, nil
}

// parseTitleScraperResult 解析爬取结果
func parseTitleScraperResult(content string) *TitleScraperResult {
	result := &TitleScraperResult{}

	// 尝试解析 JSON
	var jsonResult map[string]string
	if err := json.Unmarshal([]byte(content), &jsonResult); err == nil {
		result.URL = jsonResult["url"]
		result.Title = jsonResult["title"]
		result.H1 = jsonResult["h1"]
		result.Content = jsonResult["content"]
		return result
	}

	// 降级：从文本中提取
	result.Title = extractField(content, "title", "标题")
	result.H1 = extractField(content, "h1", "主标题")
	result.Content = extractField(content, "content", "正文")

	return result
}

// extractField 从文本中提取字段
func extractField(content, field, alias string) string {
	// 简单提取：查找 "field": "value" 或 field: value 模式
	_ = []string{
		fmt.Sprintf(`"%s"[:\s]+"([^"]+)"`, field),
		fmt.Sprintf(`"%s"[:\s]+'([^']+)'`, field),
		fmt.Sprintf(`%s[:\s]+([^\n]+)`, field),
	}

	// 使用简化实现直接查找
	if idx := findFieldIndex(content, field); idx >= 0 {
		return extractValue(content, idx)
	}

	return ""
}

func findFieldIndex(content, field string) int {
	searchStr := fmt.Sprintf(`"%s"`, field)
	for i := 0; i <= len(content)-len(searchStr); i++ {
		if content[i:i+len(searchStr)] == searchStr {
			return i
		}
	}
	return -1
}

func extractValue(content string, startIdx int) string {
	// 查找冒号后的值
	colonIdx := -1
	for i := startIdx; i < len(content) && i < startIdx+50; i++ {
		if content[i] == ':' {
			colonIdx = i
			break
		}
	}
	if colonIdx < 0 {
		return ""
	}

	// 查找引号
	quoteStart := -1
	for i := colonIdx; i < len(content) && i < colonIdx+20; i++ {
		if content[i] == '"' {
			quoteStart = i + 1
			break
		}
	}
	if quoteStart < 0 {
		return ""
	}

	// 查找结束引号
	quoteEnd := -1
	for i := quoteStart; i < len(content); i++ {
		if content[i] == '"' && content[i-1] != '\\' {
			quoteEnd = i
			break
		}
	}
	if quoteEnd < 0 {
		return ""
	}

	return content[quoteStart:quoteEnd]
}

// NewTitleScraperAgent 创建 Title Scraper Agent 子图
func NewTitleScraperAgent(ctx context.Context, scraperTool tool.InvokableTool) *compose.Graph[string, string] {
	cag := compose.NewGraph[string, string]()

	// 创建 LLM 模型
	llmModel, err := llm.NewChatModel(ctx)
	if err != nil {
		panic(fmt.Sprintf("创建 LLM 模型失败: %v", err))
	}

	// 为模型添加工具
	toolInfo, err := scraperTool.Info(ctx)
	if err != nil {
		panic(fmt.Sprintf("获取工具信息失败: %v", err))
	}

	modelWithTools, err := llmModel.WithTools([]*schema.ToolInfo{toolInfo})
	if err != nil {
		panic(fmt.Sprintf("添加工具失败: %v", err))
	}

	// 创建 ReAct Agent
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		MaxStep:          10,
		ToolCallingModel: modelWithTools,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{scraperTool},
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

	// 添加 load 节点 - 准备 prompt
	_ = cag.AddLambdaNode("load", compose.InvokableLambdaWithOption(func(ctx context.Context, input string, opts ...any) ([]*schema.Message, error) {
		var state *models.FlowState
		if err := compose.ProcessState[*models.FlowState](ctx, func(_ context.Context, s *models.FlowState) error {
			state = s
			state.URL = input
			return nil
		}); err != nil {
			return nil, err
		}
		return loadTitleScraperPrompt(ctx, state, scraperTool)
	}))

	// 添加 agent 节点
	_ = cag.AddLambdaNode("agent", agentLambda)

	// 添加 router 节点
	_ = cag.AddLambdaNode("router", compose.InvokableLambdaWithOption(func(ctx context.Context, input *schema.Message, opts ...any) (string, error) {
		var next string
		err := compose.ProcessState[*models.FlowState](ctx, func(_ context.Context, state *models.FlowState) error {
			var err error
			next, err = routerTitleScraper(ctx, input, state)
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
