import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'

interface ScoreCardProps {
  score: number
}

export function ScoreCard({ score }: ScoreCardProps) {
  const getScoreColor = (score: number) => {
    if (score >= 80) return 'text-green-500'
    if (score >= 60) return 'text-yellow-500'
    if (score >= 40) return 'text-orange-500'
    return 'text-red-500'
  }

  const getScoreLabel = (score: number) => {
    if (score >= 80) return '优秀'
    if (score >= 60) return '良好'
    if (score >= 40) return '一般'
    return '需改进'
  }

  const getScoreBgColor = (score: number) => {
    if (score >= 80) return 'bg-green-500/10 border-green-200 dark:border-green-900'
    if (score >= 60) return 'bg-yellow-500/10 border-yellow-200 dark:border-yellow-900'
    if (score >= 40) return 'bg-orange-500/10 border-orange-200 dark:border-orange-900'
    return 'bg-red-500/10 border-red-200 dark:border-red-900'
  }

  const getProgressColor = (score: number) => {
    if (score >= 80) return 'bg-green-500'
    if (score >= 60) return 'bg-yellow-500'
    if (score >= 40) return 'bg-orange-500'
    return 'bg-red-500'
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">总体评分</CardTitle>
        <CardDescription>基于内容质量和 SEO 优化的综合评分</CardDescription>
      </CardHeader>
      <CardContent>
        <div className={`flex flex-col items-center justify-center py-6 rounded-lg border ${getScoreBgColor(score)}`}>
          <span className={`text-6xl font-bold ${getScoreColor(score)}`}>
            {score}
          </span>
          <span className="text-sm font-medium text-muted-foreground mt-2">
            {getScoreLabel(score)}
          </span>
          <div className="w-full max-w-xs mt-4 px-4">
            <div className="h-2 rounded-full bg-muted overflow-hidden">
              <div
                className={`h-full rounded-full transition-all duration-500 ${getProgressColor(score)}`}
                style={{ width: `${score}%` }}
              />
            </div>
          </div>
        </div>
        <div className="mt-4 grid grid-cols-4 gap-2 text-center text-xs text-muted-foreground">
          <div className="space-y-1">
            <div className="w-full h-1 rounded bg-red-500" />
            <span>0-39</span>
          </div>
          <div className="space-y-1">
            <div className="w-full h-1 rounded bg-orange-500" />
            <span>40-59</span>
          </div>
          <div className="space-y-1">
            <div className="w-full h-1 rounded bg-yellow-500" />
            <span>60-79</span>
          </div>
          <div className="space-y-1">
            <div className="w-full h-1 rounded bg-green-500" />
            <span>80-100</span>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
