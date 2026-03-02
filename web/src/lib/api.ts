import axios, { type AxiosInstance, type AxiosResponse } from 'axios'
import type { ApiResponse } from './types'

// 创建 Axios 实例
const api: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 120000, // 2 分钟超时，适应 GEO 分析的长时间请求
  headers: {
    'Content-Type': 'application/json',
  },
})

// 响应拦截器
api.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const { data } = response
    // 业务错误处理
    if (data.code !== 0) {
      return Promise.reject(new Error(data.message || '请求失败'))
    }
    return response
  },
  (error) => {
    // 网络错误处理
    const message = error.response?.data?.message || error.message || '网络错误'
    return Promise.reject(new Error(message))
  }
)

export default api

// GEO 分析 API
export const geoApi = {
  // 即时分析
  analyze: (url: string) =>
    api.post<ApiResponse>('/geo/analyze', { url }),

  // 获取 SSE 流式分析的 URL
  getStreamUrl: () => '/api/v1/geo/analyze/stream',
}
