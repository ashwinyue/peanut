/*
 * Copyright 2025 Peanut Authors
 *
 * GEO Flow Builder - 构建 GEO Agent Graph
 */

package flow

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/compose"

	"github.com/solariswu/peanut/internal/agent/geo/flow/agents"
	"github.com/solariswu/peanut/internal/agent/geo/models"
	"github.com/solariswu/peanut/internal/agent/geo/tools"
)

// agentHandOff 子图流转函数
// 参考 deer-go: branch 函数的 input 类型是 Graph 的输出类型 (string)，不是 *State
// 需要通过 compose.ProcessState 来访问 state
func agentHandOff(ctx context.Context, input string) (next string, err error) {
	fmt.Printf("[agentHandOff] 被调用, input=%s\n", input)
	err = compose.ProcessState[*State](ctx, func(_ context.Context, state *State) error {
		if state.Goto == "" {
			next = compose.END
			fmt.Println("[agentHandOff] state.Goto 为空, 跳转到 END")
		} else {
			next = state.Goto
			fmt.Printf("[agentHandOff] 从 state.Goto 获取下一步: %s\n", next)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("[agentHandOff] ProcessState 失败: %v\n", err)
	}
	fmt.Printf("[agentHandOff] 返回 next=%s\n", next)
	return next, err
}

// BuildGraph 构建 GEO Flow Graph（使用默认 CheckPointStore）
// Deprecated: 使用 BuildGraphWithCheckpoint 代替
func BuildGraph[I, O, S any](ctx context.Context, genLocalState func(ctx context.Context) S) (compose.Runnable[I, O], error) {
	return BuildGraphWithCheckpoint[I, O, S](ctx, genLocalState, models.NewGEOCheckPoint(ctx))
}

// BuildGraphWithCheckpoint 构建 GEO Flow Graph（使用指定的 CheckPointStore）
func BuildGraphWithCheckpoint[I, O, S any](ctx context.Context, genLocalState func(ctx context.Context) S, checkPointStore compose.CheckPointStore) (compose.Runnable[I, O], error) {
	// 初始化工具
	scraper, err := tools.NewBrightDataWebScraper()
	if err != nil {
		return nil, fmt.Errorf("创建 scraper tool 失败: %w", err)
	}

	searcher, err := tools.NewBrightDataSearcher()
	if err != nil {
		return nil, fmt.Errorf("创建 searcher tool 失败: %w", err)
	}

	serp, err := tools.NewBrightDataSERPProvider()
	if err != nil {
		return nil, fmt.Errorf("创建 serp tool 失败: %w", err)
	}

	// 创建 Graph
	g := compose.NewGraph[I, O](
		compose.WithGenLocalState(genLocalState),
	)

	// 定义所有可能的节点
	outMap := map[string]bool{
		AgentTitleScraper:        true,
		AgentQueryResearcher:     true,
		AgentMainQueryExtractor:  true,
		AgentAIOverviewRetriever: true,
		AgentQuerySummarizer:     true,
		AgentContentOptimizer:    true,
		AgentContentRewriter:     true,
		compose.END:              true,
	}

	// 创建各 Agent 子图
	titleScraperGraph := agents.NewTitleScraperAgent[I, O](ctx, scraper)
	queryResearcherGraph := agents.NewQueryResearcherAgent[I, O](ctx, searcher)
	mainQueryExtractorGraph := agents.NewMainQueryExtractorAgent[I, O](ctx)
	aiOverviewRetrieverGraph := agents.NewAIOverviewRetrieverAgent[I, O](ctx, serp)
	querySummarizerGraph := agents.NewQuerySummarizerAgent[I, O](ctx)
	contentOptimizerGraph := agents.NewContentOptimizerAgent[I, O](ctx)
	contentRewriterGraph := agents.NewContentRewriterAgent[I, O](ctx)

	// 添加节点到 Graph
	_ = g.AddGraphNode(AgentTitleScraper, titleScraperGraph, compose.WithNodeName(AgentTitleScraper))
	_ = g.AddGraphNode(AgentQueryResearcher, queryResearcherGraph, compose.WithNodeName(AgentQueryResearcher))
	_ = g.AddGraphNode(AgentMainQueryExtractor, mainQueryExtractorGraph, compose.WithNodeName(AgentMainQueryExtractor))
	_ = g.AddGraphNode(AgentAIOverviewRetriever, aiOverviewRetrieverGraph, compose.WithNodeName(AgentAIOverviewRetriever))
	_ = g.AddGraphNode(AgentQuerySummarizer, querySummarizerGraph, compose.WithNodeName(AgentQuerySummarizer))
	_ = g.AddGraphNode(AgentContentOptimizer, contentOptimizerGraph, compose.WithNodeName(AgentContentOptimizer))
	_ = g.AddGraphNode(AgentContentRewriter, contentRewriterGraph, compose.WithNodeName(AgentContentRewriter))

	// 添加分支
	_ = g.AddBranch(AgentTitleScraper, compose.NewGraphBranch(agentHandOff, outMap))
	_ = g.AddBranch(AgentQueryResearcher, compose.NewGraphBranch(agentHandOff, outMap))
	_ = g.AddBranch(AgentMainQueryExtractor, compose.NewGraphBranch(agentHandOff, outMap))
	_ = g.AddBranch(AgentAIOverviewRetriever, compose.NewGraphBranch(agentHandOff, outMap))
	_ = g.AddBranch(AgentQuerySummarizer, compose.NewGraphBranch(agentHandOff, outMap))
	_ = g.AddBranch(AgentContentOptimizer, compose.NewGraphBranch(agentHandOff, outMap))
	_ = g.AddBranch(AgentContentRewriter, compose.NewGraphBranch(agentHandOff, outMap))

	// 设置起始节点
	_ = g.AddEdge(compose.START, AgentTitleScraper)

	// 编译 Graph
	runnable, err := g.Compile(ctx,
		compose.WithGraphName("GEOFlow"),
		compose.WithNodeTriggerMode(compose.AnyPredecessor),
		compose.WithCheckPointStore(checkPointStore),
	)
	if err != nil {
		return nil, err
	}

	return runnable, nil
}
