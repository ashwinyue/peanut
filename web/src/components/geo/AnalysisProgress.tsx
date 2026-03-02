import { Progress } from '@/components/ui/progress'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Loader2, CheckCircle, AlertCircle } from 'lucide-react'
import type { ProgressEvent } from '@/lib/types'

interface AnalysisProgressProps {
  progress: ProgressEvent | null
}

// Agent 名称中英文映射
const agentNameMap: Record<string, string> = {
  'Title Scraper': '标题抓取',
  'Main Query Extractor': '主查询提取',
  'Query Fanout Researcher': '查询发散研究',
  'Query Fanout Summarizer': '查询发散总结',
  'AI Overview Retriever': 'AI 概述获取',
  'AI Content Optimizer': 'AI 内容优化',
  'Content Analyzer': '内容分析',
  'Report Generator': '报告生成',
}

// 翻译 agent 名称
const translateAgentName = (agentName: string): string => {
  if (!agentName) return ''
  // 直接匹配
  if (agentNameMap[agentName]) {
    return agentNameMap[agentName]
  }
  // 模糊匹配
  const lowerName = agentName.toLowerCase()
  for (const [key, value] of Object.entries(agentNameMap)) {
    if (key.toLowerCase().includes(lowerName) || lowerName.includes(key.toLowerCase())) {
      return value
    }
  }
  return agentName
}

export function AnalysisProgress({ progress }: AnalysisProgressProps) {
  if (!progress) return null

  const percentage = progress.total > 0
    ? Math.round((progress.step / progress.total) * 100)
    : 0

  const getStatusIcon = () => {
    switch (progress.status) {
      case 'complete':
        return <CheckCircle className="h-4 w-4 text-green-500" />
      case 'error':
        return <AlertCircle className="h-4 w-4 text-destructive" />
      default:
        return <Loader2 className="h-4 w-4 animate-spin text-primary" />
    }
  }

  const getStatusBadge = () => {
    switch (progress.status) {
      case 'complete':
        return <Badge variant="success">完成</Badge>
      case 'error':
        return <Badge variant="destructive">错误</Badge>
      default:
        return <Badge variant="secondary">进行中</Badge>
    }
  }

  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg flex items-center gap-2">
            {getStatusIcon()}
            分析进度
          </CardTitle>
          {getStatusBadge()}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <Progress value={percentage} className="h-2" />
          <div className="flex justify-between text-sm text-muted-foreground">
            <span>步骤 {progress.step} / {progress.total}</span>
            <span>{percentage}%</span>
          </div>
        </div>

        {progress.agent && (
          <div className="space-y-1">
            <p className="text-sm font-medium text-muted-foreground">当前步骤</p>
            <p className="text-sm font-medium">{translateAgentName(progress.agent)}</p>
          </div>
        )}

        {progress.message && (
          <div className="space-y-1">
            <p className="text-sm font-medium text-muted-foreground">状态</p>
            <p className="text-sm">{progress.message}</p>
          </div>
        )}

        {progress.error && (
          <div className="p-3 rounded-lg bg-destructive/10 text-destructive text-sm">
            {progress.error}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
