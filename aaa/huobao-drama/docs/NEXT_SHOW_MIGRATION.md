# Next-Show 单 Agent 架构设计方案

## 1. 架构分层图

```
┌─────────────────────────────────────────────────────────────┐
│                        Web 前端                              │
│                    Vue 3 + Element Plus                      │
└─────────────────────────────────────────────────────────────┘
                              │ HTTP/SSE
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Handler 层 (Controller)                   │
│  internal/handler/http/                                     │
│  ├── drama.go           # 剧本管理 HTTP                      │
│  ├── episode.go         # 章节管理 HTTP                      │
│  ├── storyboard.go      # 分镜管理 HTTP                      │
│  ├── generation.go      # AI 生成 HTTP                       │
│  └── chat.go            # Agent 对话 HTTP (SSE 流式)          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Biz 层 (Service)                         │
│  internal/biz/                                              │
│  ├── drama/              # 剧本业务                          │
│  ├── agent/              # Agent 业务 (核心)                 │
│  │   ├── agent.go        # DramaProduction Agent 定义        │
│  │   ├── tools/          # 自定义 Tools                      │
│  │   │   ├── script.go           # 生成剧本                  │
│  │   │   ├── storyboard.go        # 生成分镜                 │
│  │   │   ├── image.go             # 生成图片                 │
│  │   │   ├── video.go             # 生成视频                 │
│  │   │   ├── merge.go             # 合成视频                 │
│  │   │   ├── character.go         # 提取角色                 │
│  │   │   └── scene.go             # 提取场景                 │
│  │   └── runtime.go      # Agent 运行时                      │
│  └── session/            # 会话管理                          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Store 层 (Repository)                     │
│  internal/store/                                             │
│  ├── drama.go           # 剧本 CRUD                          │
│  ├── episode.go         # 章节 CRUD                          │
│  ├── storyboard.go      # 分镜 CRUD                          │
│  ├── image_generation.go # 图片生成记录                      │
│  └── video_generation.go # 视频生成记录                      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Model 层 (数据模型)                       │
│  internal/model/                                             │
│  ├── drama.go           # 剧本模型 (保留)                    │
│  ├── episode.go         # 章节模型 (保留)                    │
│  ├── storyboard.go      # 分镜模型 (保留)                    │
│  ├── character.go       # 角色模型 (保留)                    │
│  └── ai_config.go       # AI 配置 (保留)                     │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. 单 Agent 设计

### Agent 类型：Plan-Execute

**Agent 配置：**

```go
const DramaProductionSystemPrompt = `你是短剧制作专家（Drama Production Agent）。你的职责是：

### 角色
- 理解用户的短剧创作需求
- 将短剧制作任务分解为清晰的执行步骤
- 按步骤执行，跟踪进度，动态调整

### 工作流程

#### 1. 规划阶段（Plan）
- 分析用户需求（题材、集数、风格等）
- 将任务分解为：
  1. 生成剧本大纲和角色
  2. 为每集提取场景
  3. 为每集生成分镜
  4. 为每集生成图片
  5. 为每集生成视频
  6. 合成最终视频

#### 2. 执行阶段（Execute）
- 按顺序执行每个步骤
- 调用对应的工具完成任务
- 记录每步的执行结果

#### 3. 反思阶段（Replan）
- 如果步骤失败，分析原因
- 必要时修改后续计划（如重新生成某张图片）
- 确保最终目标达成

### 输出格式
- 展示当前执行计划
- 标记已完成/进行中/待执行的步骤
- 说明每步的执行结果

### 可用工具
- generate_script: 生成剧本大纲和角色
- extract_scenes: 提取场景
- generate_storyboard: 生成分镜
- generate_images: 批量生成图片
- generate_videos: 批量生成视频
- merge_video: 合成最终视频
- thinking: 思考和规划
- todo_write: 更新任务列表

### 适用场景
- 短剧全流程制作
- 单独的剧本生成
- 单独的图片/视频生成
`
```

**Agent 配置定义：**

```go
GetDramaProductionAgent() *model.Agent {
    temp := 0.7
    return &model.Agent{
        ID:            "drama-production",
        Name:          "drama-production",
        DisplayName:   "短剧制作",
        Description:   "短剧制作专家，从剧本到视频的全流程自动化生产",
        AgentType:     model.AgentTypePlanExecute,
        AgentRole:     model.AgentRoleOrchestrator,
        SystemPrompt:  DramaProductionSystemPrompt,
        MaxIterations: 50,
        Temperature:   &temp,
        IsEnabled:     true,
        IsBuiltin:     true,
        Config: model.JSONMap{
            "allowed_tools": []string{
                "transfer_task",
                "thinking",
                "todo_write",
                "generate_script",
                "extract_scenes",
                "generate_storyboard",
                "generate_images",
                "generate_videos",
                "merge_video",
            },
            "replan_enabled": true,
            "concurrent_image_generation": 5,  // 并发生成图片数
            "concurrent_video_generation": 3,  // 并发生成视频数
        },
    }
}
```

---

## 3. 自定义 Tools 设计

### 3.1 generate_script（生成剧本）

```go
type GenerateScriptInput struct {
    Title       string   `json:"title"`
    Genre       string   `json:"genre"`
    Style       string   `json:"style"`
    TotalEpisodes int    `json:"total_episodes"`
    Description string   `json:"description"`
}

type GenerateScriptOutput struct {
    DramaID     uint   `json:"drama_id"`
    Characters  []Character `json:"characters"`
    Episodes    []Episode   `json:"episodes"`
}
```

### 3.2 extract_scenes（提取场景）

```go
type ExtractScenesInput struct {
    EpisodeID uint `json:"episode_id"`
}

type ExtractScenesOutput struct {
    Scenes    []Scene `json:"scenes"`
    Count     int     `json:"count"`
}
```

### 3.3 generate_storyboard（生成分镜）

```go
type GenerateStoryboardInput struct {
    EpisodeID uint `json:"episode_id"`
}

type GenerateStoryboardOutput struct {
    Storyboards []Storyboard `json:"storyboards"`
    Count       int          `json:"count"`
}
```

### 3.4 generate_images（批量生成图片）

```go
type GenerateImagesInput struct {
    EpisodeID       uint   `json:"episode_id"`
    ConcurrentLimit int    `json:"concurrent_limit"` // 并发数
}

type GenerateImagesOutput struct {
    Generated   []ImageGeneration `json:"generated"`
    Failed      []ImageGeneration `json:"failed"`
    TotalCount  int               `json:"total_count"`
    SuccessCount int              `json:"success_count"`
}
```

### 3.5 generate_videos（批量生成视频）

```go
type GenerateVideosInput struct {
    EpisodeID       uint `json:"episode_id"`
    ConcurrentLimit int  `json:"concurrent_limit"`
}

type GenerateVideosOutput struct {
    Generated   []VideoGeneration `json:"generated"`
    Failed      []VideoGeneration `json:"failed"`
    TotalCount  int               `json:"total_count"`
    SuccessCount int              `json:"success_count"`
}
```

### 3.6 merge_video（合成视频）

```go
type MergeVideoInput struct {
    EpisodeID uint `json:"episode_id"`
}

type MergeVideoOutput struct {
    VideoURL   string `json:"video_url"`
    Duration   int    `json:"duration"`
    Status     string `json:"status"`
}
```

---

## 4. 文件结构

```
huobao-drama-next-show/
├── cmd/
│   └── server/
│       └── main.go                    # 服务入口
├── internal/
│   ├── handler/
│   │   └── http/
│   │       ├── router.go              # 路由注册
│   │       ├── drama.go               # 剧本 API
│   │       ├── episode.go             # 章节 API
│   │       ├── storyboard.go          # 分镜 API
│   │       ├── generation.go          # 生成 API
│   │       ├── chat.go                # Agent 对话 API (SSE)
│   │       └── sse_handler.go         # SSE 处理器
│   ├── biz/
│   │   ├── drama/
│   │   │   └── drama.go               # 剧本业务
│   │   ├── agent/
│   │   │   ├── agent.go               # Agent 定义
│   │   │   ├── runtime.go             # Agent 运行时
│   │   │   └── tools/
│   │   │       ├── script.go          # 剧本工具
│   │   │       ├── storyboard.go      # 分镜工具
│   │   │       ├── image.go           # 图片工具
│   │   │       ├── video.go           # 视频工具
│   │   │       ├── merge.go           # 合成工具
│   │   │       ├── character.go       # 角色工具
│   │   │       └── scene.go           # 场景工具
│   │   └── session/
│   │       └── session.go             # 会话管理
│   ├── store/
│   │   ├── store.go                   # 存储层入口
│   │   ├── drama.go                   # 剧本 CRUD
│   │   ├── episode.go                 # 章节 CRUD
│   │   ├── storyboard.go              # 分镜 CRUD
│   │   ├── image_generation.go        # 图片生成记录
│   │   └── video_generation.go        # 视频生成记录
│   ├── model/
│   │   ├── drama.go                   # 剧本模型
│   │   ├── episode.go                 # 章节模型
│   │   ├── storyboard.go              # 分镜模型
│   │   ├── character.go               # 角色模型
│   │   ├── scene.go                   # 场景模型
│   │   └── ai_config.go               # AI 配置
│   └── pkg/
│       ├── agent/
│       │   └── event/                 # 事件系统
│       └── sse/                       # SSE 协议
├── pkg/
│   └── api/
│       └── v1/                        # API 定义
├── configs/
│   └── config.yaml                    # 配置文件
├── migrations/                        # 数据库迁移
├── web/                               # 前端（复用）
├── go.mod
└── README.md
```

---

## 5. 与现有代码的映射关系

| 现有模块 | 新架构位置 | 说明 |
|---------|-----------|------|
| `api/handlers/drama.go` | `internal/handler/http/drama.go` | HTTP 处理器 |
| `application/services/drama_service.go` | `internal/biz/drama/drama.go` | 业务逻辑 |
| `application/services/storyboard_service.go` | `internal/biz/agent/tools/storyboard.go` | 成为 Tool |
| `application/services/image_generation_service.go` | `internal/biz/agent/tools/image.go` | 成为 Tool |
| `application/services/video_generation_service.go` | `internal/biz/agent/tools/video.go` | 成为 Tool |
| `application/services/video_merge_service.go` | `internal/biz/agent/tools/merge.go` | 成为 Tool |
| `domain/models/` | `internal/model/` | 数据模型保留 |
| `pkg/ai/` | `internal/biz/agent/pkg/ai/` | AI 客户端复用 |
| `pkg/image/` | `internal/biz/agent/pkg/image/` | 图片客户端复用 |
| `pkg/video/` | `internal/biz/agent/pkg/video/` | 视频客户端复用 |

---

## 6. 核心交互流程

### 用户对话模式

```javascript
// 前端发送消息
ws.send(JSON.stringify({
  query: "帮我创作一个3集古风短剧，题材是仙侠恋"
}))

// SSE 流式返回
data: {"response_type":"agent_query","content":"开始短剧制作任务"}
data: {"response_type":"thinking","content":"分析需求：仙侠恋题材，3集..."}
data: {"response_type":"tool_call","tool_name":"todo_write","content":"创建执行计划"}
data: {"response_type":"tool_result","tool_name":"generate_script","content":"已生成剧本大纲"}
data: {"response_type":"tool_result","tool_name":"extract_scenes","content":"已提取5个场景"}
data: {"response_type":"tool_result","tool_name":"generate_storyboard","content":"第1集已生成12个分镜"}
data: {"response_type":"tool_result","tool_name":"generate_images","content":"第1集图片生成中（5/12）"}
data: {"response_type":"tool_result","tool_name":"generate_videos","content":"第1集视频生成中（3/12）"}
data: {"response_type":"answer","content":"短剧制作完成！视频已生成"}
```

---

## 7. API 变更

### 新增 Agent 对话接口

```
POST /api/v1/chat
- 接收用户自然语言输入
- 返回 SSE 流式事件

WebSocket /api/v1/chat/ws
- 实时双向通信
```

### 保留现有 REST API

```
GET    /api/v1/dramas           # 剧本列表
POST   /api/v1/dramas           # 创建剧本
GET    /api/v1/dramas/:id       # 剧本详情
...
```

---

## 8. 实施步骤

### Phase 1: 架构搭建（1-2天）
1. 初始化项目结构
2. 配置 Eino 和 Next-Show 依赖
3. 设置数据库连接和迁移
4. 复制现有 Model 层

### Phase 2: Store 层（1-2天）
1. 实现 Drama Store
2. 实现 Episode Store
3. 实现 Storyboard Store
4. 实现 Image/Video Generation Store

### Phase 3: Biz 层和 Tools（3-4天）
1. 实现 AI 客户端适配器
2. 实现 generate_script Tool
3. 实现 extract_scenes Tool
4. 实现 generate_storyboard Tool
5. 实现 generate_images Tool
6. 实现 generate_videos Tool
7. 实现 merge_video Tool

### Phase 4: Agent 和 Runtime（2-3天）
1. 定义 DramaProduction Agent
2. 实现 Agent 运行时
3. 集成 Plan-Execute 逻辑
4. 实现错误处理和 Replan

### Phase 5: Handler 层和 SSE（2天）
1. 实现 Chat Handler（SSE）
2. 保留现有 REST API
3. 实现事件处理器
4. 测试流式输出

### Phase 6: 前端适配（2-3天）
1. 添加 Agent 对话界面
2. 保留现有功能界面
3. 实现进度可视化
4. 测试完整流程

---

## 9. 预期收益

| 指标 | 当前架构 | Next-Show 单 Agent |
|------|----------|-------------------|
| 代码量 | 100% | ~60-70% |
| 服务文件数 | 21 个 services | ~8 个 tools |
| 编排复杂度 | 手动协调 | Agent 自动规划 |
| 用户体验 | 多次点击 | 自然语言对话 |
| 错误恢复 | 手动重试 | Agent 自动 replan |
| 进度可见性 | 轮询任务状态 | SSE 实时推送 |
