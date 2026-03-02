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

// NewAIOverviewRetrieverAgent 4. AI Overview 获取 Agent
func NewAIOverviewRetrieverAgent(serp tools.SERPProvider) (adk.Agent, error) {
	return NewAIOverviewRetrieverAgentWithModel(serp, nil)
}

// NewAIOverviewRetrieverAgentWithModel 4. AI Overview 获取 Agent（带模型）
func NewAIOverviewRetrieverAgentWithModel(serp tools.SERPProvider, llmModel model.ToolCallingChatModel) (adk.Agent, error) {
	ctx := context.Background()

	// 创建 SERP 工具
	serpTool, err := NewSERPToolAdapter(serp)
	if err != nil {
		return nil, fmt.Errorf("创建 SERP 工具失败: %w", err)
	}

	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "ai_overview_retriever",
		Description: "获取豆包元宝 AI 摘要内容",
		Instruction: `你是豆包元宝 AI 摘要专家。你的目标是模拟豆包元宝（字节跳动的生成式搜索引擎）的 AI 摘要功能。

**任务说明**：
前一个 agent 已经提取了主查询，输出如下：
{MainQuery}

基于这个主查询：
1. 使用 get_ai_overview 工具获取相关信息
2. 模拟豆包元宝生成 AI 摘要（注重权威性、时效性、结构化）
3. 提取关键信息

豆包元宝 AI 摘要特点：
- 权威性：优先引用官方来源、学术机构、权威媒体
- 时效性：标注信息发布时间
- 结构化：使用列表、表格等清晰呈现
- 中文优化：语言表达符合中文阅读习惯

请返回格式：
- 查询词
- AI 摘要内容（结构化呈现）
- 引用来源列表（标注来源类型）`,
		Model:     llmModel,
		OutputKey: "AIOverview",
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{serpTool},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("创建 ai_overview_retriever agent 失败: %w", err)
	}

	return a, nil
}
