import { Sparkles, Activity } from 'lucide-react'

export function Header() {
  return (
    <header className="sticky top-0 z-50 w-full border-b border-white/5 bg-background/80 backdrop-blur-xl">
      {/* 顶部渐变线 - 蓝青 */}
      <div className="h-[2px] w-full bg-gradient-to-r from-blue-500 via-cyan-400 to-teal-400" />

      <div className="container mx-auto flex h-16 items-center px-4">
        {/* Logo */}
        <div className="flex items-center gap-3">
          <div className="relative">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-blue-500 via-cyan-400 to-teal-400 p-[1px]">
              <div className="flex h-full w-full items-center justify-center rounded-xl bg-card">
                <Sparkles className="h-5 w-5 text-cyan-400" />
              </div>
            </div>
            {/* 光晕效果 */}
            <div className="absolute -inset-1 rounded-xl bg-gradient-to-r from-blue-500/20 to-cyan-500/20 blur-md -z-10" />
          </div>

          <div className="flex flex-col">
            <span className="font-display text-xl font-bold tracking-tight">
              <span className="text-gradient">GEO</span>
              <span className="text-foreground/90"> Analyzer</span>
            </span>
            <span className="text-[10px] text-muted-foreground tracking-wider uppercase">
              Generative Engine Optimization
            </span>
          </div>
        </div>

        {/* 导航 */}
        <nav className="flex items-center gap-1 ml-8">
          <a
            href="/"
            className="group relative px-4 py-2 text-sm font-medium transition-colors"
          >
            <span className="relative z-10 text-foreground/80 group-hover:text-foreground transition-colors">
              分析
            </span>
            <span className="absolute inset-0 rounded-lg bg-cyan-500/10 opacity-0 group-hover:opacity-100 transition-opacity" />
            <span className="absolute bottom-0 left-1/2 -translate-x-1/2 w-0 h-0.5 bg-gradient-to-r from-blue-500 to-cyan-400 group-hover:w-1/2 transition-all duration-300" />
          </a>
          <a
            href="#history"
            className="group relative px-4 py-2 text-sm font-medium transition-colors"
          >
            <span className="relative z-10 text-muted-foreground group-hover:text-foreground transition-colors">
              历史记录
            </span>
            <span className="absolute inset-0 rounded-lg bg-cyan-500/10 opacity-0 group-hover:opacity-100 transition-opacity" />
          </a>
        </nav>

        {/* 右侧状态 */}
        <div className="ml-auto flex items-center gap-4">
          {/* 实时状态指示器 */}
          <div className="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-full bg-card border border-cyan-500/20">
            <Activity className="h-3.5 w-3.5 text-emerald-400 animate-pulse" />
            <span className="text-xs text-cyan-400/80 font-mono">SYSTEM ONLINE</span>
          </div>

          {/* 版本号 */}
          <div className="hidden md:flex items-center gap-2">
            <span className="text-xs text-muted-foreground font-mono">v2.0</span>
            <span className="h-4 w-px bg-white/10" />
            <span className="text-xs text-cyan-400/60">AI Powered</span>
          </div>
        </div>
      </div>
    </header>
  )
}
