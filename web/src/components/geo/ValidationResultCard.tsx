import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { TrendingUp, TrendingDown, Minus, CheckCircle, Target, Lightbulb } from 'lucide-react'
import type { ValidationResult, PlatformType } from '@/lib/types'
import { PLATFORM_NAMES } from '@/lib/types'

interface ValidationResultCardProps {
  validationResult: ValidationResult
  platform: PlatformType
}

export function ValidationResultCard({ validationResult, platform }: ValidationResultCardProps) {
  const { original_score, optimized_score, improvement, comparison_table, suggestions } = validationResult

  const formatScore = (score: number) => Math.round(score)
  const formatPercentage = (pct: number) => {
    const sign = pct >= 0 ? '+' : ''
    return `${sign}${pct.toFixed(1)}%`
  }

  return (
    <Card className="relative overflow-hidden border-emerald-500/20">
      {/* 背景装饰 */}
      <div className="absolute inset-0 bg-gradient-to-br from-emerald-500/5 via-blue-500/5 to-purple-500/5" />
      <div className="absolute -top-20 -right-20 w-40 h-40 bg-emerald-500/10 rounded-full blur-3xl" />

      <CardHeader className="relative">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-emerald-500/10">
              <CheckCircle className="h-5 w-5 text-emerald-400" />
            </div>
            <div>
              <CardTitle className="text-lg">优化效果验证</CardTitle>
              <CardDescription>
                {PLATFORM_NAMES[platform]} 优化前后对比分析
              </CardDescription>
            </div>
          </div>

          {/* 总分提升徽章 */}
          <div className="text-right">
            <div className={`text-2xl font-bold ${improvement.percentage >= 0 ? 'text-emerald-400' : 'text-rose-400'}`}>
              {formatPercentage(improvement.percentage)}
            </div>
            <div className="text-xs text-muted-foreground">
              {formatScore(original_score.total)} → {formatScore(optimized_score.total)}
            </div>
          </div>
        </div>
      </CardHeader>

      <CardContent className="relative space-y-6">
        {/* 评分对比卡片 */}
        <div className="grid grid-cols-2 gap-4">
          {/* 原始评分 */}
          <div className="p-4 rounded-xl bg-card/50 border border-white/5">
            <div className="text-sm text-muted-foreground mb-3">优化前评分</div>
            <div className="text-3xl font-bold text-foreground/70 mb-3">
              {formatScore(original_score.total)}
            </div>
            <ScoreBreakdown score={original_score} />
          </div>

          {/* 优化后评分 */}
          <div className="p-4 rounded-xl bg-emerald-500/5 border border-emerald-500/20">
            <div className="text-sm text-emerald-400 mb-3">优化后评分</div>
            <div className="text-3xl font-bold text-emerald-400 mb-3">
              {formatScore(optimized_score.total)}
            </div>
            <ScoreBreakdown score={optimized_score} isOptimized />
          </div>
        </div>

        {/* 详细对比表格 */}
        <div className="space-y-3">
          <h4 className="text-sm font-medium flex items-center gap-2">
            <Target className="h-4 w-4 text-primary" />
            各维度详细对比
          </h4>

          <div className="space-y-2">
            {comparison_table.map((item, index) => (
              <ComparisonRow key={index} item={item} />
            ))}
          </div>
        </div>

        {/* 进一步建议 */}
        {suggestions && suggestions.length > 0 && (
          <div className="space-y-3">
            <h4 className="text-sm font-medium flex items-center gap-2">
              <Lightbulb className="h-4 w-4 text-amber-400" />
              进一步改进建议
            </h4>
            <div className="space-y-2">
              {suggestions.map((suggestion, index) => (
                <div
                  key={index}
                  className="flex items-start gap-3 p-3 rounded-xl bg-amber-500/5 border border-amber-500/10"
                >
                  <span className="flex-shrink-0 w-5 h-5 rounded-full bg-amber-500/10 flex items-center justify-center text-xs font-mono text-amber-400">
                    {index + 1}
                  </span>
                  <span className="text-sm text-foreground/90">{suggestion}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

// 评分细分
function ScoreBreakdown({ score, isOptimized = false }: { score: { authority: number; timeliness: number; structure: number; engagement: number; originality: number }; isOptimized?: boolean }) {
  const colorClass = isOptimized ? 'text-emerald-400' : 'text-foreground/60'

  return (
    <div className="space-y-1.5 text-xs">
      <div className={`flex justify-between ${colorClass}`}>
        <span>权威性</span>
        <span className="font-mono">{Math.round(score.authority)}</span>
      </div>
      <div className={`flex justify-between ${colorClass}`}>
        <span>时效性</span>
        <span className="font-mono">{Math.round(score.timeliness)}</span>
      </div>
      <div className={`flex justify-between ${colorClass}`}>
        <span>结构化</span>
        <span className="font-mono">{Math.round(score.structure)}</span>
      </div>
      <div className={`flex justify-between ${colorClass}`}>
        <span>互动性</span>
        <span className="font-mono">{Math.round(score.engagement)}</span>
      </div>
      <div className={`flex justify-between ${colorClass}`}>
        <span>原创度</span>
        <span className="font-mono">{Math.round(score.originality)}</span>
      </div>
    </div>
  )
}

// 对比行
function ComparisonRow({ item }: { item: { dimension: string; weight: number; original: number; optimized: number; diff: number; status: string; comment: string } }) {
  const getStatusIcon = () => {
    if (item.status === '提升') return <TrendingUp className="h-4 w-4 text-emerald-400" />
    if (item.status === '下降') return <TrendingDown className="h-4 w-4 text-rose-400" />
    return <Minus className="h-4 w-4 text-amber-400" />
  }

  const getStatusColor = () => {
    if (item.status === '提升') return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20'
    if (item.status === '下降') return 'text-rose-400 bg-rose-500/10 border-rose-500/20'
    return 'text-amber-400 bg-amber-500/10 border-amber-500/20'
  }

  const maxScore = Math.max(item.original, item.optimized, 100)
  const originalPercent = (item.original / maxScore) * 100
  const optimizedPercent = (item.optimized / maxScore) * 100

  return (
    <div className="p-3 rounded-xl bg-card/50 border border-white/5 space-y-2">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">{item.dimension}</span>
          <span className="text-xs text-muted-foreground">({item.weight}%)</span>
        </div>
        <Badge variant="outline" className={`text-xs ${getStatusColor()}`}>
          {getStatusIcon()}
          <span className="ml-1">{item.status} {item.diff > 0 ? '+' : ''}{item.diff}</span>
        </Badge>
      </div>

      {/* 进度条对比 */}
      <div className="space-y-1.5">
        <div className="flex items-center gap-2 text-xs">
          <span className="text-muted-foreground w-12">优化前</span>
          <div className="flex-1 h-1.5 bg-white/5 rounded-full overflow-hidden">
            <div
              className="h-full bg-foreground/30 rounded-full transition-all duration-500"
              style={{ width: `${originalPercent}%` }}
            />
          </div>
          <span className="text-muted-foreground w-8 text-right">{item.original}</span>
        </div>
        <div className="flex items-center gap-2 text-xs">
          <span className="text-emerald-400 w-12">优化后</span>
          <div className="flex-1 h-1.5 bg-white/5 rounded-full overflow-hidden">
            <div
              className="h-full bg-emerald-400 rounded-full transition-all duration-500"
              style={{ width: `${optimizedPercent}%` }}
            />
          </div>
          <span className="text-emerald-400 w-8 text-right">{item.optimized}</span>
        </div>
      </div>

      {/* 评语 */}
      {item.comment && (
        <p className="text-xs text-muted-foreground pl-14">{item.comment}</p>
      )}
    </div>
  )
}
