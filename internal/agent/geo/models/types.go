package models

// ScrapeTitleRequest 爬取标题请求
type ScrapeTitleRequest struct {
	URL string `json:"url" binding:"required,url"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query string `json:"query" binding:"required"`
}

// AnalysisRequest 分析请求
type AnalysisRequest struct {
	URL string `json:"url" binding:"required,url"`
}

// StreamAnalysisRequest 流式分析请求
type StreamAnalysisRequest struct {
	URL string `json:"url" binding:"required,url"`
}
