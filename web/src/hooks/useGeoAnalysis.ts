import { useState, useCallback } from 'react'
import { geoApi } from '@/lib/api'
import type { OptimizationReport, ProgressEvent } from '@/lib/types'

interface UseGeoAnalysisResult {
  report: OptimizationReport | null
  isLoading: boolean
  error: string | null
  progress: ProgressEvent | null
  analyze: (url: string) => Promise<void>
  analyzeWithStream: (url: string) => void
  reset: () => void
}

export function useGeoAnalysis(): UseGeoAnalysisResult {
  const [report, setReport] = useState<OptimizationReport | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [progress, setProgress] = useState<ProgressEvent | null>(null)

  // 即时分析（非流式）
  const analyze = useCallback(async (url: string) => {
    setIsLoading(true)
    setError(null)
    setReport(null)
    setProgress({
      step: 0,
      total: 1,
      agent: '',
      status: 'progress',
      message: '正在分析...',
    })

    try {
      const response = await geoApi.analyze(url)
      setReport(response.data.data as OptimizationReport)
      setProgress({
        step: 1,
        total: 1,
        agent: '',
        status: 'complete',
        message: '分析完成',
      })
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '分析失败'
      setError(errorMsg)
      setProgress((prev) => prev ? { ...prev, status: 'error', error: errorMsg } : null)
    } finally {
      setIsLoading(false)
    }
  }, [])

  // 流式分析（SSE）
  const analyzeWithStream = useCallback((url: string) => {
    setIsLoading(true)
    setError(null)
    setReport(null)
    setProgress({
      step: 0,
      total: 5,
      agent: '',
      status: 'progress',
      message: '正在连接...',
    })

    fetch('/api/v1/geo/analyze/stream', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'text/event-stream',
      },
      body: JSON.stringify({ url }),
    })
      .then(async (response) => {
        if (!response.ok) {
          const errorData = await response.json().catch(() => ({}))
          throw new Error(errorData.message || '分析请求失败')
        }

        const reader = response.body?.getReader()
        if (!reader) {
          throw new Error('无法读取响应流')
        }

        const decoder = new TextDecoder()
        let buffer = ''

        const processLine = (line: string) => {
          if (line.startsWith('event:')) {
            // 事件类型行，忽略
            return
          }

          if (line.startsWith('data:')) {
            const data = line.slice(5).trim()
            if (!data) return

            try {
              const parsed = JSON.parse(data)

              // 开始事件
              if (parsed.status === 'started') {
                setProgress({
                  step: 0,
                  total: 5,
                  agent: '',
                  status: 'progress',
                  message: '分析已开始',
                })
                return
              }

              // 进度事件
              if (parsed.step !== undefined && parsed.total !== undefined) {
                setProgress({
                  step: parsed.step,
                  total: parsed.total,
                  agent: parsed.agent || '',
                  message: parsed.message || '',
                  status: 'progress',
                })
                return
              }

              // 完成事件 - SSE 只返回 score，需要调用即时 API 获取完整报告
              if (parsed.status === 'success' && parsed.score !== undefined) {
                setProgress((prev) => ({
                  step: prev?.total || 5,
                  total: prev?.total || 5,
                  agent: '',
                  status: 'complete',
                  message: '分析完成，正在获取报告...',
                }))

                // 调用即时分析 API 获取完整报告
                geoApi.analyze(url)
                  .then((res) => {
                    setReport(res.data.data as OptimizationReport)
                  })
                  .catch((err) => {
                    setError(err instanceof Error ? err.message : '获取报告失败')
                  })
                  .finally(() => {
                    setIsLoading(false)
                  })
                return
              }

              // 错误事件
              if (parsed.error) {
                setError(parsed.error)
                setProgress((prev) => prev ? { ...prev, status: 'error', error: parsed.error } : null)
                setIsLoading(false)
              }
            } catch {
              console.error('Failed to parse SSE data:', data)
            }
          }
        }

        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          buffer += decoder.decode(value, { stream: true })
          const lines = buffer.split('\n')
          buffer = lines.pop() || ''

          for (const line of lines) {
            processLine(line)
          }
        }

        // 处理剩余的 buffer
        if (buffer.trim()) {
          processLine(buffer)
        }
      })
      .catch((err) => {
        setError(err.message)
        setProgress((prev) => prev ? { ...prev, status: 'error', error: err.message } : null)
        setIsLoading(false)
      })
  }, [])

  const reset = useCallback(() => {
    setReport(null)
    setError(null)
    setIsLoading(false)
    setProgress(null)
  }, [])

  return {
    report,
    isLoading,
    error,
    progress,
    analyze,
    analyzeWithStream,
    reset,
  }
}
