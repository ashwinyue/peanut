import React, { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { ExternalLink, Clock, RefreshCw, AlertCircle, CheckCircle, Loader2, Sparkles, FileText, Search, Lightbulb, Edit3, ChevronDown, ChevronUp, Copy, Check, Shield } from 'lucide-react'
import type { GEOAnalysisResponse, AnalysisStatus, OptimizationSuggestion, ProgressEventData, ValidationResult } from '@/lib/types'
import { PLATFORM_NAMES, PLATFORM_COLORS } from '@/lib/types'
import { useSSEProgress } from '@/hooks/useSSEProgress'
import { ScoreCard } from './ScoreCard'
import { SuggestionsCard } from './SuggestionsCard'
import { ValidationResultCard } from './ValidationResultCard'

interface AnalysisResultProps {
  analysis: GEOAnalysisResponse
  isLoading: boolean
  onRefresh: () => void
}

const statusConfig: Record<AnalysisStatus, { label: string; icon: React.ReactNode; color: string; bgColor: string; borderColor: string }> = {
  pending: {
    label: '等待处理',
    icon: <Clock className="h-4 w-4" />,
    color: 'text-slate-400',
    bgColor: 'bg-slate-500/10',
    borderColor: 'border-slate-500/20'
  },
  processing: {
    label: '正在分析',
    icon: <Loader2 className="h-4 w-4 animate-spin" />,
    color: 'text-blue-400',
    bgColor: 'bg-blue-500/10',
    borderColor: 'border-blue-500/20'
  },
  completed: {
    label: '分析完成',
    icon: <CheckCircle className="h-4 w-4" />,
    color: 'text-emerald-400',
    bgColor: 'bg-emerald-500/10',
    borderColor: 'border-emerald-500/20'
  },
  failed: {
    label: '分析失败',
    icon: <AlertCircle className="h-4 w-4" />,
    color: 'text-rose-400',
    bgColor: 'bg-rose-500/10',
    borderColor: 'border-rose-500/20'
  },
}

// 步骤名称映射
const stepNames: Record<string, { name: string; icon: typeof Search; description: string }> = {
  'title_scraper': { name: '爬取网页标题', icon: Search, description: '获取页面基础信息' },
  'query_fanout_researcher': { name: '搜索相关查询', icon: Search, description: '扩展相关搜索词' },
  'main_query_extractor': { name: '提取主查询', icon: Lightbulb, description: '识别核心关键词' },
  'ai_overview_retriever': { name: '获取 AI 摘要', icon: Sparkles, description: '抓取搜索引擎 AI 概述' },
  'query_fanout_summarizer': { name: '总结查询发散', icon: FileText, description: '整合搜索意图' },
  'ai_content_optimizer': { name: '生成优化报告', icon: Lightbulb, description: '分析优化机会' },
  'content_rewriter': { name: '生成优化文章', icon: Edit3, description: '重写优化内容' },
  'content_validator': { name: '验证优化效果', icon: Shield, description: '对比优化前后评分' },
}

export function AnalysisResult({ analysis, isLoading, onRefresh }: AnalysisResultProps) {
  const status = statusConfig[analysis.status] || statusConfig.pending
  const [expandedSections, setExpandedSections] = useState({
    aiOverview: true,
    optimizationReport: true,
    querySummary: false,
    contentGaps: true,
    optimizedArticle: false,
    validationResult: true,
  })
  const [copiedSection, setCopiedSection] = useState<string | null>(null)

  // 使用 SSE hook 监听进度
  const { progress, isConnected } = useSSEProgress(
    analysis.status === 'pending' || analysis.status === 'processing' ? analysis.id : null,
    undefined,
    () => {
      setTimeout(() => onRefresh(), 500)
    },
    (error) => {
      console.error('SSE error:', error)
      setTimeout(() => onRefresh(), 2000)
    }
  )

  const formatDate = (dateStr: string) => {
    try {
      return new Date(dateStr).toLocaleString('zh-CN')
    } catch {
      return dateStr
    }
  }

  // 解析 JSON 字符串
  const parseJsonArray = <T,>(jsonStr?: string): T[] => {
    if (!jsonStr) return []
    try {
      return JSON.parse(jsonStr)
    } catch {
      return []
    }
  }

  const suggestions = parseJsonArray<OptimizationSuggestion>(analysis.optimization_suggestions)

  // 解析验证结果
  const validationResult = parseJsonArray<ValidationResult>(analysis.validation_result)?.[0] || null

  // 获取步骤显示名称
  const getStepDisplayName = (agentName: string) => {
    return stepNames[agentName] || { name: agentName, icon: Search, description: '' }
  }

  // 复制到剪贴板
  const handleCopy = async (text: string, section: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopiedSection(section)
      setTimeout(() => setCopiedSection(null), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }

  // 切换章节展开/收起
  const toggleSection = (section: keyof typeof expandedSections) => {
    setExpandedSections(prev => ({ ...prev, [section]: !prev[section] }))
  }

  // 处理中状态 - 显示进度
  if (analysis.status === 'pending' || analysis.status === 'processing') {
    const progressPercent = progress ? Math.round((progress.step / progress.total) * 100) : 0
    const currentStep = progress ? getStepDisplayName(progress.agent_name) : null

    return (
      <div className="space-y-6">
        <Card className="relative overflow-hidden">
          {/* 背景动画 */}
          <div className="absolute inset-0 bg-gradient-to-br from-blue-500/5 via-purple-500/5 to-cyan-500/5" />
          <div className="absolute -top-20 -right-20 w-40 h-40 bg-blue-500/10 rounded-full blur-3xl animate-pulse" />

          <CardHeader className="relative">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className={`p-2 rounded-xl ${status.bgColor}`}>
                  <span className={status.color}>{status.icon}</span>
                </div>
                <div>
                  <CardTitle className="text-lg">{status.label}</CardTitle>
                  <CardDescription>正在分析您的网页内容</CardDescription>
                </div>
                {isConnected && (
                  <Badge variant="outline" className="ml-2 text-xs bg-emerald-500/10 border-emerald-500/30 text-emerald-400">
                    <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 mr-1.5 animate-pulse" />
                    实时更新
                  </Badge>
                )}
              </div>
              <Button variant="outline" size="sm" onClick={onRefresh} disabled={isLoading}>
                <RefreshCw className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
                刷新状态
              </Button>
            </div>
          </CardHeader>

          <CardContent className="relative space-y-6">
            {/* URL 信息 */}
            <div className="p-4 rounded-xl bg-card/50 border border-white/5">
              <p className="text-xs text-muted-foreground mb-2">分析 URL</p>
              <a
                href={analysis.url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 text-blue-400 hover:text-blue-300 transition-colors"
              >
                <ExternalLink className="h-4 w-4" />
                <span className="truncate">{analysis.url}</span>
              </a>
            </div>

            {/* 进度条 */}
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">总进度</span>
                <span className="text-sm font-mono text-muted-foreground">{progressPercent}%</span>
              </div>

              <div className="relative">
                <Progress value={progressPercent} className="h-2" />
                <div className="absolute inset-0 bg-gradient-to-r from-blue-500/20 via-purple-500/20 to-cyan-500/20 blur-lg" />
              </div>

              {/* 当前步骤详情 */}
              {currentStep && (
                <div className="flex items-start gap-4 p-4 rounded-xl bg-blue-500/5 border border-blue-500/20">
                  <div className="p-3 rounded-lg bg-blue-500/10">
                    <currentStep.icon className="h-5 w-5 text-blue-400" />
                  </div>
                  <div className="flex-1">
                    <p className="font-medium text-foreground">{currentStep.name}</p>
                    <p className="text-sm text-muted-foreground mt-1">{currentStep.description}</p>
                    {progress?.message && (
                      <p className="text-xs text-blue-400 mt-2 font-mono">{progress.message}</p>
                    )}
                  </div>
                  <div className="text-right">
                    <span className="text-sm font-mono text-muted-foreground">
                      {progress?.step}/{progress?.total}
                    </span>
                  </div>
                </div>
              )}
            </div>

            {/* 加载动画 */}
            <div className="flex items-center justify-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" />
              <span>{progress?.agent_name === 'content_rewriter' ? '正在生成优化内容...' : 'AI 正在分析中...'}</span>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  // 失败状态
  if (analysis.status === 'failed') {
    return (
      <Card className="relative overflow-hidden border-rose-500/30">
        <div className="absolute inset-0 bg-rose-500/5" />
        <CardHeader className="relative">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-rose-500/10">
              <AlertCircle className="h-5 w-5 text-rose-400" />
            </div>
            <div>
              <CardTitle className="text-lg">分析失败</CardTitle>
              <CardDescription>处理过程中发生错误</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="relative space-y-4">
          <div className="p-4 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-400">
            {analysis.error_message || '分析过程中发生未知错误'}
          </div>
          <div className="flex gap-2 text-sm text-muted-foreground">
            <Clock className="h-4 w-4" />
            <span>创建时间: {formatDate(analysis.created_at)}</span>
          </div>
        </CardContent>
      </Card>
    )
  }

  // 完成状态 - 显示完整报告
  return (
    <div className="space-y-6">
      {/* 头部信息卡片 */}
      <Card className="relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-emerald-500/5 via-blue-500/5 to-purple-500/5" />
        <div className="absolute -top-20 -right-20 w-40 h-40 bg-emerald-500/10 rounded-full blur-3xl" />

        <CardHeader className="relative">
          <div className="flex items-start justify-between">
            <div className="flex items-start gap-3">
              <div className={`p-2 rounded-xl ${status.bgColor}`}>
                <span className={status.color}>{status.icon}</span>
              </div>
              <div>
                <CardTitle className="text-lg">{analysis.title || '分析结果'}</CardTitle>
                <div className="flex items-center gap-2 mt-1">
                  <CardDescription className="text-xs">GEO 优化分析报告</CardDescription>
                  {analysis.platform && (
                    <Badge
                      variant="outline"
                      className={`text-xs bg-gradient-to-r ${PLATFORM_COLORS[analysis.platform]} bg-clip-text text-transparent border-white/20`}
                    >
                      <span className="bg-gradient-to-r from-blue-400 to-cyan-400 bg-clip-text text-transparent">
                        {PLATFORM_NAMES[analysis.platform]}
                      </span>
                    </Badge>
                  )}
                </div>
              </div>
            </div>
            <Button variant="outline" size="sm" onClick={onRefresh} disabled={isLoading}>
              <RefreshCw className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
              刷新
            </Button>
          </div>
        </CardHeader>

        <CardContent className="relative">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="p-3 rounded-xl bg-card/50 border border-white/5">
              <p className="text-xs text-muted-foreground mb-1">分析 URL</p>
              <a
                href={analysis.url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 text-blue-400 hover:text-blue-300 transition-colors text-sm truncate"
              >
                <ExternalLink className="h-3.5 w-3.5" />
                {analysis.url}
              </a>
            </div>
            <div className="p-3 rounded-xl bg-card/50 border border-white/5">
              <p className="text-xs text-muted-foreground mb-1">分析时间</p>
              <div className="flex items-center gap-2 text-sm">
                <Clock className="h-3.5 w-3.5 text-muted-foreground" />
                {formatDate(analysis.completed_at || analysis.updated_at)}
              </div>
            </div>
          </div>

          {analysis.main_query && (
            <div className="mt-4 flex items-center gap-2">
              <span className="text-xs text-muted-foreground">主查询:</span>
              <Badge variant="secondary" className="bg-blue-500/10 text-blue-400 border-blue-500/20">
                {analysis.main_query}
              </Badge>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 评分卡片 - 显示原始和优化后评分 */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <ScoreCard
          score={analysis.overall_score}
          title="优化前评分"
          description="原始内容 GEO 评分"
          variant="default"
        />
        <ScoreCard
          score={analysis.optimized_score || 0}
          title="优化后评分"
          description="重写后内容 GEO 评分"
          variant="improved"
        />
      </div>

      {/* AI Overview 内容 */}
      {analysis.ai_overview && (
        <CollapsibleSection
          title="AI Overview"
          description="搜索引擎 AI 生成的概述内容"
          icon={Sparkles}
          isExpanded={expandedSections.aiOverview}
          onToggle={() => toggleSection('aiOverview')}
          onCopy={() => handleCopy(analysis.ai_overview!, 'aiOverview')}
          isCopied={copiedSection === 'aiOverview'}
        >
          <div className="p-4 rounded-xl bg-blue-500/5 border border-blue-500/10">
            <p className="text-sm whitespace-pre-wrap leading-relaxed text-foreground/90">
              {analysis.ai_overview}
            </p>
          </div>
        </CollapsibleSection>
      )}

      {/* 优化报告 */}
      {analysis.optimization_report && (
        <CollapsibleSection
          title="GEO 优化报告"
          description="详细的搜索引擎优化分析报告"
          icon={FileText}
          isExpanded={expandedSections.optimizationReport}
          onToggle={() => toggleSection('optimizationReport')}
          onCopy={() => handleCopy(analysis.optimization_report!, 'optimizationReport')}
          isCopied={copiedSection === 'optimizationReport'}
        >
          <div className="p-4 rounded-xl bg-purple-500/5 border border-purple-500/10 max-h-96 overflow-y-auto">
            <div className="text-sm whitespace-pre-wrap leading-relaxed text-foreground/90">
              {analysis.optimization_report}
            </div>
          </div>
        </CollapsibleSection>
      )}

      {/* 查询发散总结 */}
      {analysis.query_fanout_summary && (
        <CollapsibleSection
          title="查询发散总结"
          description="相关搜索查询的分析总结"
          icon={Search}
          isExpanded={expandedSections.querySummary}
          onToggle={() => toggleSection('querySummary')}
        >
          <p className="text-sm text-muted-foreground leading-relaxed p-4 rounded-xl bg-card/50 border border-white/5">
            {analysis.query_fanout_summary}
          </p>
        </CollapsibleSection>
      )}

      {/* 内容差距 */}
      {analysis.content_gaps && (
        <CollapsibleSection
          title="内容差距"
          description="您的内容中缺失或需要补充的部分"
          icon={AlertCircle}
          isExpanded={expandedSections.contentGaps}
          onToggle={() => toggleSection('contentGaps')}
        >
          <div className="grid gap-2">
            {parseJsonArray<string>(analysis.content_gaps).map((gap, index) => (
              <div
                key={index}
                className="flex items-start gap-3 p-3 rounded-xl bg-amber-500/5 border border-amber-500/10"
              >
                <span className="flex-shrink-0 w-6 h-6 rounded-full bg-amber-500/10 flex items-center justify-center text-xs font-mono text-amber-400">
                  {index + 1}
                </span>
                <span className="text-sm text-foreground/90">{gap}</span>
              </div>
            ))}
          </div>
        </CollapsibleSection>
      )}

      {/* 优化建议 */}
      {suggestions.length > 0 && <SuggestionsCard suggestions={suggestions} />}

      {/* 优化后的文章 */}
      {analysis.optimized_article && (
        <CollapsibleSection
          title="AI 重写优化文章"
          description="基于 GEO 分析结果重新生成的优化版本"
          icon={Edit3}
          isExpanded={expandedSections.optimizedArticle}
          onToggle={() => toggleSection('optimizedArticle')}
          onCopy={() => handleCopy(analysis.optimized_article!, 'optimizedArticle')}
          isCopied={copiedSection === 'optimizedArticle'}
          defaultExpanded={false}
        >
          <div className="p-4 rounded-xl bg-emerald-500/5 border border-emerald-500/10 max-h-[600px] overflow-y-auto">
            <div className="text-sm whitespace-pre-wrap leading-relaxed text-foreground/90">
              {analysis.optimized_article}
            </div>
          </div>
        </CollapsibleSection>
      )}

      {/* 验证结果 */}
      {validationResult && (
        <ValidationResultCard
          validationResult={validationResult}
          platform={analysis.platform || 'doubao'}
        />
      )}
    </div>
  )
}

// 可折叠章节组件
interface CollapsibleSectionProps {
  title: string
  description: string
  icon: typeof FileText
  isExpanded: boolean
  onToggle: () => void
  onCopy?: () => void
  isCopied?: boolean
  children: React.ReactNode
  defaultExpanded?: boolean
}

function CollapsibleSection({
  title,
  description,
  icon: Icon,
  isExpanded,
  onToggle,
  onCopy,
  isCopied,
  children
}: CollapsibleSectionProps) {
  return (
    <Card className="relative overflow-hidden">
      <CardHeader className="relative">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-primary/10">
              <Icon className="h-5 w-5 text-primary" />
            </div>
            <div>
              <CardTitle className="text-lg">{title}</CardTitle>
              <CardDescription>{description}</CardDescription>
            </div>
          </div>
          <div className="flex items-center gap-2">
            {onCopy && (
              <Button
                variant="ghost"
                size="sm"
                onClick={onCopy}
                className="h-8 px-2 text-muted-foreground hover:text-foreground"
              >
                {isCopied ? (
                  <Check className="h-4 w-4 text-emerald-400" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
              </Button>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={onToggle}
              className="h-8 px-2"
            >
              {isExpanded ? (
                <ChevronUp className="h-4 w-4" />
              ) : (
                <ChevronDown className="h-4 w-4" />
              )}
            </Button>
          </div>
        </div>
      </CardHeader>

      <div className={`overflow-hidden transition-all duration-500 ${isExpanded ? 'max-h-[2000px]' : 'max-h-0'}`}>
        <CardContent>
          {children}
        </CardContent>
      </div>
    </Card>
  )
}
