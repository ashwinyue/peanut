# Service to Tool Migration Pattern

**Extracted:** 2025-02-01
**Context:** 从传统 MVC 架构迁移到 Agent 架构时

## Problem

现有项目使用传统 MVC 架构（Handler → Service → Repository），想迁移到 Agent 架构但不知道如何映射现有代码。

## Solution

使用 **Service → Tool 映射模式**，将现有服务转换为 Agent 可调用的工具：

### 核心映射关系

| 传统架构组件 | Agent 架构组件 | 说明 |
|------------|---------------|------|
| **Handler** | Handler (保留) | HTTP 请求处理层保持不变 |
| **Service** | **Tool** | 业务逻辑变为 Agent 可调用的工具 |
| **手动编排** | **Agent** | Agent 自动规划和编排工具调用 |
| **AsyncTask + 轮询** | **SSE 实时推送** | 进度通知改为流式事件 |
| **Model** | Model (保留) | 数据模型层保持不变 |

### 实施步骤

#### Step 1: 识别原子服务（Atomic Services）

找出哪些服务是独立的、可复用的原子操作：

```go
// ✅ 好的 Tool 候选（原子操作）
- GenerateScript()      // 生成剧本
- GenerateStoryboard()  // 生成分镜
- GenerateImage()       // 生成单张图片
- GenerateVideo()       // 生成单个视频
- MergeVideo()          // 合成视频

// ❌ 不适合作为 Tool（复杂编排）
- ProduceDrama()        // 整个短剧生产流程
- FinalizeEpisode()     // 章节完成（包含多个子操作）
```

#### Step 2: 定义 Tool 接口

```go
// 通用 Tool 接口
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, input any) (any, error)
}

// 示例：生成剧本 Tool
type GenerateScriptTool struct {
    aiService   *AIService
    dramaStore  *DramaStore
}

func (t *GenerateScriptTool) Name() string {
    return "generate_script"
}

func (t *GenerateScriptTool) Description() string {
    return "生成短剧剧本大纲和角色设定"
}

func (t *GenerateScriptTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
    // 1. 解析输入
    title := input["title"].(string)
    genre := input["genre"].(string)
    episodes := input["total_episodes"].(int)

    // 2. 调用原有服务逻辑
    drama, err := t.aiService.GenerateDrama(ctx, title, genre, episodes)
    if err != nil {
        return nil, err
    }

    // 3. 保存到数据库
    err = t.dramaStore.Create(ctx, drama)
    if err != nil {
        return nil, err
    }

    // 4. 返回结果
    return map[string]any{
        "drama_id":   drama.ID,
        "characters": drama.Characters,
        "episodes":   drama.Episodes,
    }, nil
}
```

#### Step 3: 注册到 Agent

```go
// Eino 框架示例
agent := &Agent{
    Name:         "drama-production",
    SystemPrompt: DramaProductionSystemPrompt,
    Tools: []Tool{
        &GenerateScriptTool{},
        &GenerateStoryboardTool{},
        &GenerateImagesTool{},
        &GenerateVideosTool{},
        &MergeVideoTool{},
    },
}
```

#### Step 4: 保留必要的同步 API

不是所有功能都需要通过 Agent，保留一些直接的 REST API：

```go
// 保留这些 API（查询类操作）
GET    /api/v1/dramas           # 剧本列表
GET    /api/v1/dramas/:id       # 剧本详情
GET    /api/v1/episodes/:id     # 章节详情

// 新增 Agent 对话 API（复杂任务）
POST   /api/v1/chat             # Agent 对话入口
WS     /api/v1/chat/ws          # WebSocket 实时通信
```

### 架构对比

#### 迁移前：传统架构

```go
// 手动编排
func (s *DramaService) ProduceDrama(ctx context.Context, req ProduceRequest) error {
    // 1. 生成剧本
    drama, err := s.scriptService.Generate(ctx, req)
    // 2. 提取场景
    scenes, err := s.sceneService.Extract(ctx, drama.Episodes[0].ID)
    // 3. 生成分镜
    storyboards, err := s.storyboardService.Generate(ctx, drama.Episodes[0].ID)
    // 4. 生成图片（循环调用）
    for _, sb := range storyboards {
        image := s.imageService.Generate(ctx, sb)
    }
    // 5. 生成视频
    for _, img := range images {
        video := s.videoService.Generate(ctx, img)
    }
    // 6. 合成视频
    s.mergeService.Merge(ctx, episodeID)
}

// 用户需要多次调用
POST /api/v1/dramas           → 创建剧本
POST /api/v1/episodes/:id/props/extract  → 提取道具
POST /api/v1/episodes/:id/storyboards    → 生成分镜
POST /api/v1/episodes/:id/images/generate → 生成图片
POST /api/v1/episodes/:id/videos/generate → 生成视频
POST /api/v1/episodes/:id/finalize       → 合成视频
```

#### 迁移后：Agent 架构

```go
// Agent 自动规划
User: "帮我创作一个3集古风短剧"
  ↓
Agent: 自动调用 Tools
  1. generate_script
  2. extract_scenes
  3. generate_storyboard
  4. generate_images (并发生成)
  5. generate_videos (并发生成)
  6. merge_video

// 用户一次调用，SSE 实时推送进度
POST /api/v1/chat
{
  "query": "帮我创作一个3集古风短剧"
}

// SSE 流式返回
data: {"response_type":"tool_call","tool_name":"generate_script"}
data: {"response_type":"tool_result","tool_name":"generate_script","content":"已生成剧本"}
data: {"response_type":"tool_call","tool_name":"generate_images","content":"生成中（5/12）"}
...
data: {"response_type":"answer","content":"短剧制作完成！"}
```

### 目录结构变化

```
迁移前：
├── api/
│   └── handlers/           # HTTP 处理器
├── application/
│   └── services/           # 业务服务（21个文件）
├── domain/
│   └── models/             # 数据模型
└── infrastructure/
    └── database/           # 数据访问

迁移后：
├── internal/
│   ├── handler/
│   │   └── http/           # HTTP 处理器（保留）
│   ├── biz/
│   │   └── agent/
│   │       ├── agent.go    # Agent 定义
│   │       └── tools/      # Tools（8个，原 services 转换）
│   ├── store/              # 数据访问（新增 Repository 层）
│   └── model/              # 数据模型（保留）
```

## Example

### 短剧项目迁移示例

#### 原有 Service

```go
// application/services/image_generation_service.go (2000+ 行)
type ImageGenerationService struct {
    db          *gorm.DB
    aiService   *AIService
    taskService *TaskService
    // ... 复杂的依赖
}

func (s *ImageGenerationService) GenerateImage(episodeID string) error {
    // 复杂的生成逻辑
    // ...
}
```

#### 转换为 Tool

```go
// internal/biz/agent/tools/image.go (200 行)
type GenerateImagesTool struct {
    aiClient  ai.AIClient
    store     *store.ImageGenerationStore
}

func (t *GenerateImagesTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
    episodeID := input["episode_id"].(uint)
    limit := input["concurrent_limit"].(int)

    // 获取分镜
    storyboards, err := t.store.GetStoryboards(ctx, episodeID)
    if err != nil {
        return nil, err
    }

    // 并发生成
    results := make([]ImageResult, 0)
    sem := make(chan struct{}, limit) // 限制并发

    for _, sb := range storyboards {
        sem <- struct{}{}
        go func(sb Storyboard) {
            defer func() { <-sem }()

            // 调用 AI 生成
            url, err := t.aiClient.GenerateImage(ctx, sb.ImagePrompt)
            results = append(results, ImageResult{StoryboardID: sb.ID, URL: url})
        }(sb)
    }

    // 等待全部完成
    for i := 0; i < cap(sem); i++ {
        sem <- struct{}{}
    }

    return map[string]any{
        "generated": results,
        "total":    len(results),
    }, nil
}
```

**代码减少：** 2000 行 → 200 行（90% 减少）

## When to Use

在以下情况触发此技能：
- 🔄 从传统架构迁移到 Agent 架构
- 🔄 需要将现有 Service 转换为 Tool
- 🔄 评估迁移的工作量和收益
- 🔄 设计混合架构（部分保留 Service，部分转为 Tool）

## Key Takeaways

1. **保留 Model 层**：数据模型不需要改变
2. **Service → Tool**：原子服务变为可调用的工具
3. **手动编排 → Agent**：用 Agent 自动规划和调用工具
4. **轮询 → SSE**：进度通知改为流式事件推送
5. **渐进式迁移**：可以先迁移部分功能，保留原有 API

## Additional Resources

- CloudWeGo Eino 文档：https://www.cloudwego.io/zh/docs/eino/
- Next-Show 项目：docs/next-show/ (完整的 Handler → Biz → Store 架构示例)
- 迁移方案文档：docs/NEXT_SHOW_MIGRATION.md
