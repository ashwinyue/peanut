// API 响应基础结构
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
}

// 分页数据结构
export interface PageData<T = unknown> {
  list: T[]
  total: number
  page: number
  page_size: number
}

// GEO 分析请求
export interface AnalysisRequest {
  url: string
}

// 对比项
export interface ComparisonItem {
  dimension: string
  your_content: string
  ai_overview: string
  similarity: string
  difference: string
}

// 优化建议
export interface OptimizationSuggestion {
  priority: 'high' | 'medium' | 'low'
  category: string
  issue: string
  suggestion: string
}

// 优化报告
export interface OptimizationReport {
  url: string
  title: string
  main_query: string
  query_fanout_summary: string
  ai_overview_content: string
  comparison_table: ComparisonItem[]
  content_gaps: string[]
  optimization_suggestions: OptimizationSuggestion[]
  overall_score: number
  timestamp: string
}

// SSE 进度事件
export interface ProgressEvent {
  step: number
  total: number
  agent: string
  message?: string
  error?: string
  status: 'progress' | 'complete' | 'error'
}

// SSE 完成事件
export interface CompleteEvent {
  status: string
  score: number
}

// SSE 错误事件
export interface ErrorEvent {
  error: string
}
