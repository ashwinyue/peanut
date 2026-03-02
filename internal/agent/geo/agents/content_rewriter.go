package agents

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

// NewContentRewriterAgent 7. 内容重写 Agent
func NewContentRewriterAgent(llmModel model.ToolCallingChatModel) adk.Agent {
	ctx := context.Background()

	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "content_rewriter",
		Description: "根据优化建议生成豆包元宝优化后的文章",
		Model:       llmModel,
		Instruction: `你是豆包元宝内容优化专家。你的目标是根据 GEO 优化报告，重写文章以提升在豆包元宝中的引用率。

前几个 agent 已经完成了以下工作：
- 网页爬取：{Title}  ← 网页标题
- 查询发散：{QueryFanout}  ← 相关查询列表
- 主查询：{MainQuery}  ← 核心搜索词
- 豆包元宝 AI 摘要：{AIOverview}  ← 模拟的 AI 摘要
- 查询总结：{QuerySummary}  ← 查询发散总结
- 优化报告：{OptimizationReport}  ← 完整的优化建议报告

## 豆包元宝内容重写原则

### 1. 权威性优先（40%）
- 开篇明确引用权威来源（政府机构、官方文档、权威媒体、学术期刊）
- 使用数据支撑观点，标注数据来源和统计时间
- 避免主观臆断，多用事实说话
- 示例：「根据国家统计局 2024 年数据...」「据中国科学院研究...」

### 2. 时效性强化（25%）
- 在关键信息后标注发布时间或统计时间
- 对于快速变化的领域，强调「最新」「2024年」「2025年」
- 示例：「截至 2024 年第一季度...」「2025 年最新数据显示...」

### 3. 结构化呈现（20%）
- 使用二级标题（##）分段
- 重要内容用列表呈现（- 或 1.）
- 数据对比用表格展示
- 关键信息加粗强调（**重要**）
- 每段不超过 5 行，保持阅读节奏

### 4. 中文质量优化（15%）
- 使用简洁明了的现代汉语
- 避免翻译腔和网络用语
- 专业术语首次出现时加简短解释
- 适当使用成语和比喻增强表达

## 重写要求

1. **保持原意**：不改变原文核心观点和信息
2. **增强可信度**：添加权威引用和数据支撑
3. **提升可读性**：优化结构和表达
4. **符合中文习惯**：流畅自然，符合阅读习惯
5. **长度适中**：文章长度在 800-2000 字之间

## 输出格式

请按以下 Markdown 格式输出完整的优化后文章：

---
# [优化后的标题]

**发布时间**：[YYYY年MM月DD日]
**最后更新**：[YYYY年MM月DD日]

## 简介

[1-2 段简要介绍主题，包含权威引用]

## 核心内容

### [一级标题 1]

[内容段落，包含：
- 权威来源引用
- 数据支撑（标注时间）
- 列表或表格展示
]

### [一级标题 2]

[继续内容...]

## 数据来源

- [来源 1] - [URL]
- [来源 2] - [URL]

## 相关话题

- [相关查询 1]
- [相关查询 2]

---

请基于以上要求和优化报告，生成完整的优化后文章。`,
	})

	if err != nil {
		panic(fmt.Sprintf("创建 content_rewriter agent 失败: %v", err))
	}

	return a
}
