package models

import "time"

// ScrapedTitle 爬取的网页标题
type ScrapedTitle struct {
	Title string `json:"title"`
	H1    string `json:"h1"`
	URL   string `json:"url"`
}

// QueryFanout 查询发散结果
type QueryFanout struct {
	OriginalQuery  string    `json:"original_query"`
	RelatedQueries []string  `json:"related_queries"`
	Timestamp      time.Time `json:"timestamp"`
}

// MainQuery 主查询
type MainQuery struct {
	Query    string   `json:"query"`
	Keywords []string `json:"keywords"`
	Intent   string   `json:"intent"`
}

// AIOverview AI Overview 内容
type AIOverview struct {
	Query   string   `json:"query"`
	Summary string   `json:"summary"`
	Sources []string `json:"sources"`
	Snippet string   `json:"snippet"`
}

// QueryFanoutSummary 查询发散总结
type QueryFanoutSummary struct {
	OriginalQuery string   `json:"original_query"`
	Summary       string   `json:"summary"`
	KeyTopics     []string `json:"key_topics"`
}

// OptimizationSuggestion 优化建议
type OptimizationSuggestion struct {
	Priority   string `json:"priority"` // high, medium, low
	Category   string `json:"category"`
	Issue      string `json:"issue"`
	Suggestion string `json:"suggestion"`
}

// ComparisonItem 对比项
type ComparisonItem struct {
	Dimension   string `json:"dimension"`
	YourContent string `json:"your_content"`
	AIOverview  string `json:"ai_overview"`
	Similarity  string `json:"similarity"`
	Difference  string `json:"difference"`
}

// OptimizationReport 优化报告
type OptimizationReport struct {
	URL                     string                   `json:"url"`
	Title                   string                   `json:"title"`
	MainQuery               string                   `json:"main_query"`
	QueryFanoutSummary      string                   `json:"query_fanout_summary"`
	AIOverviewContent       string                   `json:"ai_overview_content"`
	ComparisonTable         []ComparisonItem         `json:"comparison_table"`
	ContentGaps             []string                 `json:"content_gaps"`
	OptimizationSuggestions []OptimizationSuggestion `json:"optimization_suggestions"`
	OverallScore            int                      `json:"overall_score"`
	Timestamp               time.Time                `json:"timestamp"`
}

// AnalysisResponse 分析响应
type AnalysisResponse struct {
	Success bool                `json:"success"`
	Data    *OptimizationReport `json:"data,omitempty"`
	Error   string              `json:"error,omitempty"`
}

// ProgressEvent 进度事件
type ProgressEvent struct {
	Step    int    `json:"step"`
	Total   int    `json:"total"`
	Agent   string `json:"agent"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Status  string `json:"status"` // progress, complete, error
}
