package agents

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

// NewMainQueryExtractorAgent 3. 主查询提取 Agent
func NewMainQueryExtractorAgent(llmModel model.ToolCallingChatModel) (adk.Agent, error) {
	ctx := context.Background()

	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "main_query_extractor",
		Description: "从相关查询中提取核心搜索词",
		Model:       llmModel,
		OutputKey:   "MainQuery",
		Instruction: `你是查询分析专家。你的目标是识别用户最核心的搜索意图。

**任务说明**：
请查看对话历史中的上下文信息，特别是 query_fanout_researcher 输出的相关查询列表。

请基于这些信息：
1. 分析这些查询的共同主题
2. 提取最核心的主查询词（1-3个）
3. 识别搜索关键词
4. 判断搜索意图

请返回格式：
主查询: [核心查询词]
关键词: [关键词列表]
意图: [信息查询/交易/导航等]`,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 main_query_extractor agent 失败: %w", err)
	}

	return a, nil
}
