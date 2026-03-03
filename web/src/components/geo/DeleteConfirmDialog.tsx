import { useEffect, useLayoutEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { AlertTriangle, Trash2, X } from 'lucide-react'

interface DeleteConfirmDialogProps {
  isOpen: boolean
  onClose: () => void
  onConfirm: () => void
  title?: string
  description?: string
}

export function DeleteConfirmDialog({
  isOpen,
  onClose,
  onConfirm,
  title = '确认删除',
  description = '此操作无法撤销，确定要删除这条记录吗？',
}: DeleteConfirmDialogProps) {
  const [isVisible, setIsVisible] = useState(false)
  const [isAnimating, setIsAnimating] = useState(false)

  // 使用 useLayoutEffect 处理同步的 DOM 动画状态
  useLayoutEffect(() => {
    if (isOpen) {
      setIsVisible(true)
      // 在下一帧触发动画
      const rafId = requestAnimationFrame(() => {
        setIsAnimating(true)
      })
      return () => cancelAnimationFrame(rafId)
    } else {
      setIsAnimating(false)
      // 等待动画结束后隐藏
      const timer = setTimeout(() => {
        setIsVisible(false)
      }, 200)
      return () => clearTimeout(timer)
    }
  }, [isOpen])

  // 按 ESC 关闭
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose()
      }
    }
    window.addEventListener('keydown', handleEscape)
    return () => window.removeEventListener('keydown', handleEscape)
  }, [isOpen, onClose])

  if (!isVisible) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* 背景遮罩 */}
      <div
        className={`absolute inset-0 bg-black/60 backdrop-blur-sm transition-opacity duration-200 ${
          isAnimating ? 'opacity-100' : 'opacity-0'
        }`}
        onClick={onClose}
      />

      {/* 弹窗内容 */}
      <Card
        className={`relative w-full max-w-md glass-card border-rose-500/20 shadow-2xl shadow-rose-500/10 transition-all duration-200 ${
          isAnimating ? 'opacity-100 scale-100 translate-y-0' : 'opacity-0 scale-95 translate-y-4'
        }`}
      >
        {/* 关闭按钮 */}
        <button
          onClick={onClose}
          className="absolute top-4 right-4 p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-white/5 transition-colors"
        >
          <X className="h-4 w-4" />
        </button>

        {/* 图标区域 */}
        <div className="flex justify-center pt-8 pb-2">
          <div className="relative">
            {/* 外圈动画 */}
            <div className="absolute inset-0 rounded-full bg-rose-500/20 animate-ping" />
            {/* 图标容器 */}
            <div className="relative w-16 h-16 rounded-full bg-gradient-to-br from-rose-500/20 to-orange-500/20 border border-rose-500/30 flex items-center justify-center">
              <AlertTriangle className="h-8 w-8 text-rose-400" />
            </div>
          </div>
        </div>

        <CardHeader className="text-center pb-2">
          <CardTitle className="text-xl font-semibold text-rose-400">
            {title}
          </CardTitle>
        </CardHeader>

        <CardContent className="space-y-6">
          <p className="text-center text-muted-foreground text-sm leading-relaxed">
            {description}
          </p>

          {/* 按钮组 */}
          <div className="flex gap-3">
            <Button
              variant="outline"
              onClick={onClose}
              className="flex-1 h-11 border-white/10 hover:bg-white/5"
            >
              <X className="h-4 w-4 mr-2" />
              取消
            </Button>
            <Button
              variant="default"
              onClick={onConfirm}
              className="flex-1 h-11 bg-gradient-to-r from-rose-500 to-orange-500 hover:from-rose-600 hover:to-orange-600 text-white border-0 shadow-lg shadow-rose-500/25"
            >
              <Trash2 className="h-4 w-4 mr-2" />
              确认删除
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
