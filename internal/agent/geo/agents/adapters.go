package agents

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/solariswu/peanut/internal/agent/geo/tools"
)

// ScrapeRequest 爬取请求
type ScrapeRequest struct {
	URL string `json:"url" jsonschema:"description=要爬取的网页 URL"`
}

// ScrapeResponse 爬取响应
type ScrapeResponse struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	H1    string `json:"h1"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query string `json:"query" jsonschema:"description=搜索查询词"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	OriginalQuery  string   `json:"original_query"`
	RelatedQueries []string `json:"related_queries"`
}

// AIOverviewRequest AI Overview 请求
type AIOverviewRequest struct {
	Query string `json:"query" jsonschema:"description=要查询的内容"`
}

// AIOverviewResponse AI Overview 响应
type AIOverviewResponse struct {
	Query   string   `json:"query"`
	Summary string   `json:"summary"`
	Sources []string `json:"sources"`
}

// NewScraperToolAdapter 创建网页爬取工具适配器
func NewScraperToolAdapter(scraper tools.WebScraper) (tool.InvokableTool, error) {
	return utils.InferTool[ScrapeRequest, ScrapeResponse](
		"scrape_webpage",
		"爬取网页内容，提取标题和主要文本",
		func(ctx context.Context, req ScrapeRequest) (ScrapeResponse, error) {
			result, err := scraper.Scrape(ctx, req.URL)
			if err != nil {
				return ScrapeResponse{}, fmt.Errorf("爬取网页失败: %w", err)
			}

			return ScrapeResponse{
				URL:   result.URL,
				Title: result.Title,
				H1:    result.H1,
			}, nil
		},
	)
}

// NewSearcherToolAdapter 创建搜索工具适配器
func NewSearcherToolAdapter(searcher tools.Searcher) (tool.InvokableTool, error) {
	return utils.InferTool[SearchRequest, SearchResponse](
		"search_queries",
		"基于查询词搜索相关内容",
		func(ctx context.Context, req SearchRequest) (SearchResponse, error) {
			result, err := searcher.Search(ctx, req.Query)
			if err != nil {
				return SearchResponse{}, fmt.Errorf("搜索失败: %w", err)
			}

			return SearchResponse{
				OriginalQuery:  result.OriginalQuery,
				RelatedQueries: result.RelatedQueries,
			}, nil
		},
	)
}

// NewSERPToolAdapter 创建 AI Overview 获取工具适配器
func NewSERPToolAdapter(serp tools.SERPProvider) (tool.InvokableTool, error) {
	return utils.InferTool[AIOverviewRequest, AIOverviewResponse](
		"get_ai_overview",
		"获取生成式搜索引擎的 AI 摘要内容",
		func(ctx context.Context, req AIOverviewRequest) (AIOverviewResponse, error) {
			result, err := serp.GetAIOverview(ctx, req.Query)
			if err != nil {
				return AIOverviewResponse{}, fmt.Errorf("获取 AI Overview 失败: %w", err)
			}

			return AIOverviewResponse{
				Query:   result.Query,
				Summary: result.Summary,
				Sources: result.Sources,
			}, nil
		},
	)
}
