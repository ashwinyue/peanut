import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { TrendingUp, Target, Zap, Award, ArrowUpRight } from 'lucide-react'

interface ScoreCardProps {
  score: number
  title?: string
  description?: string
  variant?: 'default' | 'improved'
}

export function ScoreCard({
  score,
  title = '总体评分',
  description = '基于内容质量和 SEO 优化的综合评分',
  variant = 'default'
}: ScoreCardProps) {
  const [animatedScore, setAnimatedScore] = useState(0)
  const [isAnimating, setIsAnimating] = useState(true)

  // 数字滚动动画
  useEffect(() => {
    const duration = 1500
    const steps = 60
    const increment = score / steps
    let current = 0

    const timer = setInterval(() => {
      current += increment
      if (current >= score) {
        setAnimatedScore(score)
        setIsAnimating(false)
        clearInterval(timer)
      } else {
        setAnimatedScore(Math.round(current))
      }
    }, duration / steps)

    return () => clearInterval(timer)
  }, [score])

  const getScoreConfig = (score: number) => {
    if (score >= 80) return {
      color: 'text-emerald-400',
      bgColor: 'from-emerald-500/20 to-teal-500/10',
      borderColor: 'border-emerald-500/30',
      glowColor: 'shadow-emerald-500/20',
      label: '优秀',
      description: '内容质量极佳，SEO表现优异',
      icon: Award,
      gradient: 'from-emerald-400 via-teal-400 to-cyan-500'
    }
    if (score >= 60) return {
      color: 'text-cyan-400',
      bgColor: 'from-cyan-500/20 to-sky-500/10',
      borderColor: 'border-cyan-500/30',
      glowColor: 'shadow-cyan-500/20',
      label: '良好',
      description: '内容质量不错，仍有提升空间',
      icon: TrendingUp,
      gradient: 'from-cyan-400 via-sky-400 to-blue-500'
    }
    if (score >= 40) return {
      color: 'text-amber-400',
      bgColor: 'from-amber-500/20 to-orange-500/10',
      borderColor: 'border-amber-500/30',
      glowColor: 'shadow-amber-500/20',
      label: '一般',
      description: '需要针对性优化提升',
      icon: Target,
      gradient: 'from-amber-400 via-yellow-400 to-orange-500'
    }
    return {
      color: 'text-rose-400',
      bgColor: 'from-rose-500/20 to-red-500/10',
      borderColor: 'border-rose-500/30',
      glowColor: 'shadow-rose-500/20',
      label: '需改进',
      description: '急需优化，建议优先处理',
      icon: Zap,
      gradient: 'from-rose-400 via-red-400 to-orange-500'
    }
  }

  const config = getScoreConfig(score)
  const Icon = config.icon
  const circumference = 2 * Math.PI * 120
  const strokeDashoffset = circumference - (animatedScore / 100) * circumference

  // 优化后评分使用更醒目的样式
  const isImproved = variant === 'improved'

  return (
    <Card className={`relative overflow-hidden ${isImproved ? 'border-emerald-500/20' : ''}`}>
      {/* 背景装饰 */}
      <div className={`absolute inset-0 bg-gradient-to-br ${config.bgColor} opacity-50`} />
      <div className={`absolute -top-20 -right-20 w-40 h-40 bg-gradient-to-br ${config.gradient} opacity-10 blur-3xl rounded-full`} />

      {/* 优化后评分的特殊装饰 */}
      {isImproved && (
        <>
          <div className="absolute inset-0 bg-gradient-to-br from-emerald-500/5 via-blue-500/5 to-purple-500/5" />
          <div className="absolute top-0 right-0 p-2">
            <div className="flex items-center gap-1 px-2 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs">
              <ArrowUpRight className="h-3 w-3" />
              <span>优化后</span>
            </div>
          </div>
        </>
      )}

      <CardHeader className="relative">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-lg flex items-center gap-2">
              <Icon className={`h-5 w-5 ${config.color}`} />
              {title}
            </CardTitle>
            <CardDescription className="text-xs mt-0.5">{description}</CardDescription>
          </div>
          <div className="hidden sm:flex items-center gap-2">
            <div className={`px-3 py-1 rounded-full border ${config.borderColor} ${config.color} text-sm font-medium bg-card/50`}>
              {config.label}
            </div>
          </div>
        </div>
      </CardHeader>

      <CardContent className="relative">
        <div className="flex flex-col md:flex-row items-center gap-6">
          {/* 圆形进度条 */}
          <div className="relative flex-shrink-0">
            {/* 外发光 */}
            <div className={`absolute inset-0 rounded-full blur-xl opacity-30 ${config.glowColor}`} />

            <svg width="160" height="160" className="transform -rotate-90">
              {/* 背景圆环 */}
              <circle
                cx="80"
                cy="80"
                r="70"
                fill="none"
                stroke="hsl(var(--muted))"
                strokeWidth="10"
                strokeOpacity="0.1"
              />
              {/* 进度圆环 */}
              <circle
                cx="80"
                cy="80"
                r="70"
                fill="none"
                stroke={isImproved ? "url(#gradient-improved)" : "url(#gradient-default)"}
                strokeWidth="10"
                strokeLinecap="round"
                strokeDasharray={circumference}
                strokeDashoffset={strokeDashoffset}
                className="transition-all duration-1000 ease-out"
                style={{
                  filter: 'drop-shadow(0 0 8px currentColor)'
                }}
              />
              {/* 渐变定义 */}
              <defs>
                <linearGradient id="gradient-default" x1="0%" y1="0%" x2="100%" y2="0%">
                  <stop offset="0%" stopColor="hsl(207, 90%, 54%)" />
                  <stop offset="50%" stopColor="hsl(199, 89%, 48%)" />
                  <stop offset="100%" stopColor="hsl(188, 94%, 43%)" />
                </linearGradient>
                <linearGradient id="gradient-improved" x1="0%" y1="0%" x2="100%" y2="0%">
                  <stop offset="0%" stopColor="hsl(160, 84%, 39%)" />
                  <stop offset="50%" stopColor="hsl(152, 69%, 45%)" />
                  <stop offset="100%" stopColor="hsl(142, 76%, 36%)" />
                </linearGradient>
              </defs>
            </svg>

            {/* 中心分数 */}
            <div className="absolute inset-0 flex flex-col items-center justify-center">
              <span className={`text-5xl font-bold font-display ${isImproved ? 'text-emerald-400' : config.color} ${isAnimating ? 'scale-110' : 'scale-100'} transition-transform duration-300`}>
                {animatedScore}
              </span>
              <span className="text-xs text-muted-foreground mt-0.5">分</span>
            </div>
          </div>

          {/* 评分详情 */}
          <div className="flex-1 space-y-3 w-full">
            {/* 评分等级说明 */}
            <div className={`p-3 rounded-xl border ${isImproved ? 'bg-emerald-500/5 border-emerald-500/10' : 'bg-card/50 border-white/5'}`}>
              <p className="text-sm text-muted-foreground">{config.description}</p>
            </div>

            {/* 分段进度条 */}
            <div className="space-y-2">
              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>评分分布</span>
                <span className="font-mono">0-100</span>
              </div>

              <div className="grid grid-cols-4 gap-1.5">
                {[
                  { range: '0-39', color: 'bg-rose-500', label: '需改进' },
                  { range: '40-59', color: 'bg-amber-500', label: '一般' },
                  { range: '60-79', color: 'bg-cyan-500', label: '良好' },
                  { range: '80-100', color: 'bg-emerald-500', label: '优秀' },
                ].map((item, index) => {
                  const isActive = score >= parseInt(item.range.split('-')[0]) &&
                    score <= parseInt(item.range.split('-')[1] || '100')
                  return (
                    <div key={index} className="space-y-1">
                      <div
                        className={`h-1.5 rounded-full ${item.color} transition-all duration-500 ${isActive ? 'opacity-100 shadow-lg shadow-current/30' : 'opacity-30'}`}
                      />
                      <div className="text-center">
                        <div className="text-[9px] text-muted-foreground font-mono">{item.range}</div>
                        <div className={`text-[9px] ${isActive ? 'text-foreground font-medium' : 'text-muted-foreground/50'}`}>
                          {item.label}
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
