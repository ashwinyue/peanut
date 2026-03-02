import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Loader2, Link2, Sparkles, ChevronDown } from 'lucide-react'
import { geoAnalysisApi } from '@/lib/api'
import type { PlatformConfig, PlatformType } from '@/lib/types'
import { PLATFORM_NAMES } from '@/lib/types'

const formSchema = z.object({
  url: z.string().url('请输入有效的 URL'),
  platform: z.string(),
})

type FormData = {
  url: string
  platform: string
}

interface URLInputFormProps {
  onSubmit: (url: string, platform: string) => void
  isLoading?: boolean
}

export function URLInputForm({ onSubmit, isLoading }: URLInputFormProps) {
  const [platforms, setPlatforms] = useState<PlatformConfig[]>([])
  const [showPlatformMenu, setShowPlatformMenu] = useState(false)
  const [selectedPlatform, setSelectedPlatform] = useState<PlatformType>('doubao')

  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
  } = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      url: '',
      platform: 'doubao',
    },
  })

  // 获取平台列表
  useEffect(() => {
    geoAnalysisApi.getPlatforms()
      .then(response => {
        if (response.data.data) {
          setPlatforms(response.data.data)
        }
      })
      .catch(console.error)
  }, [])

  const onFormSubmit = (data: FormData) => {
    onSubmit(data.url, data.platform)
  }

  const handlePlatformSelect = (platform: PlatformType) => {
    setSelectedPlatform(platform)
    setValue('platform', platform)
    setShowPlatformMenu(false)
  }

  const currentPlatform = platforms.find(p => p.type === selectedPlatform)

  return (
    <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-5">
      {/* 输入区域 */}
      <div className="space-y-3">
        <div className="flex items-center gap-2">
          <div className="flex h-6 w-6 items-center justify-center rounded-md bg-gradient-to-br from-blue-500/20 to-cyan-500/20">
            <Link2 className="h-3.5 w-3.5 text-cyan-400" />
          </div>
          <Label
            htmlFor="url"
            className="text-sm font-medium text-foreground/90"
          >
            目标网页 URL
          </Label>
        </div>

        <div className="relative group">
          {/* 渐变边框效果 */}
          <div className="absolute -inset-[1px] rounded-xl bg-gradient-to-r from-blue-500 via-cyan-400 to-teal-400 opacity-0 group-hover:opacity-50 group-focus-within:opacity-100 transition-opacity duration-300 blur-[2px]" />

          <div className="relative flex gap-2">
            <div className="relative flex-1">
              <Input
                id="url"
                placeholder="https://example.com/article"
                {...register('url')}
                disabled={isLoading}
                className="h-12 pl-4 pr-4 text-base bg-card/50 border-white/10 focus:border-cyan-500/50 focus:ring-2 focus:ring-cyan-500/20 rounded-xl transition-all duration-300 placeholder:text-muted-foreground/50"
              />
              {/* 聚焦时的发光效果 */}
              <div className="absolute inset-0 rounded-xl bg-gradient-to-r from-blue-500/5 to-cyan-500/5 opacity-0 group-focus-within:opacity-100 transition-opacity pointer-events-none" />
            </div>

            {/* 平台选择器 */}
            <div className="relative">
              <button
                type="button"
                onClick={() => setShowPlatformMenu(!showPlatformMenu)}
                disabled={isLoading}
                className="h-12 px-4 flex items-center gap-2 rounded-xl bg-card/80 border border-white/10 hover:border-white/20 transition-colors text-sm whitespace-nowrap"
              >
                <span className="text-muted-foreground">平台:</span>
                <span className="font-medium">{PLATFORM_NAMES[selectedPlatform]}</span>
                <ChevronDown className={`h-4 w-4 text-muted-foreground transition-transform ${showPlatformMenu ? 'rotate-180' : ''}`} />
              </button>

              {/* 平台下拉菜单 */}
              {showPlatformMenu && (
                <div className="absolute top-full left-0 mt-2 w-56 rounded-xl bg-card border border-white/10 shadow-xl z-50 overflow-hidden">
                  <div className="p-2 space-y-1">
                    {platforms.map((platform) => (
                      <button
                        key={platform.type}
                        type="button"
                        onClick={() => handlePlatformSelect(platform.type)}
                        className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-left transition-colors ${
                          selectedPlatform === platform.type
                            ? 'bg-primary/10 text-primary'
                            : 'hover:bg-white/5'
                        }`}
                      >
                        <div className={`w-2 h-2 rounded-full bg-gradient-to-r ${getPlatformColor(platform.type)}`} />
                        <div className="flex-1 min-w-0">
                          <div className="text-sm font-medium">{platform.name}</div>
                          <div className="text-xs text-muted-foreground truncate">{platform.description}</div>
                        </div>
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>

            <Button
              type="submit"
              disabled={isLoading}
              className="h-12 px-6 relative overflow-hidden group/btn"
            >
              {/* 按钮背景动画 */}
              <div className="absolute inset-0 bg-gradient-to-r from-blue-600 via-cyan-500 to-teal-500 bg-[length:200%_100%] animate-shimmer" />

              <span className="relative flex items-center gap-2 font-medium">
                {isLoading ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    分析中...
                  </>
                ) : (
                  <>
                    <Sparkles className="h-4 w-4" />
                    开始分析
                  </>
                )}
              </span>
            </Button>
          </div>
        </div>

        {errors.url && (
          <div className="flex items-center gap-2 text-sm text-rose-400 animate-fade-in">
            <div className="h-1 w-1 rounded-full bg-rose-400" />
            {errors.url.message}
          </div>
        )}

        {/* 选中平台权重提示 */}
        {currentPlatform && (
          <div className="flex items-center gap-4 text-xs text-muted-foreground">
            <span>权重配置:</span>
            <div className="flex items-center gap-3">
              <span className="px-2 py-0.5 rounded-full bg-blue-500/10 text-blue-400">权威性 {currentPlatform.weight.authority}%</span>
              <span className="px-2 py-0.5 rounded-full bg-purple-500/10 text-purple-400">时效性 {currentPlatform.weight.timeliness}%</span>
              <span className="px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400">结构化 {currentPlatform.weight.structure}%</span>
              {currentPlatform.weight.engagement > 10 && (
                <span className="px-2 py-0.5 rounded-full bg-amber-500/10 text-amber-400">互动性 {currentPlatform.weight.engagement}%</span>
              )}
            </div>
          </div>
        )}
      </div>

      {/* 提示文字 */}
      <div className="flex items-center gap-4 text-xs text-muted-foreground">
        <div className="flex items-center gap-1.5">
          <div className="h-1.5 w-1.5 rounded-full bg-cyan-400 animate-pulse" />
          <span>AI 驱动的深度分析</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="h-1.5 w-1.5 rounded-full bg-blue-400" />
          <span>多平台 GEO 优化</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="h-1.5 w-1.5 rounded-full bg-teal-400" />
          <span>智能优化建议</span>
        </div>
      </div>
    </form>
  )
}

// 获取平台颜色
function getPlatformColor(type: PlatformType): string {
  const colors: Record<PlatformType, string> = {
    doubao: 'from-blue-500 to-cyan-500',
    wechat: 'from-emerald-500 to-green-500',
    zhihu: 'from-blue-600 to-indigo-500',
    xiaohongshu: 'from-rose-500 to-pink-500',
    wenxin: 'from-purple-500 to-violet-500',
    yuanbao: 'from-cyan-500 to-teal-500',
  }
  return colors[type] || 'from-gray-500 to-gray-400'
}
