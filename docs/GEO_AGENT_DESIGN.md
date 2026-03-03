# GEO Agent 设计文档

> **GEO** (Generative Engine Optimization，生成式引擎优化) —— SEO 的 AI 时代进化版
>
> 基于 [Eino](https://github.com/cloudwego/eino) 框架构建的智能体，用于优化内容在生成式搜索引擎（ChatGPT、Perplexity、Google AI Overviews）中的引用和推荐率。

---

## 背景与动机

### 从 SEO 到 GEO

| 维度 | SEO (传统) | GEO (AI 时代) |
|------|-----------|--------------|
| **目标** | 搜索结果排名更高 | 被 AI 引用/推荐 |
| **优化对象** | 关键词、链接、页面结构 | 内容质量、权威性、AI 可读性 |
| **核心策略** | 技术优化、外链建设 | 知识图谱、结构化数据、E-E-A-T |
| **搜索结果** | 10 个蓝色链接 | AI 生成摘要 + 引用来源 |
| **用户行为** | 点击链接进入网站 | 直接获取答案（零点击） |

### 生成式搜索引擎的影响

- **ChatGPT Search**：OpenAI 的搜索功能
- **Perplexity**：AI 原生搜索引擎
- **Google AI Overviews**：Google 的 AI 摘要
- **Bing Copilot**：微软的 AI 助手

**关键变化**：用户不再点击链接，而是直接获取 AI 生成的答案。因此，优化目标从「点击率」转变为「引用率」。

---

## GEO Agent 架构设计

### 系统架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                     Peanut GEO Agent System                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Web Frontend (React + Vite)                              │  │
│  │  - URL 输入表单                                            │  │
│  │  - 实时进度展示 (SSE)                                       │  │
│  │  - 优化结果展示                                            │  │
│  │  - 历史记录管理                                            │  │
│  └──────────────────────────────────────────────────────────┘  │
│                           ↓                                      │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  API Layer (Gin)                                          │  │
│  │  - /api/v1/geo/analysis                                   │  │
│  │  - SSE 进度推送                                            │  │
│  └──────────────────────────────────────────────────────────┘  │
│                           ↓                                      │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Service Layer                                            │  │
│  │  - GEOAnalysisService                                     │  │
│  │  - Progress Manager                                       │  │
│  └──────────────────────────────────────────────────────────┘  │
│                           ↓                                      │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  GEO Agent Flow (Eino Graph)                              │  │
│  │                                                           │  │
│  │  ┌─────────┐    ┌─────────┐    ┌─────────┐               │  │
│  │  │ Title   │───→│ Query   │───→│ Main    │               │  │
│  │  │ Scraper │    │ Research│    │ Query   │               │  │
│  │  └─────────┘    └────┬────┘    │ Extract │               │  │
│  │                      │         └────┬────┘               │  │
│  │                      ↓              │                     │  │
│  │  ┌─────────┐    ┌─────────┐         │                     │  │
│  │  │ Content │←───│ Query   │         │                     │  │
│  │  │ Rewriter│←───│ Summary │←────────┘                     │  │
│  │  └────┬────┘    └─────────┘                               │  │
│  │       ↑                                                   │  │
│  │  ┌────┴────┐    ┌─────────┐                               │  │
│  │  │ Content │←───│ AI      │                               │  │
│  │  │ Optimizer│    │ Overview│                               │  │
│  │  └─────────┘    └─────────┘                               │  │
│  │                                                           │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
├─────────────────────────────────────────────────────────────────┤
│  External Services                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │ Bright Data  │  │   豆包 LLM   │  │    SQLite/PostgreSQL │  │
│  │ Web Unlocker │  │  (Ark API)   │  │                      │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 工作流程

```
用户输入 URL
    │
    ▼
┌─────────────────┐
│ 1. TitleScraper │──→ 爬取网页标题和内容 (Bright Data Web Unlocker)
└────────┬────────┘
         │
         ▼
┌────────────────────┐
│ 2. QueryResearcher │──→ Google 搜索获取相关查询 (Bright Data SERP)
└────────┬───────────┘
         │
         ├──→ ┌──────────────────┐
         │    │ 5. QuerySummarizer│──→ 总结查询发散结果
         │    └──────────────────┘
         │
         ▼
┌──────────────────────┐
│ 3. MainQueryExtractor│──→ 提炼核心搜索词和意图
└────────┬─────────────┘
         │
         ▼
┌───────────────────────┐
│ 4. AIOverviewRetriever│──→ 获取 AI Overview 摘要
└────────┬──────────────┘
         │
         ▼
┌───────────────────┐
│ 6. ContentOptimizer│──→ 对比分析生成优化报告
└────────┬──────────┘
         │
         ▼
┌──────────────────┐
│ 7. ContentRewriter│──→ 生成优化后的文章内容
└──────────────────┘
```

---

## 核心功能模块

### 1. Flow Graph 架构

基于 Eino 的 `compose.Graph` 实现，使用状态机模式管理 Agent 流转：

```go
// internal/agent/geo/flow/builder.go
func BuildGraph(ctx context.Context) (compose.Runnable[string, string], error) {
    g := compose.NewGraph[string, string](
        compose.WithGenLocalState(GenLocalState),
    )

    // 添加 7 个 Agent 节点
    _ = g.AddGraphNode(AgentTitleScraper, titleScraperGraph)
    _ = g.AddGraphNode(AgentQueryResearcher, queryResearcherGraph)
    // ... 其他节点

    // 添加分支控制
    _ = g.AddBranch(AgentTitleScraper, compose.NewGraphBranch(agentHandOff, outMap))
    // ... 其他分支

    // 编译 Graph
    return g.Compile(ctx, compose.WithGraphName("GEOFlow"))
}
```

### 2. 状态管理

使用 `FlowState` 在各 Agent 间传递数据：

```go
// internal/agent/geo/models/flow.go
type FlowState struct {
    // 输入参数
    URL          string `json:"url,omitempty"`
    PlatformType string `json:"platform_type,omitempty"`

    // 步骤 1: 网页爬取结果
    Title   string `json:"title,omitempty"`
    Content string `json:"content,omitempty"`

    // 步骤 2: 查询发散结果
    QueryFanout   []string       `json:"query_fanout,omitempty"`
    SearchResults []SearchResult `json:"search_results,omitempty"`

    // 步骤 3: 主查询提取
    MainQuery    string   `json:"main_query,omitempty"`
    Keywords     []string `json:"keywords,omitempty"`
    SearchIntent string   `json:"search_intent,omitempty"`

    // 步骤 4: AI 摘要
    AIOverview string   `json:"ai_overview,omitempty"`
    Sources    []string `json:"sources,omitempty"`

    // 步骤 5: 查询总结
    QuerySummary string `json:"query_summary,omitempty"`

    // 步骤 6: 优化报告
    Report *OptimizationReport `json:"report,omitempty"`

    // 步骤 7: 重写后的文章
    OptimizedArticle string `json:"optimized_article,omitempty"`

    // 流程控制
    Goto     string `json:"goto,omitempty"`  // 下一个 Agent
    Step     int    `json:"step,omitempty"`
    MaxSteps int    `json:"max_steps,omitempty"`
}
```

### 3. Agent 实现

每个 Agent 是一个独立的子图（SubGraph），实现特定的业务逻辑：

```go
// internal/agent/geo/flow/agents/title_scraper.go
func NewTitleScraperAgent(ctx context.Context, scraper *tools.BrightDataWebScraper) (compose.Graph, error) {
    g := compose.NewGraph[string, string]()

    // 添加处理节点
    _ = g.AddLambdaNode("process", compose.InvokableLambda(
        func(ctx context.Context, input string) (string, error) {
            var state *models.FlowState
            // 从 context 获取状态
            _ = compose.ProcessState[*models.FlowState](ctx, func(_ context.Context, s *models.FlowState) error {
                state = s
                return nil
            })

            // 爬取网页
            result, err := scraper.Scrape(ctx, state.URL)
            if err != nil {
                return "", err
            }

            // 更新状态
            state.Title = result.Title
            state.Content = result.Content
            state.Goto = AgentQueryResearcher  // 设置下一个 Agent

            return result.Title, nil
        },
    ))

    _ = g.AddEdge(compose.START, "process")
    _ = g.AddEdge("process", compose.END)

    return g, nil
}
```

### 4. 工具集成

#### Bright Data 工具

```go
// internal/agent/geo/tools/brightdata.go

// Web Scraper - 爬取网页内容
type BrightDataWebScraper struct {
    config     *BrightDataConfig
    httpClient *http.Client
}

func (s *BrightDataWebScraper) Scrape(ctx context.Context, targetURL string) (*models.ScrapedTitle, error) {
    // 使用 Bright Data Web Unlocker API
}

// SERP Searcher - 搜索相关查询
type BrightDataSearcher struct {
    config     *BrightDataConfig
    httpClient *http.Client
}

func (s *BrightDataSearcher) Search(ctx context.Context, query string) (*models.QueryFanout, error) {
    // 使用 Bright Data SERP API
}

// SERP Provider - 获取 AI Overview
type BrightDataSERPProvider struct {
    config     *BrightDataConfig
    httpClient *http.Client
}

func (p *BrightDataSERPProvider) GetAIOverview(ctx context.Context, query string) (*models.AIOverview, error) {
    // 调用 brd_ai_overview=2 获取 Google AI Overview
}
```

### 5. 输出结果

```go
// internal/agent/geo/models/response.go
type OptimizationReport struct {
    URL                     string                   `json:"url"`
    Title                   string                   `json:"title"`
    MainQuery               string                   `json:"main_query"`
    QueryFanout             string                   `json:"query_fanout"`
    QueryFanoutSummary      string                   `json:"query_fanout_summary"`
    AIOverview              string                   `json:"ai_overview"`
    ComparisonTable         []ComparisonItem         `json:"comparison_table"`
    ContentGaps             []string                 `json:"content_gaps"`
    OptimizationSuggestions []OptimizationSuggestion `json:"optimization_suggestions"`
    OptimizationReport      string                   `json:"optimization_report"`
    OptimizedArticle        string                   `json:"optimized_article"`
    OverallScore            int                      `json:"overall_score"`
    Timestamp               time.Time                `json:"timestamp"`
}
```

---

## 技术栈

### 核心框架

| 组件 | 选型 | 说明 |
|------|------|------|
| **Agent 框架** | [Eino](https://github.com/cloudwego/eino) | 字节开源的 Go AI Agent 开发框架 |
| **Web 框架** | Gin | API 服务 |
| **ORM** | GORM | 数据持久化 |
| **数据库** | SQLite (默认) / PostgreSQL | 存储分析记录 |
| **前端** | React 19 + TypeScript + Vite | 用户界面 |
| **UI 组件** | shadcn/ui + TailwindCSS | 组件库 |
| **状态管理** | Zustand | 前端状态 |
| **数据获取** | TanStack Query | API 数据获取 |

### 外部服务

| 服务 | 用途 | 环境变量 |
|------|------|---------|
| **Bright Data** | 网页爬取 & SERP 数据 | `BRIGHT_DATA_API_KEY`, `BRIGHT_DATA_ZONE` |
| **豆包 LLM** | 内容分析与生成 | `ARK_API_KEY`, `ARK_BASE_URL`, `ARK_MODEL` |

---

## 目录结构

```
internal/agent/geo/
├── service.go                 # GEO 服务入口
├── flow/                      # Flow Graph 实现
│   ├── builder.go            # Graph 构建器
│   ├── state.go              # 状态定义导出
│   ├── consts.go             # Agent 名称常量
│   ├── interface.go          # 接口定义
│   └── agents/               # 各 Agent 实现
│       ├── title_scraper.go
│       ├── query_researcher.go
│       ├── main_query_extractor.go
│       ├── ai_overview_retriever.go
│       ├── query_summarizer.go
│       ├── content_optimizer.go
│       └── content_rewriter.go
├── models/                    # 数据模型
│   ├── flow.go               # FlowState 定义
│   ├── response.go           # API 响应模型
│   ├── platform.go           # 平台配置
│   └── validation.go         # 验证规则
├── tools/                     # 外部工具
│   ├── brightdata.go         # Bright Data 集成
│   └── scraper.go            # 爬取工具
├── llm/                       # LLM 配置
│   ├── model.go              # 模型配置
│   └── ark.go                # 豆包 API 封装
└── parser/                    # 内容解析
    └── parser.go             # HTML/Markdown 解析

web/src/components/geo/        # 前端组件
├── URLInputForm.tsx          # URL 输入表单
├── AnalysisResult.tsx        # 分析结果展示
├── ValidationResultCard.tsx  # 验证结果卡片
└── ...
```

---

## API 设计

### 创建分析任务

```http
POST /api/v1/geo/analysis
Content-Type: application/json

{
  "url": "https://example.com/article",
  "platform": "google"
}

Response:
{
  "success": true,
  "data": {
    "id": 1,
    "url": "https://example.com/article",
    "platform": "google",
    "status": "processing",
    "created_at": "2025-03-03T10:00:00Z"
  }
}
```

### 获取分析进度 (SSE)

```http
GET /api/v1/geo/analysis/{id}/progress
Accept: text/event-stream

Events:
data: {"step": 1, "total": 7, "agent": "title_scraper", "status": "progress"}

data: {"step": 2, "total": 7, "agent": "query_researcher", "status": "progress"}

data: {"step": 7, "total": 7, "agent": "content_rewriter", "status": "completed"}
```

### 获取分析结果

```http
GET /api/v1/geo/analysis/{id}

Response:
{
  "success": true,
  "data": {
    "id": 1,
    "url": "https://example.com/article",
    "title": "文章标题",
    "main_query": "核心查询词",
    "query_fanout": "相关查询列表",
    "query_fanout_summary": "查询发散总结",
    "ai_overview": "AI 摘要内容",
    "comparison_table": [...],
    "content_gaps": [...],
    "optimization_suggestions": [...],
    "optimization_report": "优化报告内容",
    "optimized_article": "重写后的文章",
    "overall_score": 85,
    "status": "completed"
  }
}
```

### 获取支持的平台

```http
GET /api/v1/geo/analysis/platforms

Response:
{
  "success": true,
  "data": [
    {
      "id": "google",
      "name": "Google AI Overview",
      "description": "Google 搜索的 AI 摘要功能",
      "weight": 1.0
    },
    {
      "id": "perplexity",
      "name": "Perplexity AI",
      "description": "AI 原生搜索引擎",
      "weight": 0.9
    },
    {
      "id": "chatgpt",
      "name": "ChatGPT Search",
      "description": "OpenAI 的搜索功能",
      "weight": 0.85
    }
  ]
}
```

---

## 环境配置

### 必需环境变量

```bash
# Bright Data (必需)
BRIGHT_DATA_API_KEY=your_api_key_here
BRIGHT_DATA_ZONE=serp_api
BRIGHT_DATA_WEB_UNLOCKER_ZONE=web_unlocker

# 豆包 LLM (必需)
ARK_API_KEY=your_ark_api_key
ARK_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
ARK_MODEL=your-model-name
```

### 可选配置

```bash
# 数据库 (默认 SQLite)
DATABASE_URL=postgresql://user:pass@localhost/peanut

# 服务器配置
PORT=8080
GIN_MODE=release
```

---

## 开发指南

### 本地运行

```bash
# 1. 安装依赖
cd web && npm install

# 2. 启动后端
make dev

# 3. 启动前端 (新终端)
cd web && npm run dev
```

### 添加新 Agent

1. 在 `internal/agent/geo/flow/agents/` 创建新 Agent 文件
2. 实现 `NewXXXAgent` 函数，返回 `compose.Graph`
3. 在 `internal/agent/geo/flow/consts.go` 添加 Agent 名称常量
4. 在 `internal/agent/geo/flow/builder.go` 注册 Agent 节点和分支

---

## 参考资源

### 框架文档
- [Eino 官方文档](https://cloudwego.io/zh/docs/eino/overview/)
- [Eino GitHub](https://github.com/cloudwego/eino)
- [Eino 示例集合](https://github.com/cloudwego/eino-examples)

### GEO 相关资源
- [GEO (Generative Engine Optimization) 介绍](https://www.searchenginejournal.com/generative-engine-optimization/)
- [Perplexity 官方博客](https://blog.perplexity.ai/)
- [Google AI Overviews 指南](https://developers.google.com/search/docs/appearance/ai-overview)

### 技术文档
- [Bright Data API 文档](https://docs.brightdata.com/)
- [豆包 API 文档](https://www.volcengine.com/docs/82379)

---

## 许可证

本文档遵循项目许可证。
