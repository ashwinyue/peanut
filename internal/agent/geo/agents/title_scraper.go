package agents

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/solariswu/peanut/internal/agent/geo/tools"
)

// NewTitleScraperAgent 1. 标题爬取 Agent
func NewTitleScraperAgent(scraper tools.WebScraper) adk.Agent {
	return NewTitleScraperAgentWithModel(scraper, nil)
}

// NewTitleScraperAgentWithModel 1. 标题爬取 Agent（带模型）
func NewTitleScraperAgentWithModel(scraper tools.WebScraper, llmModel model.ToolCallingChatModel) adk.Agent {
	ctx := context.Background()

	// 创建爬取工具
	scraperTool, err := NewScraperToolAdapter(scraper)
	if err != nil {
		panic(fmt.Sprintf("创建爬取工具失败: %v", err))
	}

	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "title_scraper",
		Description: "爬取网页标题和 H1 标签",
		Instruction: `你是网页爬虫专家。你的目标是提取网页的标题和主要标题（H1）。

当用户给出一个 URL 时，请使用 scrape_webpage 工具爬取该网页并返回：
- URL: 网页地址
- Title: 网页标题
- H1: 主标题内容

请以结构化的格式返回结果。`,
		Model: llmModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{scraperTool},
			},
		},
	})

	if err != nil {
		panic(fmt.Sprintf("创建 title_scraper agent 失败: %v", err))
	}

	return a
}
