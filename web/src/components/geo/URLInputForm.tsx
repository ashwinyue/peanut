import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Loader2 } from 'lucide-react'

const formSchema = z.object({
  url: z.string().url('请输入有效的 URL'),
})

type FormData = z.infer<typeof formSchema>

interface URLInputFormProps {
  onSubmit: (url: string) => void
  isLoading?: boolean
}

export function URLInputForm({ onSubmit, isLoading }: URLInputFormProps) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      url: '',
    },
  })

  const onFormSubmit = (data: FormData) => {
    onSubmit(data.url)
  }

  return (
    <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="url">网页 URL</Label>
        <div className="flex gap-2">
          <Input
            id="url"
            placeholder="https://example.com"
            {...register('url')}
            disabled={isLoading}
            className="flex-1"
          />
          <Button type="submit" disabled={isLoading}>
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                分析中...
              </>
            ) : (
              '开始分析'
            )}
          </Button>
        </div>
        {errors.url && (
          <p className="text-sm text-destructive">{errors.url.message}</p>
        )}
      </div>
    </form>
  )
}
