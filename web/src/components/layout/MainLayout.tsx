import type { ReactNode } from 'react'
import { Header } from './Header'

interface MainLayoutProps {
  children: ReactNode
}

export function MainLayout({ children }: MainLayoutProps) {
  return (
    <div className="relative min-h-screen overflow-hidden">
      {/* 背景效果 */}
      <div className="fixed inset-0 tech-grid -z-10" />

      {/* 渐变光晕 - 蓝青色 */}
      <div className="fixed top-0 left-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-[100px] -z-10 animate-pulse-slow" />
      <div className="fixed bottom-0 right-1/4 w-96 h-96 bg-cyan-500/10 rounded-full blur-[100px] -z-10 animate-pulse-slow" style={{ animationDelay: '1s' }} />
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-teal-500/5 rounded-full blur-[120px] -z-10" />

      {/* 动态光效 */}
      <div className="fixed top-1/4 right-0 w-64 h-64 bg-sky-500/5 rounded-full blur-[80px] -z-10 animate-float" />

      <Header />
      <main className="container mx-auto py-8 px-4">
        {children}
      </main>
    </div>
  )
}
