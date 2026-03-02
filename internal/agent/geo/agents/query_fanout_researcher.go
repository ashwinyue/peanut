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

// NewQueryFanoutResearcherAgent 2. 查询发散研究 Agent
func NewQueryFanoutResearcherAgent(searcher tools.Searcher) (adk.Agent, error) {
	return NewQueryFanoutResearcherAgentWithModel(searcher, nil)
}

// NewQueryFanoutResearcherAgentWithModel 2. 查询发散研究 Agent（带模型）
func NewQueryFanoutResearcherAgentWithModel(searcher tools.Searcher, llmModel model.ToolCallingChatModel) (adk.Agent, error) {
	ctx := context.Background()

	// 创建搜索工具
	searcherTool, err := NewSearcherToolAdapter(searcher)
	if err != nil {
		return nil, fmt.Errorf("创建搜索工具失败: %w", err)
	}

	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "query_fanout_researcher",
		Description: "基于标题进行搜索，获取相关查询",
		Instruction: `你是搜索研究专家。你的目标是发现与主题相关的所有潜在搜索查询。

前一个 agent 已经爬取了网页标题：
{Title}  // ← 占位符，会被前一个 agent 的输出替换

请基于这个标题：
1. 使用 search_queries 工具进行搜索
2. 分析搜索结果
3. 总结出相关查询列表（至少5个）

请返回格式：
- 原始查询
- 相关查询列表（编号）`,
		Model:     llmModel,
		OutputKey: "QueryFanout",
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{searcherTool},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("创建 query_fanout_researcher agent 失败: %w", err)
	}

	return a, nil
}
