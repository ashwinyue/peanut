import { useState } from 'react'
import { URLInputForm } from '@/components/geo/URLInputForm'
import { AnalysisProgress } from '@/components/geo/AnalysisProgress'
import { OptimizationReport } from '@/components/geo/OptimizationReport'
import { useGeoAnalysis } from '@/hooks/useGeoAnalysis'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { AlertCircle, RefreshCw } from 'lucide-react'

export function GeoAnalysisPage() {
  const [useStream, setUseStream] = useState(false)
  const {
    report,
    isLoading,
    error,
    progress,
    analyze,
    analyzeWithStream,
    reset,
  } = useGeoAnalysis()

  const handleSubmit = (url: string) => {
    reset()
    if (useStream) {
      analyzeWithStream(url)
    } else {
      analyze(url)
    }
  }

  const handleReset = () => {
    reset()
  }

  return (
    <div className="space-y-6">
      {/* 标题区域 */}
      <div className="space-y-2">
        <h1 className="text-3xl font-bold">GEO 分析</h1>
        <p className="text-muted-foreground">
          分析网页内容的搜索引擎优化潜力，获取 AI Overview 对比和优化建议
        </p>
      </div>

      {/* 输入区域 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">输入 URL</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* 分析模式选择 */}
          <div className="flex items-center gap-4">
            <span className="text-sm text-muted-foreground">分析模式：</span>
            <div className="flex gap-2">
              <Button
                variant={!useStream ? 'default' : 'outline'}
                size="sm"
                onClick={() => setUseStream(false)}
                disabled={isLoading}
              >
                即时分析
              </Button>
              <Button
                variant={useStream ? 'default' : 'outline'}
                size="sm"
                onClick={() => setUseStream(true)}
                disabled={isLoading}
              >
                流式分析
              </Button>
            </div>
            <span className="text-xs text-muted-foreground">
              {useStream ? '实时显示分析进度' : '等待完整结果返回'}
            </span>
          </div>

          {/* URL 输入表单 */}
          <URLInputForm onSubmit={handleSubmit} isLoading={isLoading} />
        </CardContent>
      </Card>

      {/* 错误提示 */}
      {error && (
        <Card className="border-destructive">
          <CardContent className="flex items-center gap-2 py-4">
            <AlertCircle className="h-5 w-5 text-destructive" />
            <span className="text-destructive">{error}</span>
          </CardContent>
        </Card>
      )}

      {/* 进度显示 */}
      {isLoading && progress && <AnalysisProgress progress={progress} />}

      {/* 分析报告 */}
      {report && !isLoading && (
        <div className="space-y-4">
          <div className="flex justify-end">
            <Button variant="outline" onClick={handleReset}>
              <RefreshCw className="h-4 w-4 mr-2" />
              重新分析
            </Button>
          </div>
          <OptimizationReport report={report} />
        </div>
      )}
    </div>
  )
}
