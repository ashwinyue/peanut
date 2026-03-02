package tools

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/solariswu/peanut/internal/agent/geo/models"
)

// DuckDuckGoSearcher 使用 DuckDuckGo 搜索引擎（免费开源）
type DuckDuckGoSearcher struct {
	client duckduckgo.Search
}

// NewDuckDuckGoSearcher 创建 DuckDuckGo 搜索器
func NewDuckDuckGoSearcher() (*DuckDuckGoSearcher, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := duckduckgo.NewSearch(ctx, &duckduckgo.Config{
		Timeout:    30 * time.Second,
		MaxResults: 10,
		Region:     duckduckgo.RegionCN, // 中国地区
	})
	if err != nil {
		return nil, fmt.Errorf("创建 DuckDuckGo 客户端失败: %w", err)
	}

	return &DuckDuckGoSearcher{
		client: client,
	}, nil
}

// Search 执行搜索
func (s *DuckDuckGoSearcher) Search(ctx context.Context, query string) (*models.QueryFanout, error) {
	if query == "" {
		return nil, fmt.Errorf("搜索查询不能为空")
	}

	// 调用 DuckDuckGo 文本搜索
	request := &duckduckgo.TextSearchRequest{
		Query: query,
	}

	response, err := s.client.TextSearch(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("DuckDuckGo 搜索失败: %w", err)
	}

	// 转换响应为 QueryFanout
	relatedQueries := make([]string, 0, len(response.Results))
	for _, result := range response.Results {
		// 将搜索结果作为相关查询
		relatedQueries = append(relatedQueries, result.Title)
		
		// 如果需要更多查询，可以基于结果生成
		if len(relatedQueries) >= 10 {
			break
		}
	}

	// 如果没有找到足够的查询，添加一些基于原始查询的变体
	if len(relatedQueries) < 5 {
		variations := []string{
			query + " 指南",
			query + " 教程",
			query + " 最佳实践",
			"如何 " + query,
			query + " 方法",
		}
		for _, v := range variations {
			if len(relatedQueries) >= 10 {
				break
			}
			relatedQueries = append(relatedQueries, v)
		}
	}

	return &models.QueryFanout{
		OriginalQuery:  query,
		RelatedQueries: relatedQueries,
		Timestamp:      time.Now(),
	}, nil
}

// DuckDuckGoSERPProvider 使用 DuckDuckGo 作为 SERP 提供者
type DuckDuckGoSERPProvider struct {
	client duckduckgo.Search
}

// NewDuckDuckGoSERPProvider 创建 DuckDuckGo SERP 提供者
func NewDuckDuckGoSERPProvider() (*DuckDuckGoSERPProvider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := duckduckgo.NewSearch(ctx, &duckduckgo.Config{
		Timeout:    30 * time.Second,
		MaxResults: 10,
		Region:     duckduckgo.RegionCN,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("创建 DuckDuckGo 客户端失败: %w", err)
	}

	return &DuckDuckGoSERPProvider{
		client: client,
	}, nil
}

// GetAIOverview 获取 AI Overview（使用 DuckDuckGo 搜索结果模拟）
func (p *DuckDuckGoSERPProvider) GetAIOverview(ctx context.Context, query string) (*models.AIOverview, error) {
	if query == "" {
		return nil, fmt.Errorf("搜索查询不能为空")
	}

	// 调用 DuckDuckGo 文本搜索
	request := &duckduckgo.TextSearchRequest{
		Query: query,
	}

	response, err := p.client.TextSearch(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("DuckDuckGo 搜索失败: %w", err)
	}

	// 从搜索结果中提取信息，模拟 AI Overview
	var summary string
	var sources []string

	if len(response.Results) > 0 {
		// 从前几个结果生成摘要
		topResults := response.Results
		if len(topResults) > 5 {
			topResults = topResults[:5]
		}

		summary = fmt.Sprintf("基于 %s 的搜索结果，我们找到了以下关键信息：\n\n", query)
		
		for i, result := range topResults {
			summary += fmt.Sprintf("%d. %s\n", i+1, result.Title)
			if result.Summary != "" {
				summary += fmt.Sprintf("   %s\n", result.Summary)
			}
			sources = append(sources, result.URL)
		}

		// 生成综合摘要
		if len(topResults) > 0 {
			summary += fmt.Sprintf("\n综合摘要：关于「%s」，搜索结果显示了多个相关资源。", query)
		}
	}

	// 如果没有结果，提供默认消息
	if summary == "" {
		summary = fmt.Sprintf("关于「%s」的搜索未找到直接结果，建议尝试不同的关键词。", query)
	}

	// 提取关键点作为 snippet
	snippet := ""
	if len(response.Results) > 0 && response.Results[0].Summary != "" {
		snippet = truncateString(response.Results[0].Summary, 200)
	} else {
		snippet = fmt.Sprintf("搜索「%s」相关的信息", query)
	}

	return &models.AIOverview{
		Query:   query,
		Summary: summary,
		Sources: sources,
		Snippet: snippet,
	}, nil
}

// truncateString 截断字符串到指定长度
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	
	// 尝试在单词边界截断
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return s
}
