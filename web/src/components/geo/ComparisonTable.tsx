import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import type { ComparisonItem } from '@/lib/types'

interface ComparisonTableProps {
  items: ComparisonItem[]
}

export function ComparisonTable({ items }: ComparisonTableProps) {
  if (!items || items.length === 0) return null

  const getSimilarityVariant = (
    similarity: string
  ): 'success' | 'warning' | 'destructive' | 'secondary' => {
    const s = similarity.toLowerCase()
    if (s === 'high') return 'success'
    if (s === 'medium') return 'warning'
    if (s === 'low') return 'destructive'
    return 'secondary'
  }

  const getSimilarityLabel = (similarity: string): string => {
    const s = similarity.toLowerCase()
    if (s === 'high') return '高'
    if (s === 'medium') return '中'
    if (s === 'low') return '低'
    return similarity
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">内容对比</CardTitle>
        <CardDescription>
          您的内容与 AI Overview 的对比分析
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b">
                <th className="text-left py-3 px-2 font-medium">维度</th>
                <th className="text-left py-3 px-2 font-medium">您的内容</th>
                <th className="text-left py-3 px-2 font-medium">AI Overview</th>
                <th className="text-left py-3 px-2 font-medium">相似度</th>
                <th className="text-left py-3 px-2 font-medium">差异</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item, index) => (
                <tr key={index} className="border-b last:border-0 hover:bg-muted/50">
                  <td className="py-3 px-2 font-medium">{item.dimension}</td>
                  <td className="py-3 px-2 text-muted-foreground max-w-xs">
                    <p className="line-clamp-3">{item.your_content}</p>
                  </td>
                  <td className="py-3 px-2 text-muted-foreground max-w-xs">
                    <p className="line-clamp-3">{item.ai_overview}</p>
                  </td>
                  <td className="py-3 px-2">
                    <Badge variant={getSimilarityVariant(item.similarity)}>
                      {getSimilarityLabel(item.similarity)}
                    </Badge>
                  </td>
                  <td className="py-3 px-2 text-muted-foreground max-w-xs">
                    <p className="line-clamp-3">{item.difference}</p>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>
  )
}
