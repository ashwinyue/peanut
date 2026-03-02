import axios, { type AxiosInstance, type AxiosResponse } from 'axios'
import type { ApiResponse, PageData, GEOAnalysisResponse, GEOAnalysisListRequest, ProgressEventData, PlatformConfig } from './types'

// 获取 API 基础 URL
export function getApiBaseUrl(): string {
  // 优先使用环境变量（生产环境）
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL
  }
  // 默认使用相对路径（开发环境）
  return '/api/v1'
}

// 创建 Axios 实例
const api: AxiosInstance = axios.create({
  baseURL: getApiBaseUrl(),
  timeout: 120000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 响应拦截器
api.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const { data } = response
    if (data.code !== 0) {
      return Promise.reject(new Error(data.message || '请求失败'))
    }
    return response
  },
  (error) => {
    const message = error.response?.data?.message || error.message || '网络错误'
    return Promise.reject(new Error(message))
  }
)

export default api

// GEO 分析任务 API
export const geoAnalysisApi = {
  // 创建分析任务
  create: (url: string, platform?: string) =>
    api.post<ApiResponse<GEOAnalysisResponse>>('/geo/analysis', { url, platform }),

  // 获取分析详情
  getById: (id: number) =>
    api.get<ApiResponse<GEOAnalysisResponse>>(`/geo/analysis/${id}`),

  // 获取分析列表
  list: (params?: GEOAnalysisListRequest) =>
    api.get<ApiResponse<PageData<GEOAnalysisResponse>>>('/geo/analysis', { params }),

  // 删除分析
  delete: (id: number) =>
    api.delete<ApiResponse>(`/geo/analysis/${id}`),

  // 获取支持的平台列表
  getPlatforms: () =>
    api.get<ApiResponse<PlatformConfig[]>>('/geo/analysis/platforms'),

  // 订阅进度更新（SSE）
  subscribeProgress: (
    id: number,
    onProgress: (data: ProgressEventData) => void,
    onError?: (error: string) => void
  ): (() => void) => {
    const eventSource = new EventSource(`${getApiBaseUrl()}/geo/analysis/${id}/progress`)

    eventSource.addEventListener('progress', (event) => {
      try {
        const data = JSON.parse(event.data) as ProgressEventData
        onProgress(data)

        // 如果是终态，关闭连接
        if (data.status === 'completed' || data.status === 'failed') {
          eventSource.close()
        }
      } catch {
        console.error('Failed to parse progress event')
      }
    })

    eventSource.onerror = () => {
      onError?.('连接中断')
      eventSource.close()
    }

    // 返回取消订阅函数
    return () => {
      eventSource.close()
    }
  },
}
