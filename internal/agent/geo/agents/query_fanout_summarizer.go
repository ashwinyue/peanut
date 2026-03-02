package agents

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

// NewQueryFanoutSummarizerAgent 5. 查询发散总结 Agent
func NewQueryFanoutSummarizerAgent(llmModel model.ToolCallingChatModel) (adk.Agent, error) {
	ctx := context.Background()

	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "query_fanout_summarizer",
		Description: "总结查询发散的关键主题",
		Model:       llmModel,
		OutputKey:   "QuerySummary",
		Instruction: `你是内容总结专家。你的目标是从相关查询中提炼关键主题和用户关注点。

前一个 agent 已经完成了查询发散研究：
[QueryFanout]  // ← 会被前一个 agent 的输出替换

请基于这些相关查询：
1. 分析这些查询背后的用户意图
2. 提炼出关键主题（3-5个）
3. 生成简洁的总结（100-200字）

请返回：
【查询发散总结】
[总结内容]

【关键主题】
1. 主题一
2. 主题二
...`,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 query_fanout_summarizer agent 失败: %w", err)
	}

	return a, nil
}
