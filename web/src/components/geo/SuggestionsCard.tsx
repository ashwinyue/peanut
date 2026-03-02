import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import type { OptimizationSuggestion } from '@/lib/types'
import { AlertCircle, AlertTriangle, Info } from 'lucide-react'

interface SuggestionsCardProps {
  suggestions: OptimizationSuggestion[]
}

export function SuggestionsCard({ suggestions }: SuggestionsCardProps) {
  if (!suggestions || suggestions.length === 0) return null

  const groupedSuggestions = {
    high: suggestions.filter((s) => s.priority.toLowerCase() === 'high'),
    medium: suggestions.filter((s) => s.priority.toLowerCase() === 'medium'),
    low: suggestions.filter((s) => s.priority.toLowerCase() === 'low'),
  }

  const getPriorityIcon = (priority: string) => {
    switch (priority.toLowerCase()) {
      case 'high':
        return <AlertCircle className="h-4 w-4 text-red-500" />
      case 'medium':
        return <AlertTriangle className="h-4 w-4 text-yellow-500" />
      default:
        return <Info className="h-4 w-4 text-blue-500" />
    }
  }

  const getPriorityVariant = (
    priority: string
  ): 'destructive' | 'warning' | 'secondary' => {
    switch (priority.toLowerCase()) {
      case 'high':
        return 'destructive'
      case 'medium':
        return 'warning'
      default:
        return 'secondary'
    }
  }

  const getPriorityLabel = (priority: string): string => {
    switch (priority.toLowerCase()) {
      case 'high':
        return '高优先级'
      case 'medium':
        return '中优先级'
      default:
        return '低优先级'
    }
  }

  const renderSuggestionGroup = (
    priority: 'high' | 'medium' | 'low',
    items: OptimizationSuggestion[]
  ) => {
    if (items.length === 0) return null

    return (
      <div className="space-y-3">
        <div className="flex items-center gap-2 sticky top-0 bg-card py-1">
          {getPriorityIcon(priority)}
          <span className="font-medium">{getPriorityLabel(priority)}</span>
          <Badge variant={getPriorityVariant(priority)}>
            {items.length} 项
          </Badge>
        </div>
        <div className="space-y-3 pl-6">
          {items.map((item, index) => (
            <div
              key={index}
              className="p-4 rounded-lg border bg-card"
            >
              <div className="flex items-center gap-2 mb-2">
                <Badge variant="outline" className="text-xs">
                  {item.category}
                </Badge>
              </div>
              <p className="text-sm font-medium mb-2">{item.issue}</p>
              <p className="text-sm text-muted-foreground">
                <span className="font-medium text-foreground">建议：</span>
                {item.suggestion}
              </p>
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">优化建议</CardTitle>
        <CardDescription>
          根据分析结果提供的改进建议
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          {renderSuggestionGroup('high', groupedSuggestions.high)}
          {renderSuggestionGroup('medium', groupedSuggestions.medium)}
          {renderSuggestionGroup('low', groupedSuggestions.low)}
        </div>
      </CardContent>
    </Card>
  )
}
