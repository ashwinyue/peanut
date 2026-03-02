import { useEffect, useRef, useState } from 'react'
import type { ProgressEventData } from '@/lib/types'

import { getApiBaseUrl } from '@/lib/api'

interface UseSSEProgressResult {
  progress: ProgressEventData | null
  isConnected: boolean
  error: string | null
}

/**
 * SSE 进度监听 Hook
 * @param analysisId 分析任务 ID
 * @param onProgress 进度回调
 * @param onComplete 完成回调
 * @param onError 错误回调
 */
export function useSSEProgress(
  analysisId: number | null,
  onProgress?: (data: ProgressEventData) => void,
  onComplete?: (data: ProgressEventData) => void,
  onError?: (error: string) => void
): UseSSEProgressResult {
  const [progress, setProgress] = useState<ProgressEventData | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const eventSourceRef = useRef<EventSource | null>(null)

  // 使用 ref 保存回调函数，避免依赖变化触发重新连接
  const callbacksRef = useRef({ onProgress, onComplete, onError })
  useEffect(() => {
    callbacksRef.current = { onProgress, onComplete, onError }
  }, [onProgress, onComplete, onError])

  useEffect(() => {
    if (!analysisId) {
      return
    }

    // 关闭之前的连接
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }

    // 重置状态
    setProgress(null)
    setIsConnected(false)
    setError(null)

    // 创建 SSE 连接
    const url = `${getApiBaseUrl()}/geo/analysis/${analysisId}/progress`

    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    eventSource.onopen = () => {
      setIsConnected(true)
      setError(null)
    }

    // 监听 connected 事件
    eventSource.addEventListener('connected', (event) => {
      try {
        const data = JSON.parse(event.data)
        console.log('SSE connected:', data)
      } catch {
        console.error('Failed to parse connected event')
      }
    })

    // 监听 progress 事件
    eventSource.addEventListener('progress', (event) => {
      try {
        const data = JSON.parse(event.data) as ProgressEventData
        setProgress(data)

        // 调用回调
        if (callbacksRef.current.onProgress) {
          callbacksRef.current.onProgress(data)
        }

        // 检查是否完成
        if (data.status === 'completed') {
          if (callbacksRef.current.onComplete) {
            callbacksRef.current.onComplete(data)
          }
          // 关闭连接
          eventSource.close()
        } else if (data.status === 'failed') {
          if (callbacksRef.current.onError) {
            callbacksRef.current.onError(data.message || '分析失败')
          }
          eventSource.close()
        }
      } catch {
        console.error('Failed to parse progress event')
      }
    })

    eventSource.onerror = () => {
      const errorMsg = 'SSE 连接错误'
      setError(errorMsg)
      setIsConnected(false)
      if (callbacksRef.current.onError) {
        callbacksRef.current.onError(errorMsg)
      }
    }

    return () => {
      // 清理连接
      if (eventSourceRef.current) {
        eventSourceRef.current.close()
        eventSourceRef.current = null
      }
      setIsConnected(false)
    }
  }, [analysisId]) // 只依赖 analysisId，避免回调函数变化导致重建

  return {
    progress,
    isConnected,
    error,
  }
}
