import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Eye, Trash2, ExternalLink, Clock, Search, BarChart3, ArrowRight, Layers } from 'lucide-react'
import type { GEOAnalysisResponse, AnalysisStatus } from '@/lib/types'
import { PLATFORM_NAMES, PLATFORM_COLORS } from '@/lib/types'

interface AnalysisListProps {
  list: GEOAnalysisResponse[]
  isLoading: boolean
  onView: (id: number) => void
  onDelete: (id: number) => void
}

const statusConfig: Record<AnalysisStatus, {
  label: string
  variant: 'default' | 'secondary' | 'outline' | 'destructive'
  icon: typeof Clock
  color: string
  bgColor: string
}> = {
  pending: {
    label: '等待中',
    variant: 'secondary',
    icon: Clock,
    color: 'text-slate-400',
    bgColor: 'bg-slate-500/10',
  },
  processing: {
    label: '分析中',
    variant: 'default',
    icon: BarChart3,
    color: 'text-blue-400',
    bgColor: 'bg-blue-500/10',
  },
  completed: {
    label: '已完成',
    variant: 'outline',
    icon: BarChart3,
    color: 'text-emerald-400',
    bgColor: 'bg-emerald-500/10',
  },
  failed: {
    label: '失败',
    variant: 'destructive',
    icon: Clock,
    color: 'text-rose-400',
    bgColor: 'bg-rose-500/10',
  },
}

export function AnalysisList({ list, isLoading, onView, onDelete }: AnalysisListProps) {
  const formatDate = (dateStr: string) => {
    try {
      const date = new Date(dateStr)
      const now = new Date()
      const diff = now.getTime() - date.getTime()
      const days = Math.floor(diff / (1000 * 60 * 60 * 24))
      const hours = Math.floor(diff / (1000 * 60 * 60))
      const minutes = Math.floor(diff / (1000 * 60))

      if (minutes < 1) return '刚刚'
      if (minutes < 60) return `${minutes} 分钟前`
      if (hours < 24) return `${hours} 小时前`
      if (days < 7) return `${days} 天前`

      return date.toLocaleString('zh-CN', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
    } catch {
      return dateStr
    }
  }

  if (isLoading && list.length === 0) {
    return (
      <Card className="glass-card">
        <CardContent className="py-16 text-center">
          <div className="relative inline-flex mb-4">
            <div className="w-12 h-12 rounded-full border-2 border-blue-500/30 border-t-blue-500 animate-spin" />
          </div>
          <p className="text-muted-foreground">加载中...</p>
        </CardContent>
      </Card>
    )
  }

  if (list.length === 0) {
    return (
      <Card className="glass-card">
        <CardContent className="py-16 text-center">
          <div className="relative inline-flex mb-6">
            <div className="w-20 h-20 rounded-2xl bg-gradient-to-br from-blue-500/10 to-purple-500/10 flex items-center justify-center">
              <Search className="h-8 w-8 text-muted-foreground/50" />
            </div>
            <div className="absolute -bottom-1 -right-1 w-8 h-8 rounded-full bg-card border border-white/10 flex items-center justify-center">
              <span className="text-lg">🔍</span>
            </div>
          </div>
          <h3 className="text-lg font-medium mb-2">暂无分析记录</h3>
          <p className="text-sm text-muted-foreground max-w-sm mx-auto">
            创建一个新的分析任务，开始优化您的网页内容 SEO 表现
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-4">
      {list.map((item, index) => {
        const status = statusConfig[item.status] || statusConfig.pending
        const Icon = status.icon
        const hasScore = item.overall_score > 0 && item.status === 'completed'

        return (
          <Card
            key={item.id}
            className="group relative overflow-hidden glass-card hover:border-white/20 transition-all duration-500"
            style={{ animationDelay: `${index * 50}ms` }}
          >
            {/* 背景光效 */}
            <div className={`absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-500 ${status.bgColor}`} />

            <CardContent className="relative py-5">
              <div className="flex items-start gap-4">
                {/* 状态指示器 */}
                <div className={`flex-shrink-0 w-12 h-12 rounded-xl ${status.bgColor} flex items-center justify-center`}>
                  <Icon className={`h-5 w-5 ${status.color} ${item.status === 'processing' ? 'animate-pulse' : ''}`} />
                </div>

                {/* 内容区域 */}
                <div className="flex-1 min-w-0">
                  <div className="flex flex-wrap items-center gap-2 mb-2">
                    <Badge
                      variant={status.variant}
                      className={`text-xs ${item.status === 'processing' ? 'animate-pulse' : ''}`}
                    >
                      {status.label}
                    </Badge>

                    {/* 平台标签 */}
                    {item.platform && (
                      <Badge
                        variant="outline"
                        className={`text-xs gap-1 border-white/10 bg-gradient-to-r ${PLATFORM_COLORS[item.platform]} bg-clip-text text-transparent`}
                      >
                        <Layers className="h-3 w-3 text-muted-foreground" />
                        <span className="text-muted-foreground">{PLATFORM_NAMES[item.platform]}</span>
                      </Badge>
                    )}

                    {hasScore && (
                      <>
                        <Badge variant="outline" className="text-xs gap-1">
                          <BarChart3 className="h-3 w-3" />
                          原始: {item.overall_score}
                        </Badge>
                        {item.optimized_score > 0 && (
                          <Badge variant="outline" className="text-xs gap-1 border-emerald-500/20 text-emerald-400">
                            <ArrowRight className="h-3 w-3" />
                            优化后: {item.optimized_score}
                          </Badge>
                        )}
                      </>
                    )}

                    <span className="text-xs text-muted-foreground flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      {formatDate(item.created_at)}
                    </span>
                  </div>

                  <h4 className="font-medium text-foreground truncate mb-1 group-hover:text-blue-400 transition-colors">
                    {item.title || '无标题'}
                  </h4>

                  <a
                    href={item.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm text-muted-foreground hover:text-blue-400 flex items-center gap-1 truncate transition-colors group/link"
                  >
                    <span className="truncate">{item.url}</span>
                    <ExternalLink className="h-3 w-3 flex-shrink-0 opacity-0 group-hover/link:opacity-100 transition-opacity" />
                  </a>

                  {item.main_query && (
                    <div className="mt-2 flex items-center gap-2">
                      <span className="text-xs text-muted-foreground">主查询:</span>
                      <span className="text-xs px-2 py-0.5 rounded-full bg-blue-500/10 text-blue-400 border border-blue-500/20">
                        {item.main_query}
                      </span>
                    </div>
                  )}
                </div>

                {/* 操作按钮 */}
                <div className="flex flex-col gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onView(item.id)}
                    className="opacity-0 group-hover:opacity-100 translate-x-2 group-hover:translate-x-0 transition-all duration-300"
                  >
                    <Eye className="h-4 w-4 mr-1" />
                    查看
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onDelete(item.id)}
                    className="opacity-0 group-hover:opacity-100 translate-x-2 group-hover:translate-x-0 transition-all duration-300 delay-75 text-rose-400 hover:text-rose-400 hover:bg-rose-500/10"
                  >
                    <Trash2 className="h-4 w-4 mr-1" />
                    删除
                  </Button>
                </div>
              </div>

              {/* 悬停箭头 */}
              <div className="absolute right-4 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 transition-all duration-300 translate-x-4 group-hover:translate-x-0">
                <ArrowRight className="h-5 w-5 text-blue-400" />
              </div>
            </CardContent>
          </Card>
        )
      })}
    </div>
  )
}
