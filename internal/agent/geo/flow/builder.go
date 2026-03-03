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
	"github.com/solariswu/peanut/internal/agent/geo/tools"
)

// agentHandOff 子图流转函数
func agentHandOff(ctx context.Context, input string) (next string, err error) {
	var gotoAgent string
	_ = compose.ProcessState[*State](ctx, func(_ context.Context, state *State) error {
		gotoAgent = state.Goto
		return nil
	})
	if gotoAgent == "" {
		gotoAgent = compose.END
	}
	return gotoAgent, nil
}

// BuildGraph 构建 GEO Flow Graph
func BuildGraph(ctx context.Context) (compose.Runnable[string, string], error) {
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
	g := compose.NewGraph[string, string](
		compose.WithGenLocalState(GenLocalState),
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
	titleScraperGraph := agents.NewTitleScraperAgent(ctx, scraper)
	queryResearcherGraph := agents.NewQueryResearcherAgent(ctx, searcher)
	mainQueryExtractorGraph := agents.NewMainQueryExtractorAgent(ctx)
	aiOverviewRetrieverGraph := agents.NewAIOverviewRetrieverAgent(ctx, serp)
	querySummarizerGraph := agents.NewQuerySummarizerAgent(ctx)
	contentOptimizerGraph := agents.NewContentOptimizerAgent(ctx)
	contentRewriterGraph := agents.NewContentRewriterAgent(ctx)

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
	)
	if err != nil {
		return nil, err
	}

	return runnable, nil
}
