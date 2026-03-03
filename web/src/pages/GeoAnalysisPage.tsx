import { useEffect, useState } from 'react'
import { URLInputForm } from '@/components/geo/URLInputForm'
import { AnalysisResult } from '@/components/geo/AnalysisResult'
import { AnalysisList } from '@/components/geo/AnalysisList'
import { useGeoAnalysis } from '@/hooks/useGeoAnalysis'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { AlertCircle, History, Plus, Sparkles, Globe, Zap, Layers } from 'lucide-react'

export function GeoAnalysisPage() {
  const [viewMode, setViewMode] = useState<'create' | 'list' | 'detail'>('create')
  const {
    currentAnalysis,
    analysisList,
    isLoading,
    error,
    createAnalysis,
    getAnalysis,
    fetchList,
    deleteAnalysis,
    reset,
  } = useGeoAnalysis()

  // 加载历史列表
  useEffect(() => {
    if (viewMode === 'list') {
      fetchList()
    }
  }, [viewMode, fetchList])

  // 创建新分析
  const handleSubmit = async (url: string, platform: string) => {
    reset()
    const analysis = await createAnalysis(url, platform)
    if (analysis) {
      setViewMode('detail')
    }
  }

  // 查看详情
  const handleViewDetail = async (id: number) => {
    await getAnalysis(id)
    setViewMode('detail')
  }

  // 删除分析
  const handleDelete = async (id: number) => {
    if (confirm('确定要删除这条分析记录吗？')) {
      await deleteAnalysis(id)
    }
  }

  // 返回创建页面
  const handleBackToCreate = () => {
    reset()
    setViewMode('create')
  }

  return (
    <div className="space-y-8 max-w-6xl mx-auto">
      {/* 标题区域 - 科技感设计 */}
      <div className="relative">
        {/* 背景装饰 */}
        <div className="absolute -top-8 -left-8 w-32 h-32 bg-blue-500/10 rounded-full blur-3xl" />
        <div className="absolute -bottom-8 -right-8 w-32 h-32 bg-cyan-500/10 rounded-full blur-3xl" />

        <div className="relative flex flex-col md:flex-row md:items-center justify-between gap-6 p-6 rounded-2xl glass-card">
          <div className="space-y-3">
            <div className="flex items-center gap-3">
              <div className="p-2.5 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 border border-cyan-500/20">
                <Sparkles className="h-6 w-6 text-cyan-400" />
              </div>
              <div>
                <h1 className="text-2xl md:text-3xl font-bold font-display">
                  <span className="text-gradient">GEO 分析</span>
                </h1>
              </div>
            </div>
            <p className="text-muted-foreground max-w-md pl-1">
              利用 AI 技术分析网页内容在 Google AI Overview 中的 GEO 优化潜力，获取专业的优化建议和重写方案
            </p>
          </div>

          {/* 功能特性标签 */}
          <div className="flex flex-wrap gap-2">
            <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-blue-500/10 border border-blue-500/20 text-xs text-blue-400">
              <Layers className="h-3.5 w-3.5" />
              <span>多平台支持</span>
            </div>
            <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-cyan-500/10 border border-cyan-500/20 text-xs text-cyan-400">
              <Zap className="h-3.5 w-3.5" />
              <span>GEO 优化</span>
            </div>
            <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-teal-500/10 border border-teal-500/20 text-xs text-teal-400">
              <Sparkles className="h-3.5 w-3.5" />
              <span>AI 重写</span>
            </div>
          </div>
        </div>
      </div>

      {/* 导航按钮 */}
      <div className="flex gap-3">
        <Button
          variant={viewMode === 'create' || viewMode === 'detail' ? 'default' : 'outline'}
          onClick={handleBackToCreate}
          className="gap-2"
        >
          <Plus className="h-4 w-4" />
          新建分析
        </Button>
        <Button
          variant={viewMode === 'list' ? 'default' : 'outline'}
          onClick={() => setViewMode('list')}
          className="gap-2"
        >
          <History className="h-4 w-4" />
          历史记录
          {analysisList.length > 0 && (
            <span className="ml-1 text-xs bg-background/30 px-1.5 py-0.5 rounded-full">
              {analysisList.length}
            </span>
          )}
        </Button>
      </div>

      {/* 错误提示 */}
      {error && (
        <Card className="border-rose-500/30 bg-rose-500/5">
          <CardContent className="flex items-center gap-3 py-4">
            <div className="p-2 rounded-lg bg-rose-500/10">
              <AlertCircle className="h-5 w-5 text-rose-400" />
            </div>
            <span className="text-rose-400">{error}</span>
          </CardContent>
        </Card>
      )}

      {/* 创建分析 */}
      {viewMode === 'create' && (
        <Card className="relative glass-card">
          <div className="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-blue-500/5 to-cyan-500/5 rounded-full blur-3xl" />

          <CardHeader className="relative">
            <CardTitle className="text-lg flex items-center gap-2">
              <div className="p-1.5 rounded-md bg-cyan-500/10">
                <Globe className="h-4 w-4 text-cyan-400" />
              </div>
              输入 URL
            </CardTitle>
          </CardHeader>
          <CardContent className="relative">
            <URLInputForm onSubmit={handleSubmit} isLoading={isLoading} />
          </CardContent>
        </Card>
      )}

      {/* 历史列表 */}
      {viewMode === 'list' && (
        <AnalysisList
          list={analysisList}
          isLoading={isLoading}
          onView={handleViewDetail}
          onDelete={handleDelete}
        />
      )}

      {/* 分析详情 */}
      {viewMode === 'detail' && currentAnalysis && (
        <AnalysisResult
          analysis={currentAnalysis}
          isLoading={isLoading}
          onRefresh={() => getAnalysis(currentAnalysis.id)}
        />
      )}
    </div>
  )
}
