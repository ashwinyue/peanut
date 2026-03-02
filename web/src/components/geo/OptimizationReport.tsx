import type { OptimizationReport as OptimizationReportType } from '@/lib/types'
import { ScoreCard } from './ScoreCard'
import { ComparisonTable } from './ComparisonTable'
import { SuggestionsCard } from './SuggestionsCard'
import { ContentGaps } from './ContentGaps'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { ExternalLink, Clock } from 'lucide-react'

interface OptimizationReportProps {
  report: OptimizationReportType
}

export function OptimizationReport({ report }: OptimizationReportProps) {
  const formatDate = (timestamp: string) => {
    try {
      return new Date(timestamp).toLocaleString('zh-CN')
    } catch {
      return timestamp
    }
  }

  return (
    <div className="space-y-6">
      {/* 基本信息卡片 */}
      <Card>
        <CardHeader>
          <div className="flex items-start justify-between">
            <div>
              <CardTitle className="text-lg">分析结果</CardTitle>
              <CardDescription>网页内容 GEO 优化分析报告</CardDescription>
            </div>
            {report.timestamp && (
              <div className="flex items-center gap-1 text-sm text-muted-foreground">
                <Clock className="h-4 w-4" />
                {formatDate(report.timestamp)}
              </div>
            )}
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">分析 URL</p>
              <a
                href={report.url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-1 text-primary hover:underline truncate"
              >
                {report.url}
                <ExternalLink className="h-3 w-3 shrink-0" />
              </a>
            </div>
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">页面标题</p>
              <p className="font-medium truncate">{report.title || '无标题'}</p>
            </div>
          </div>
          {report.main_query && (
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">主查询</p>
              <Badge variant="secondary">{report.main_query}</Badge>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 评分和总结 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ScoreCard score={report.overall_score} />
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">查询发散总结</CardTitle>
            <CardDescription>相关搜索查询的分析总结</CardDescription>
          </CardHeader>
          <CardContent>
            {report.query_fanout_summary ? (
              <p className="text-sm text-muted-foreground leading-relaxed">
                {report.query_fanout_summary}
              </p>
            ) : (
              <p className="text-sm text-muted-foreground italic">暂无数据</p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* AI Overview 内容 */}
      {report.ai_overview_content && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">AI Overview 内容</CardTitle>
            <CardDescription>搜索引擎 AI 生成的概述内容</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="p-4 rounded-lg bg-muted/50 border">
              <p className="text-sm whitespace-pre-wrap leading-relaxed">
                {report.ai_overview_content}
              </p>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 对比表格 */}
      {report.comparison_table && report.comparison_table.length > 0 && (
        <ComparisonTable items={report.comparison_table} />
      )}

      {/* 内容差距 */}
      {report.content_gaps && report.content_gaps.length > 0 && (
        <ContentGaps gaps={report.content_gaps} />
      )}

      {/* 优化建议 */}
      {report.optimization_suggestions && report.optimization_suggestions.length > 0 && (
        <SuggestionsCard suggestions={report.optimization_suggestions} />
      )}
    </div>
  )
}
