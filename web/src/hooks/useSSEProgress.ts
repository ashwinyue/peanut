import { useState, useCallback, useRef } from 'react'
import type { ProgressEvent } from '@/lib/types'

interface UseSSEProgressOptions {
  onComplete?: (score: number) => void
  onError?: (error: string) => void
}

export function useSSEProgress(options: UseSSEProgressOptions = {}) {
  const [progress, setProgress] = useState<ProgressEvent | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const eventSourceRef = useRef<EventSource | null>(null)

  const connect = useCallback((url: string) => {
    // 关闭现有连接
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    setProgress({
      step: 0,
      total: 5,
      agent: '',
      status: 'progress',
      message: '正在连接...',
    })
    setIsConnected(true)

    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    eventSource.onopen = () => {
      setProgress((prev) => ({
        ...prev!,
        message: '已连接，等待分析开始...',
      }))
    }

    eventSource.addEventListener('start', () => {
      setProgress({
        step: 0,
        total: 5,
        agent: '',
        status: 'progress',
        message: '分析已开始',
      })
    })

    eventSource.addEventListener('progress', (event) => {
      try {
        const data = JSON.parse(event.data)
        setProgress({
          step: data.step,
          total: data.total,
          agent: data.agent,
          message: data.message,
          status: 'progress',
        })
      } catch {
        console.error('Failed to parse progress event')
      }
    })

    eventSource.addEventListener('complete', (event) => {
      try {
        const data = JSON.parse(event.data)
        setProgress((prev) => ({
          ...prev!,
          step: prev?.total || 5,
          status: 'complete',
          message: '分析完成',
        }))
        setIsConnected(false)
        eventSource.close()
        options.onComplete?.(data.score)
      } catch {
        console.error('Failed to parse complete event')
      }
    })

    eventSource.addEventListener('error', (event) => {
      try {
        const data = JSON.parse((event as MessageEvent).data)
        setProgress((prev) => ({
          ...prev!,
          status: 'error',
          error: data.error,
        }))
        setIsConnected(false)
        eventSource.close()
        options.onError?.(data.error)
      } catch {
        // 如果无法解析错误数据，可能是连接错误
        if (eventSource.readyState === EventSource.CLOSED) {
          setProgress((prev) => ({
            ...prev!,
            status: 'error',
            error: '连接已关闭',
          }))
          setIsConnected(false)
          options.onError?.('连接已关闭')
        }
      }
    })

    eventSource.onerror = () => {
      if (eventSource.readyState === EventSource.CLOSED) {
        setIsConnected(false)
      }
    }
  }, [options])

  const disconnect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }
    setIsConnected(false)
    setProgress(null)
  }, [])

  return {
    progress,
    isConnected,
    connect,
    disconnect,
  }
}
