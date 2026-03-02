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

  useEffect(() => {
    if (!analysisId) {
      return
    }

    // 关闭之前的连接
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }

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
      } catch (err) {
        console.error('Failed to parse connected event:', err)
      }
    })

    // 监听 progress 事件
    eventSource.addEventListener('progress', (event) => {
      try {
        const data = JSON.parse(event.data) as ProgressEventData
        setProgress(data)

        // 调用回调
        if (onProgress) {
          onProgress(data)
        }

        // 检查是否完成
        if (data.status === 'completed') {
          if (onComplete) {
            onComplete(data)
          }
          // 关闭连接
          eventSource.close()
        } else if (data.status === 'failed') {
          if (onError) {
            onError(data.message || '分析失败')
          }
          eventSource.close()
        }
      } catch (err) {
        console.error('Failed to parse progress event:', err)
      }
    })

    eventSource.onerror = () => {
      const errorMsg = 'SSE 连接错误'
      setError(errorMsg)
      setIsConnected(false)
      if (onError) {
        onError(errorMsg)
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
  }, [analysisId, onProgress, onComplete, onError])

  return {
    progress,
    isConnected,
    error,
  }
}
