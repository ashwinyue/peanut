# AI视频创作工具对话式Agent设计方案

> **版本**: v2.0
> **更新日期**: 2026-02-02
> **核心变化**: 从传统 CRUD+AI 升级为真正的 Eino Agent 架构

---

## 一、项目背景与现状分析

### 1.1 现状分析：传统 CRUD + AI 架构

**重要发现**：当前项目本质上是 **传统 CRUD + AI API 调用**，而不是真正的 Agent。

```
┌─────────────────────────────────────────────────────────────┐
│  当前项目架构                                                 │
├─────────────────────────────────────────────────────────────┤
│  本质 = 传统 CRUD + AI API 调用                            │
│                                                             │
│  智能程度：                                                  │
│  ❌ 无自主决策 - 只能被动响应用户操作                         │
│  ❌ 无任务规划 - 用户必须手动规划每个步骤                      │
│  ❌ 无流程编排 - 各个服务独立运行，无协调                      │
│  ❌ 无错误恢复 - AI失败即报错，需手动重试                     │
│  ❌ 无上下文理解 - 每次调用都是独立的                         │
└─────────────────────────────────────────────────────────────┘
```

**典型代码模式**：

```go
// 当前代码：固定流程，无智能决策
func (s *ScriptGenerationService) GenerateCharacters(req *GenerateCharactersRequest) {
    // 1. 创建任务
    task, _ := s.taskService.CreateTask("character_generation", req.DramaID)

    // 2. 异步处理
    go s.processCharacterGeneration(task.ID, req)

    // 3. 返回任务ID
    return task.ID, nil
}

// 后台处理：固定的 AI 调用流程
func processCharacterGeneration(taskID string, req *GenerateCharactersRequest) {
    // 调用 AI → 解析 JSON → 保存数据库
    // ❌ 如果失败，直接报错
    // ❌ 无重试，无决策，无调整
}
```

**详细分析**：详见 [CURRENT_ARCHITECTURE_ANALYSIS.md](./CURRENT_ARCHITECTURE_ANALYSIS.md)

### 1.2 核心痛点
| 痛点 | 具体表现 | 根本原因 |
|------|---------|---------|
| 流程割裂 | 6个步骤需逐个点击，无法快速跳转 | 无流程编排能力 |
| 重复操作 | 批量修改需逐个镜头调整 | 无批量智能处理 |
| 学习成本高 | 专业界面功能复杂，新手难以上手 | 缺少自然语言交互 |
| 反馈滞后 | 修改效果无法实时预览 | 被动响应，无主动建议 |
| 错误处理 | AI 失败需手动重试 | 无自动恢复机制 |
| 上下文丢失 | 每次操作都是独立的 | 无状态管理 |

---

## 二、设计目标

### 2.1 核心目标：从传统架构升级为 Agent 架构

```
┌─────────────────────────────────────────────────────────────┐
│  升级目标：传统 CRUD + AI → Eino Agent                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  从：被动响应用户操作                                         │
│  到：主动理解目标，智能规划任务                                │
│                                                             │
│  从：固定流程，失败即停                                       │
│  到：动态调整，自动恢复                                       │
│                                                             │
│  从：用户手动协调各个步骤                                     │
│  到：Agent 自动编排流程                                       │
│                                                             │
│  从：每次操作独立，无上下文                                   │
│  到：完整记忆，理解当前进度                                    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 具体设计目标

**保留专业视频编辑能力的同时，引入 Eino Agent 作为"智能副驾"**，实现：

| 目标 | 说明 | 当前状态 | 目标状态 |
|------|------|---------|---------|
| **自然语言操控** | 用对话代替复杂操作 | ❌ 只能点击按钮 | ✅ 语音/文本指令 |
| **智能任务规划** | 自动分解复杂任务 | ❌ 用户手动规划 | ✅ Agent 自动规划 |
| **自主流程编排** | 协调各个步骤执行 | ❌ 用户手动协调 | ✅ Agent 自动编排 |
| **错误自动恢复** | 失败后自动重试 | ❌ 失败即报错 | ✅ 智能恢复 |
| **上下文感知** | 理解当前进度 | ❌ 无状态记忆 | ✅ 完整上下文 |
| **主动建议** | 检测问题并提示 | ❌ 被动响应 | ✅ 主动建议 |
| **批量操作** | 一键处理多个镜头 | ❌ 逐个操作 | ✅ 批量智能处理 |
| **人机协作** | 关键决策人工确认 | ❌ 无决策点 | ✅ HITL 确认 |

### 2.3 Eino Agent 核心能力

```
┌─────────────────────────────────────────────────────────────┐
│                  Eino Agent 核心能力                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. 理解能力 (Understanding)                                │
│     └─ 理解用户目标："制作3集古风短剧"                     │
│     └─ 理解上下文："角色已生成，下一步生成分镜"              │
│     └─ 理解约束条件："预算有限，用便宜的模型"                │
│                                                             │
│  2. 规划能力 (Planning)                                     │
│     └─ 自动分解任务：目标 → 步骤序列                         │
│     └─ 动态调整：根据执行结果修改计划                        │
│     └─ 优先级排序：先做重要的，后做可选的                    │
│                                                             │
│  3. 执行能力 (Execution)                                    │
│     └─ 协调 Tool 调用：按顺序执行各个步骤                    │
│     └─ 并行处理：独立的任务同时执行                          │
│     └─ 错误恢复：失败后自动重试或调整策略                    │
│                                                             │
│  4. 记忆能力 (Memory)                                       │
│     └─ 短期记忆：当前会话的上下文                            │
│     └─ 长期记忆：用户偏好和历史记录                          │
│     └─ 状态管理：跟踪每个任务的进度                          │
│                                                             │
│  5. 建议能力 (Suggestion)                                   │
│     └─ 主动检测：发现潜在问题                                │
│     └─ 智能建议：提供优化方案                                │
│     └─ 学习优化：根据反馈改进建议                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 三、Eino Agent 核心架构

### 3.1 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Eino Agent 整体架构                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐     ┌──────────────┐     ┌────────────┐ │
│  │   用户界面    │────▶│   Agent      │────▶│   Tool     │ │
│  │  (Vue3 UI)   │     │  (智能体)     │     │  (工具集)   │ │
│  └──────────────┘     └──────────────┘     └────────────┘ │
│         │                    │                    │        │
│         │                    │                    ▼        │
│         │                    │           ┌──────────────┐  │
│         │                    │           │   Service    │  │
│         │                    │           │  (业务服务)   │  │
│         │                    │           └──────────────┘  │
│         │                    │                    │        │
│         │                    │                    ▼        │
│         │                    ▼           ┌──────────────┐  │
│         │           ┌──────────────┐    │  Database    │  │
│         │           │   Memory     │    │   AI API     │  │
│         │           │  (记忆管理)   │    │   Storage    │  │
│         │           └──────────────┘    └──────────────┘  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 核心组件

#### 3.2.1 Agent（智能体）

```go
// Agent 是整个系统的核心，负责理解和决策
type DramaAgent struct {
    // 核心组件
    llm     eino.Component  // 大语言模型
    memory  eino.Component  // 记忆管理
    planner eino.Component  // 任务规划器

    // 工具集
    tools   *ToolRegistry   // 所有可用工具

    // 状态管理
    state   *AgentState     // 当前状态
    context *AgentContext   // 执行上下文
}

// Agent 的核心执行流程
func (a *DramaAgent) Execute(query string) (*AgentResult, error) {
    // 1. 理解用户意图
    intent := a.ParseIntent(query)

    // 2. 检索相关上下文
    ctx := a.memory.Retrieve(intent)

    // 3. 规划任务
    plan := a.planner.Plan(intent, ctx)

    // 4. 执行计划
    result := a.ExecutePlan(plan)

    // 5. 更新记忆
    a.memory.Store(result)

    // 6. 返回结果
    return result, nil
}
```

#### 3.2.2 Tool（工具）

```go
// Tool 是 Agent 可调用的能力封装
type Tool interface {
    // 工具名称
    Name() string

    // 工具描述（用于 LLM 理解）
    Description() string

    // 输入参数定义
    InputSchema() *Schema

    // 执行工具
    Execute(ctx context.Context, input *ToolInput) (*ToolOutput, error)
}

// 示例：生成角色工具
type GenerateCharactersTool struct {
    service *ScriptGenerationService
}

func (t *GenerateCharactersTool) Name() string {
    return "generate_characters"
}

func (t *GenerateCharactersTool) Description() string {
    return "根据剧本大纲生成角色列表，包括姓名、角色定位、描述、性格等"
}

func (t *GenerateCharactersTool) InputSchema() *Schema {
    return &Schema{
        Type: "object",
        Properties: map[string]*Property{
            "drama_id":   {Type: "string", Description: "剧本ID"},
            "count":      {Type: "integer", Description: "生成数量"},
            "outline":    {Type: "string", Description: "剧本大纲"},
        },
        Required: []string{"drama_id", "outline"},
    }
}

func (t *GenerateCharactersTool) Execute(ctx context.Context, input *ToolInput) (*ToolOutput, error) {
    // 调用底层 Service
    taskID, err := t.service.GenerateCharacters(input.ToRequest())
    if err != nil {
        return nil, err
    }

    // 等待任务完成（带重试）
    result := t.service.WaitForTask(taskID)

    return &ToolOutput{
        Success: result.Success,
        Data:    result.Data,
        Message: fmt.Sprintf("成功生成 %d 个角色", len(result.Data)),
    }, nil
}
```

#### 3.2.3 Memory（记忆）

```go
// Memory 管理上下文和历史记录
type Memory interface {
    // 存储信息
    Store(key string, value interface{})

    // 检索信息
    Retrieve(key string) (interface{}, bool)

    // 搜索相关历史
    Search(query string) []MemoryItem

    // 更新状态
    UpdateState(state *AgentState)
}

// 示例：短期记忆（会话级别）
type ShortTermMemory struct {
    messages   []Message         // 对话历史
    tasks      map[string]*Task  // 任务状态
    context    map[string]interface{} // 上下文数据
}

// 示例：长期记忆（跨会话）
type LongTermMemory struct {
    userPrefs  *UserPreferences  // 用户偏好
    history    []Session         // 历史会话
    patterns   []Pattern         // 识别的模式
}
```

#### 3.2.4 Planner（规划器）

```go
// Planner 负责任务规划和动态调整
type Planner interface {
    // 规划任务
    Plan(intent *Intent, context *Context) *Plan

    // 调整计划
    Adjust(plan *Plan, feedback *Feedback) *Plan

    // 验证计划
    Validate(plan *Plan) error
}

// 示例：任务分解
type DramaPlanner struct {
    llm    eino.Component
    tools  *ToolRegistry
}

func (p *DramaPlanner) Plan(intent *Intent, ctx *Context) *Plan {
    // 1. 理解目标
    goal := p.parseGoal(intent)  // "制作3集古风短剧"

    // 2. 分解任务
    steps := p.decompose(goal)  // ["生成剧本", "生成角色", ...]

    // 3. 排序和优化
    steps = p.optimize(steps, ctx)

    // 4. 构建计划
    return &Plan{
        Goal:  goal,
        Steps: steps,
        State: "pending",
    }
}
```

### 3.3 架构分层

```
┌─────────────────────────────────────────────────────────────┐
│                      架构分层设计                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  表示层 (Presentation)                               │    │
│  │  - Vue3 UI（主界面 + Agent 对话区）                   │    │
│  │  - 实时消息推送（WebSocket）                          │    │
│  └─────────────────────────────────────────────────────┘    │
│                          ↓                                   │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  Agent 层 (智能体层)                                  │    │
│  │  - DramaAgent（主智能体）                             │    │
│  │  - 子Agent（CharacterAgent, StoryboardAgent...）     │    │
│  └─────────────────────────────────────────────────────┘    │
│                          ↓                                   │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  Tool 层 (工具层)                                    │    │
│  │  - DramaTool（剧本操作）                             │    │
│  │  - CharacterTool（角色操作）                         │    │
│  │  - StoryboardTool（分镜操作）                        │    │
│  │  - ImageTool（图片生成）                             │    │
│  │  - VideoTool（视频生成）                             │    │
│  └─────────────────────────────────────────────────────┘    │
│                          ↓                                   │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  Service 层 (业务服务层) - 现有代码                   │    │
│  │  - DramaService                                      │    │
│  │  - ScriptGenerationService                          │    │
│  │  - StoryboardService                                │    │
│  │  - ImageGenerationService                           │    │
│  │  - VideoGenerationService                           │    │
│  └─────────────────────────────────────────────────────┘    │
│                          ↓                                   │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  基础设施层 (Infrastructure)                         │    │
│  │  - Database                                          │    │
│  │  - AI API                                            │    │
│  │  - Storage                                           │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 3.4 核心流程

```
┌─────────────────────────────────────────────────────────────┐
│                   Agent 核心执行流程                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  用户输入 "制作3集古风短剧"                                   │
│     ↓                                                       │
│  ┌─────────────────────────────────────────────┐           │
│  │  1. 理解意图 (Understanding)                │           │
│  │     目标: 制作短剧                            │           │
│  │     约束: 3集，古风                           │           │
│  └─────────────────────────────────────────────┘           │
│     ↓                                                       │
│  ┌─────────────────────────────────────────────┐           │
│  │  2. 检索上下文 (Context Retrieval)           │           │
│  │     检查: 是否已有剧本？角色？章节？            │           │
│  │     加载: 相关历史记录和用户偏好                │           │
│  └─────────────────────────────────────────────┘           │
│     ↓                                                       │
│  ┌─────────────────────────────────────────────┐           │
│  │  3. 任务规划 (Planning)                      │           │
│  │     分解:                                    │           │
│  │     1. 创建剧本                               │           │
│  │     2. 生成角色（3-5个）                      │           │
│  │     3. 为每集生成章节                          │           │
│  │     4. 为每集生成分镜                          │           │
│  │     5. 生成图片                                │           │
│  │     6. 生成视频                                │           │
│  └─────────────────────────────────────────────┘           │
│     ↓                                                       │
│  ┌─────────────────────────────────────────────┐           │
│  │  4. 执行计划 (Execution)                     │           │
│  │     for step in plan.steps:                  │           │
│  │       tool = SelectTool(step)                │           │
│  │       result = tool.Execute(step.input)      │           │
│  │                                              │           │
│  │       if result.Error:                       │           │
│  │         if CanRetry(step):                   │           │
│  │           RetryStep(step)                    │           │
│  │         else:                                │           │
│  │           SkipStep(step)                     │           │
│  │           AdjustPlan(plan)                   │           │
│  └─────────────────────────────────────────────┘           │
│     ↓                                                       │
│  ┌─────────────────────────────────────────────┐           │
│  │  5. 更新记忆 (Memory Update)                 │           │
│  │     保存: 执行结果和当前状态                  │           │
│  └─────────────────────────────────────────────┘           │
│     ↓                                                       │
│  ┌─────────────────────────────────────────────┐           │
│  │  6. 主动建议 (Proactive Suggestion)          │           │
│  │     检测: 是否有问题或优化空间？              │           │
│  │     提示: "建议添加背景音乐"                  │           │
│  └─────────────────────────────────────────────┘           │
│     ↓                                                       │
│  返回结果给用户                                              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 四、现有架构迁移到 Eino Agent

### 4.1 Service → Tool 迁移策略

详见：📘 [SERVICE_TO_TOOL_MIGRATION_PATTERN.md](./service-to-tool-migration-pattern.md)

**核心原则**：
1. **保留现有 Service**：作为 Tool 的底层实现
2. **Tool 封装 Service**：提供 Agent 可调用的接口
3. **渐进式迁移**：先迁移核心功能，再迁移边缘功能
4. **双向兼容**：UI 可以继续调用 Service，也可以调用 Tool

**迁移示例**：

```go
// 现有 Service（保持不变）
type ScriptGenerationService struct {
    db  *gorm.DB
    log *logger.Logger
    // ...
}

// 新增 Tool（封装 Service）
type GenerateCharactersTool struct {
    service *ScriptGenerationService
}

func NewGenerateCharactersTool(service *ScriptGenerationService) *GenerateCharactersTool {
    return &GenerateCharactersTool{service: service}
}

func (t *GenerateCharactersTool) Execute(ctx context.Context, input *ToolInput) (*ToolOutput, error) {
    // 1. 转换输入
    req := &GenerateCharactersRequest{
        DramaID: input.GetString("drama_id"),
        Count:   input.GetInt("count"),
        Outline: input.GetString("outline"),
    }

    // 2. 调用 Service
    taskID, err := t.service.GenerateCharacters(req)
    if err != nil {
        return nil, fmt.Errorf("生成角色失败: %w", err)
    }

    // 3. 等待完成（带超时和重试）
    result, err := t.waitForTask(taskID, 5*time.Minute)
    if err != nil {
        return nil, fmt.Errorf("等待任务完成失败: %w", err)
    }

    // 4. 返回结构化结果
    return &ToolOutput{
        Success: true,
        Data:    result,
        Message: fmt.Sprintf("成功生成 %d 个角色", len(result.Characters)),
    }, nil
}

func (t *GenerateCharactersTool) waitForTask(taskID string, timeout time.Duration) (*CharactersResult, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return nil, errors.New("任务超时")
        case <-ticker.C:
            task, err := t.service.GetTask(taskID)
            if err != nil {
                continue
            }

            if task.Status == "completed" {
                return task.Result, nil
            }

            if task.Status == "failed" {
                return nil, fmt.Errorf("任务失败: %s", task.Error)
            }
        }
    }
}
```

### 4.2 Tool 注册表

```go
// ToolRegistry 管理所有可用工具
type ToolRegistry struct {
    tools map[string]Tool
}

func NewToolRegistry() *ToolRegistry {
    return &ToolRegistry{
        tools: make(map[string]Tool),
    }
}

func (r *ToolRegistry) Register(tool Tool) {
    r.tools[tool.Name()] = tool
}

func (r *ToolRegistry) Get(name string) (Tool, bool) {
    tool, ok := r.tools[name]
    return tool, ok
}

func (r *ToolRegistry) List() []Tool {
    tools := make([]Tool, 0, len(r.tools))
    for _, tool := range r.tools {
        tools = append(tools, tool)
    }
    return tools
}

// 初始化所有工具
func InitToolRegistry(services *Services) *ToolRegistry {
    registry := NewToolRegistry()

    // 剧本相关工具
    registry.Register(NewCreateDramaTool(services.Drama))
    registry.Register(NewUpdateDramaTool(services.Drama))
    registry.Register(NewGetDramaTool(services.Drama))

    // 角色相关工具
    registry.Register(NewGenerateCharactersTool(services.ScriptGeneration))
    registry.Register(NewGetCharactersTool(services.Drama))

    // 分镜相关工具
    registry.Register(NewGenerateStoryboardTool(services.Storyboard))
    registry.Register(NewUpdateStoryboardTool(services.Storyboard))

    // 图片相关工具
    registry.Register(NewGenerateImageTool(services.ImageGeneration))
    registry.Register(NewGetImageTool(services.ImageGeneration))

    // 视频相关工具
    registry.Register(NewGenerateVideoTool(services.VideoGeneration))
    registry.Register(NewMergeVideoTool(services.VideoMerge))

    return registry
}
```

### 4.3 渐进式迁移路径

```
Phase 1: 基础设施搭建
  ├─ 集成 Eino 框架
  ├─ 实现 DramaAgent 基础类
  ├─ 实现Tool封装层
  └─ 实现Memory管理

Phase 2: 核心功能迁移
  ├─ 剧本管理（CRUD）
  ├─ 角色生成
  ├─ 分镜生成
  └─ 状态查询

Phase 3: 生成功能迁移
  ├─ 图片生成
  ├─ 视频生成
  └─ 视频合成

Phase 4: 高级能力实现
  ├─ 任务规划器
  ├─ 错误恢复
  ├─ 主动建议
  └─ 学习优化

Phase 5: 用户体验优化
  ├─ 对话界面
  ├─ 进度可视化
  ├─ 实时反馈
  └─ 快捷操作
```

---

## 五、界面设计：Cursor模式

### 3.1 界面布局

```
┌─────────────────────────────────────────────────────────────┐
│  顶部导航栏（剧本/角色/场景/分镜/制作/作品）          积分:190 │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────┐  ┌──────────────────┐  ┌───────────────┐     │
│  │ 分镜列表  │  │    媒体预览       │  │   参数设置     │     │  ← 主界面
│  │          │  │    [视频预览区]    │  │  （完全保留）  │     │    完全
│  │ 镜头1    │  │                  │  │               │     │    保留
│  │ 镜头2    │  │   [时间轴编辑器]  │  │               │     │
│  │ ...      │  │                  │  │               │     │
│  └──────────┘  └──────────────────┘  └───────────────┘     │
├─────────────────────────────────────────────────────────────┤
│  🤖 Agent侧边栏（可折叠）                                    │  ← 新增
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │    对话
│                                                             │    区域
│  [对话流：用户指令 + Agent回复卡片]                           │
│                                                             │
│  [快捷指令] [历史] [帮助]  输入框...              [发送]    │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 设计原则
- **主界面不变**：保留原有专业视频编辑功能，老用户零迁移成本
- **侧边栏增强**：新增Agent对话区作为操作加速器
- **双向联动**：对话提及内容 ↔ 主界面高亮/跳转

---

## 四、交互设计

### 4.1 消息卡片类型

| 卡片类型 | 用途 | 示例 |
|---------|------|------|
| **用户指令** | 显示用户输入 | "把镜头2改成侧脸，时长4秒" |
| **确认结果** | 操作成功反馈 | ✓ 已修改镜头2<br>时长: 3s→4s<br>[撤销] [预览] |
| **进度卡片** | 批量任务进度 | 🎨 生成中...<br>镜头1 ████░░ 40%<br>[取消] [后台运行] |
| **预览卡片** | 效果确认 | [缩略图]<br>符合预期吗？<br>[确认] [重生成] [编辑] |
| **主动建议** | Agent检测到问题 | 💡 建议添加0.5秒过渡<br>[添加] [忽略] |
| **异常警告** | 错误/冲突提示 | ⚠️ 镜头3缺少角色绑定<br>[一键修复] |

### 4.2 典型交互流程

**场景1：批量修改**
```
用户：所有德古拉镜头改成冷色调，时长+2秒
Agent：✓ 找到4个镜头（1,2,5,7）
       正在批量修改...
       [预览效果] [确认应用] [撤销]
用户：确认
Agent：✓ 已修改4个镜头，主界面自动刷新
```

**场景2：智能生成**
```
用户：生成德古拉睁眼特写，要恐怖一点
Agent：理解：德古拉 | 睁眼 | 特写 | 恐怖风格
       生成提示词："眼睑缓慢移动...深渊瞳孔..."
       [直接生成] [调整风格] [手动编辑]
用户：直接生成
Agent：🎨 生成中... → ✅ 完成！已插入为镜头2b
```

**场景3：主动建议**
```
Agent：💡 检测到镜头3-4节奏过快，建议添加过渡
用户：添加
Agent：✓ 已添加黑场过渡，主界面时间轴更新
```

### 4.3 快捷操作栏
```
┌─────────────────────────────────────────┐
│  [最近] [生成] [修改] [查询] [帮助]      │  ← 一键触发常用指令
│                                         │
│  输入指令或按 ⌘K 唤醒...        [语音🎤] │
└─────────────────────────────────────────┘
```

---

## 五、功能映射

| 原操作 | Agent自然语言指令 | 自动化程度 |
|-------|------------------|-----------|
| 修改镜头时长 | "第2个镜头改成5秒" | 全自动 |
| 批量生成图片 | "前5个镜头生成关键帧" | 全自动 |
| 调整提示词 | "让德古拉更凶狠" | AI辅助改写 |
| 视频合成 | "按当前顺序合成预览" | 全自动 |
| 角色替换 | "把德牧换成黑猫" | 半自动（需确认） |
| 节奏优化 | "整体加快20%" | AI建议+确认 |
| 添加音效 | "给窗户异响配雷声" | 全自动（检索素材） |

---

## 六、联动规则

| 对话中提及 | 主界面反应 |
|-----------|-----------|
| "镜头2" | 自动选中，高亮边框 |
| "时间轴第3秒" | 播放头跳转00:03 |
| "预览" | 预览区开始播放 |
| "对比" | 分屏显示修改前后 |
| "全屏" | 侧边栏折叠，主界面最大化 |

---

## 七、唤醒方式

| 方式 | 操作 | 场景 |
|------|------|------|
| 快捷键 | `⌘K` / `Ctrl+K` | 全局快速唤醒 |
| 选中触发 | 选中镜头自动弹出建议 | 上下文感知 |
| 右键菜单 | 右键镜头 → "用Agent修改" | 精确操作 |
| 语音输入 | 长按空格说话 | 快速指令 |

---

## 八、实施路径

### Phase 1：指令层（MVP）
- 添加Agent侧边栏，解析自然语言触发原有功能
- 支持：生成、修改时长、批量操作、进度展示

### Phase 2：智能层
- 上下文感知（分析镜头关系、节奏）
- 主动建议（检测到问题时提示）

### Phase 3：自动化层
- "一键成片"：输入剧本，Agent自动完成全流程
- 人工介入点：风格确认、重点镜头审核

---

## 九、设计优势

| 维度 | 效果 |
|------|------|
| 学习成本 | 自然语言替代复杂界面，新手友好 |
| 操作效率 | 一句话完成批量操作 |
| 专业保留 | 主界面完全不变，兼容老用户 |
| 渐进增强 | 新手用Agent，老手用快捷键 |
| 错误回退 | 对话历史=操作历史，随时撤销 |
| 创作流畅 | 无需页面切换，对话流持续上下文 |

---

## 十、关键设计决策

1. **不替换原有界面**：Agent作为增强层，而非替代层
2. **消息驱动交互**：用对话卡片承载状态，而非面板切换
3. **双向实时联动**：对话与主界面双向同步，操作即所见
4. **Human-in-the-Loop**：关键决策必须人工确认，AI辅助而非取代
5. **渐进式智能**：从被动响应到主动建议，逐步提升自动化程度