/*
 * Copyright 2025 Peanut Authors
 *
 * GEO Agent 常量定义
 * 参考: eino-examples/flow/agent/deer-go/biz/consts/consts.go
 */

package flow

// Agent 名称常量
const (
	// Agent 名称
	AgentTitleScraper        = "title_scraper"
	AgentQueryResearcher     = "query_researcher"
	AgentMainQueryExtractor  = "main_query_extractor"
	AgentAIOverviewRetriever = "ai_overview_retriever"
	AgentQuerySummarizer     = "query_summarizer"
	AgentContentOptimizer    = "content_optimizer"
	AgentContentRewriter     = "content_rewriter"
	AgentValidator           = "content_validator"

	// 流程控制
	StepStart = "start"
	StepEnd   = "end"
)
