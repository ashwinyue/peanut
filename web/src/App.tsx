import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MainLayout } from '@/components/layout/MainLayout'
import { GeoAnalysisPage } from '@/pages/GeoAnalysisPage'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <MainLayout>
        <GeoAnalysisPage />
      </MainLayout>
    </QueryClientProvider>
  )
}

export default App
