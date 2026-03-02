import { useState } from 'react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import type { OptimizationSuggestion } from '@/lib/types'
import {
  AlertCircle,
  AlertTriangle,
  Info,
  ChevronDown,
  Lightbulb,
  Target,
  CheckCircle2
} from 'lucide-react'

interface SuggestionsCardProps {
  suggestions: OptimizationSuggestion[]
}

export function SuggestionsCard({ suggestions }: SuggestionsCardProps) {
  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>({
    high: true,
    medium: true,
    low: true,
  })

  if (!suggestions || suggestions.length === 0) return null

  const groupedSuggestions = {
    high: suggestions.filter((s) => s.priority.toLowerCase() === 'high'),
    medium: suggestions.filter((s) => s.priority.toLowerCase() === 'medium'),
    low: suggestions.filter((s) => s.priority.toLowerCase() === 'low'),
  }

  const priorityConfig = {
    high: {
      icon: AlertCircle,
      color: 'text-rose-400',
      bgColor: 'bg-rose-500/10',
      borderColor: 'border-rose-500/20',
      badgeVariant: 'destructive' as const,
      label: '高优先级',
      description: '建议优先处理这些关键问题',
      glowColor: 'shadow-rose-500/10',
    },
    medium: {
      icon: AlertTriangle,
      color: 'text-amber-400',
      bgColor: 'bg-amber-500/10',
      borderColor: 'border-amber-500/20',
      badgeVariant: 'default' as const,
      label: '中优先级',
      description: '这些问题会显著影响 SEO 表现',
      glowColor: 'shadow-amber-500/10',
    },
    low: {
      icon: Info,
      color: 'text-blue-400',
      bgColor: 'bg-blue-500/10',
      borderColor: 'border-blue-500/20',
      badgeVariant: 'secondary' as const,
      label: '低优先级',
      description: '可选的优化建议',
      glowColor: 'shadow-blue-500/10',
    },
  }

  const toggleGroup = (priority: 'high' | 'medium' | 'low') => {
    setExpandedGroups(prev => ({
      ...prev,
      [priority]: !prev[priority]
    }))
  }

  const renderSuggestionGroup = (
    priority: 'high' | 'medium' | 'low',
    items: OptimizationSuggestion[]
  ) => {
    if (items.length === 0) return null

    const config = priorityConfig[priority]
    const Icon = config.icon
    const isExpanded = expandedGroups[priority]

    return (
      <div className={`rounded-xl border ${config.borderColor} overflow-hidden transition-all duration-300`}>
        {/* 分组标题栏 */}
        <button
          onClick={() => toggleGroup(priority)}
          className={`w-full flex items-center justify-between p-4 ${config.bgColor} hover:brightness-110 transition-all`}
        >
          <div className="flex items-center gap-3">
            <div className={`p-2 rounded-lg bg-card/50 ${config.color}`}>
              <Icon className="h-4 w-4" />
            </div>
            <div className="text-left">
              <div className="flex items-center gap-2">
                <span className="font-medium text-foreground">{config.label}</span>
                <Badge variant={config.badgeVariant} className="text-xs">
                  {items.length} 项
                </Badge>
              </div>
              <p className="text-xs text-muted-foreground mt-0.5">{config.description}</p>
            </div>
          </div>
          <div className={`p-1 rounded-md transition-transform duration-300 ${isExpanded ? 'rotate-180' : ''}`}>
            <ChevronDown className="h-4 w-4 text-muted-foreground" />
          </div>
        </button>

        {/* 建议列表 */}
        <div className={`overflow-hidden transition-all duration-500 ${isExpanded ? 'max-h-[2000px]' : 'max-h-0'}`}>
          <div className="p-4 space-y-3">
            {items.map((item, index) => (
              <SuggestionItem
                key={index}
                item={item}
                index={index}
                color={config.color}
              />
            ))}
          </div>
        </div>
      </div>
    )
  }

  return (
    <Card className="relative overflow-hidden">
      {/* 背景装饰 */}
      <div className="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-blue-500/5 to-cyan-500/5 rounded-full blur-3xl" />

      <CardHeader className="relative">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-lg flex items-center gap-2">
              <Lightbulb className="h-5 w-5 text-amber-400" />
              优化建议
            </CardTitle>
            <CardDescription>根据分析结果提供的改进建议</CardDescription>
          </div>
          <div className="hidden sm:flex items-center gap-2 text-xs text-muted-foreground">
            <CheckCircle2 className="h-4 w-4 text-emerald-400" />
            <span>共 {suggestions.length} 条建议</span>
          </div>
        </div>
      </CardHeader>

      <CardContent className="relative space-y-4">
        {renderSuggestionGroup('high', groupedSuggestions.high)}
        {renderSuggestionGroup('medium', groupedSuggestions.medium)}
        {renderSuggestionGroup('low', groupedSuggestions.low)}
      </CardContent>
    </Card>
  )
}

// 单个建议项组件
function SuggestionItem({
  item,
  index,
  color
}: {
  item: OptimizationSuggestion
  index: number
  color: string
}) {
  return (
    <div
      className="group relative p-4 rounded-lg bg-card/50 border border-white/5 hover:border-white/10 hover:bg-card/80 transition-all duration-300"
      style={{ animationDelay: `${index * 50}ms` }}
    >
      {/* 序号标记 */}
      <div className="absolute -left-2 -top-2 w-6 h-6 rounded-full bg-card border border-white/10 flex items-center justify-center text-xs font-mono text-muted-foreground">
        {index + 1}
      </div>

      <div className="space-y-3">
        {/* 类别标签 */}
        <div className="flex items-center gap-2">
          <Badge variant="outline" className="text-xs bg-cyan-500/5 border-cyan-500/20 text-cyan-400">
            <Target className="h-3 w-3 mr-1" />
            {item.category}
          </Badge>
        </div>

        {/* 问题描述 */}
        <div className="flex items-start gap-2">
          <AlertCircle className={`h-4 w-4 ${color} mt-0.5 flex-shrink-0`} />
          <p className="text-sm font-medium text-foreground">{item.issue}</p>
        </div>

        {/* 建议内容 - 默认直接展示 */}
        <div className="pt-2 pl-6 border-l-2 border-cyan-500/20 ml-1">
          <p className="text-sm text-muted-foreground leading-relaxed">
            <span className="text-emerald-400 font-medium">建议：</span>
            {item.suggestion}
          </p>
        </div>
      </div>
    </div>
  )
}
