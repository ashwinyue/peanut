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

// 分析状态
export type AnalysisStatus = 'pending' | 'processing' | 'completed' | 'failed'

// 平台类型 - AI 搜索引擎
export type PlatformType = 'google'

// 平台配置
export interface PlatformConfig {
  type: PlatformType
  name: string
  description: string
  weight: {
    authority: number
    timeliness: number
    structure: number
    engagement: number
    originality: number
  }
}

// 验证结果 - 评分详情
export interface ScoreDetail {
  authority: number
  timeliness: number
  structure: number
  engagement: number
  originality: number
  total: number
}

// 验证结果 - 提升情况
export interface Improvement {
  total_diff: number
  percentage: number
  authority_diff: number
  timeliness_diff: number
  structure_diff: number
  engagement_diff: number
  originality_diff: number
}

// 验证结果 - 单项对比
export interface ComparisonItem {
  dimension: string
  weight: number
  original: number
  optimized: number
  diff: number
  status: '提升' | '持平' | '下降'
  comment: string
}

// 验证结果
export interface ValidationResult {
  original_score: ScoreDetail
  optimized_score: ScoreDetail
  improvement: Improvement
  comparison_table: ComparisonItem[]
  suggestions: string[]
}

// GEO 分析任务响应
export interface GEOAnalysisResponse {
  id: number
  url: string
  title: string
  main_query: string
  platform: PlatformType
  overall_score: number
  optimized_score: number
  status: AnalysisStatus
  error_message?: string
  query_fanout?: string
  ai_overview?: string
  query_fanout_summary?: string
  optimization_report?: string
  optimized_article?: string
  content_gaps?: string // JSON 字符串
  optimization_suggestions?: string // JSON 字符串
  validation_result?: string // JSON 字符串
  created_at: string
  updated_at: string
  completed_at?: string
}

// 创建分析任务请求
export interface GEOAnalysisCreateRequest {
  url: string
  platform?: PlatformType
}

// 列表查询请求
export interface GEOAnalysisListRequest {
  page?: number
  page_size?: number
  status?: string
  order_by?: string
  order_desc?: boolean
}

// ============ 进度相关类型 ============

// SSE 进度事件（后端 progress.Manager 推送）
export interface ProgressEventData {
  analysis_id: number
  step: number
  total: number
  agent_name: string  // 后端使用 agent_name
  message: string
  status: 'processing' | 'completed' | 'failed'
  score?: number // 完成时返回
}

// ============ 报告展示相关类型 ============

// 对比项
export interface ComparisonItemOld {
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

// 平台显示名称映射 - AI 搜索引擎
export const PLATFORM_NAMES: Record<PlatformType, string> = {
  google: 'Google AI Overview',
}

// 平台颜色映射 - AI 搜索引擎
export const PLATFORM_COLORS: Record<PlatformType, string> = {
  google: 'from-blue-500 to-cyan-500',
}
