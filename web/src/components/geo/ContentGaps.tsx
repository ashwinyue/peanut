import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { FileWarning } from 'lucide-react'

interface ContentGapsProps {
  gaps: string[]
}

export function ContentGaps({ gaps }: ContentGapsProps) {
  if (!gaps || gaps.length === 0) return null

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">内容差距</CardTitle>
        <CardDescription>
          您的内容中缺失或需要补充的部分，填补这些差距可以提升内容完整性
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {gaps.map((gap, index) => (
            <div
              key={index}
              className="flex items-start gap-3 p-3 rounded-lg border bg-orange-50 dark:bg-orange-950/20 border-orange-200 dark:border-orange-900"
            >
              <FileWarning className="h-5 w-5 text-orange-500 shrink-0 mt-0.5" />
              <span className="text-sm leading-relaxed">{gap}</span>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
