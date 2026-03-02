import { useState, useCallback } from 'react'
import { geoAnalysisApi } from '@/lib/api'
import type { GEOAnalysisResponse } from '@/lib/types'

interface UseGeoAnalysisResult {
  currentAnalysis: GEOAnalysisResponse | null
  analysisList: GEOAnalysisResponse[]
  isLoading: boolean
  error: string | null
  createAnalysis: (url: string, platform?: string) => Promise<GEOAnalysisResponse | null>
  getAnalysis: (id: number) => Promise<void>
  fetchList: (params?: { page?: number; page_size?: number; status?: string }) => Promise<void>
  deleteAnalysis: (id: number) => Promise<boolean>
  reset: () => void
}

export function useGeoAnalysis(): UseGeoAnalysisResult {
  const [currentAnalysis, setCurrentAnalysis] = useState<GEOAnalysisResponse | null>(null)
  const [analysisList, setAnalysisList] = useState<GEOAnalysisResponse[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 创建分析任务
  const createAnalysis = useCallback(async (url: string, platform?: string): Promise<GEOAnalysisResponse | null> => {
    setIsLoading(true)
    setError(null)

    try {
      const response = await geoAnalysisApi.create(url, platform)
      const analysis = response.data.data as GEOAnalysisResponse
      setCurrentAnalysis(analysis)
      return analysis
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '创建分析任务失败'
      setError(errorMsg)
      return null
    } finally {
      setIsLoading(false)
    }
  }, [])

  // 获取分析详情
  const getAnalysis = useCallback(async (id: number) => {
    setIsLoading(true)
    setError(null)

    try {
      const response = await geoAnalysisApi.getById(id)
      setCurrentAnalysis(response.data.data as GEOAnalysisResponse)
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '获取分析详情失败'
      setError(errorMsg)
    } finally {
      setIsLoading(false)
    }
  }, [])

  // 获取分析列表
  const fetchList = useCallback(async (params?: { page?: number; page_size?: number; status?: string }) => {
    setIsLoading(true)
    setError(null)

    try {
      const response = await geoAnalysisApi.list(params)
      const pageData = response.data.data
      setAnalysisList(pageData?.list || [])
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '获取分析列表失败'
      setError(errorMsg)
    } finally {
      setIsLoading(false)
    }
  }, [])

  // 删除分析
  const deleteAnalysis = useCallback(async (id: number): Promise<boolean> => {
    setIsLoading(true)
    setError(null)

    try {
      await geoAnalysisApi.delete(id)
      // 从列表中移除
      setAnalysisList((prev) => prev.filter((item) => item.id !== id))
      // 如果当前查看的是被删除的项，清空
      if (currentAnalysis?.id === id) {
        setCurrentAnalysis(null)
      }
      return true
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '删除失败'
      setError(errorMsg)
      return false
    } finally {
      setIsLoading(false)
    }
  }, [currentAnalysis])

  const reset = useCallback(() => {
    setCurrentAnalysis(null)
    setError(null)
    setIsLoading(false)
  }, [])

  return {
    currentAnalysis,
    analysisList,
    isLoading,
    error,
    createAnalysis,
    getAnalysis,
    fetchList,
    deleteAnalysis,
    reset,
  }
}
