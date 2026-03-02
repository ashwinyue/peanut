import { Search } from 'lucide-react'

export function Header() {
  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container mx-auto flex h-14 items-center px-4">
        <div className="flex items-center space-x-2">
          <Search className="h-6 w-6 text-primary" />
          <span className="text-xl font-bold">GEO Analyzer</span>
        </div>
        <nav className="flex items-center space-x-6 ml-6">
          <a
            href="/"
            className="text-sm font-medium text-muted-foreground transition-colors hover:text-primary"
          >
            分析
          </a>
        </nav>
        <div className="ml-auto flex items-center space-x-4">
          <span className="text-sm text-muted-foreground">
            Generative Engine Optimization
          </span>
        </div>
      </div>
    </header>
  )
}
