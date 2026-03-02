# LangGraph 架构迁移方案

> **文档日期**: 2026-02-25
> **文档类型**: 架构设计
> **目标**: 从现有异步任务模式迁移到 LangGraph Graph Agent 模式

---

## 目录

- [一、现状分析](#一现状分析)
- [二、模式对比](#二模式对比)
- [三、LangGraph 架构设计](#三langgraph-架构设计)
- [四、异步任务新形式](#四异步任务新形式)
- [五、推荐架构](#五推荐架构)
- [六、具体实现建议](#六具体实现建议)
- [七、迁移路径](#七迁移路径)

---

## 一、现状分析

### 1.1 现有异步任务模式

```
前端 → API → 创建Task → 返回task_id → 前端轮询 → 后台goroutine处理
```

**存在的问题**：

| 问题 | 说明 |
|------|------|
| ❌ 流程割裂 | 各阶段独立，缺乏整体编排 |
| ❌ 无法中途干预 | 用户无法在关键节点确认或调整 |
| ❌ 状态分散 | 状态分散在多个数据库表中 |
| ❌ 重试/恢复困难 | 任务失败后难以从断点恢复 |
| ❌ 用户体验差 | 需要不断轮询，无法实时反馈 |

### 1.2 期望的 LangGraph 模式

```
对话输入 → Graph 编排 → interrupt 暂停 → 用户确认 → continue 继续
```

**核心优势**：

| 优势 | 说明 |
|------|------|
| ✅ 有状态机 | 流程可视化，状态清晰 |
| ✅ Human-in-the-loop | 中断确认机制 |
| ✅ Checkpointing | 状态持久化，支持暂停/恢复 |
| ✅ 可重试 | 从任意节点恢复执行 |
| ✅ Streaming | 流式输出，实时反馈 |

---

## 二、模式对比

### 2.1 异步任务是否还需要？

**答案：需要，但角色变化**

| 场景 | 现有模式 | LangGraph 模式 |
|------|----------|----------------|
| LLM 生成（角色/分镜） | 异步任务 + 轮询 | **同步节点**（流式输出） |
| 图片生成 | 异步任务 + 轮询 | **异步节点**（等待回调） |
| 视频生成 | 异步任务 + 轮询 | **异步节点**（等待回调） |
| 视频合成 | 异步任务 + 轮询 | **异步节点**（等待回调） |

### 2.2 架构对比图

```
┌─────────────────────────────────────────────────────────────────────┐
│                        现有模式                                      │
├─────────────────────────────────────────────────────────────────────┤
│  前端 → API → 创建Task → 返回task_id → 前端轮询 → 后台goroutine处理  │
│                                                                      │
│  [Drama] → [Character Task] → [Episode Task] → [Storyboard Task]    │
│     ↓           ↓                ↓                ↓                 │
│   轮询       轮询              轮询             轮询                 │
│     ↓           ↓                ↓                ↓                 │
│  [Image Task] → [Video Task] → [Merge Task]                         │
│     ↓              ↓               ↓                                │
│   轮询           轮询            轮询                                │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                     LangGraph 模式                                   │
├─────────────────────────────────────────────────────────────────────┤
│  对话输入 → Graph 编排 → interrupt 暂停 → 用户确认 → continue 继续   │
│                                                                      │
│  [Parse Outline] → [Generate Characters] → ⏸ [Confirm Characters]   │
│                                                             ↓        │
│  [Confirm Final] ⏸ ← [Merge Videos] ← [Generate Videos] ← [Episodes]│
│                                                             ↓        │
│                                              [Generate Images] ⏸     │
│                                                      ↓               │
│                                              (异步等待回调)           │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 三、LangGraph 架构设计

### 3.1 状态定义

```python
from typing import TypedDict, Annotated
import operator

class DramaState(TypedDict):
    """短剧创作状态"""
    # 基本信息
    drama_id: str
    outline: str
    title: str

    # 生成内容
    characters: list[dict]
    episodes: list[dict]
    storyboards: list[dict]
    scenes: list[dict]
    props: list[dict]

    # 媒体资源
    images: dict[str, str]      # {storyboard_id: image_url}
    videos: dict[str, str]      # {storyboard_id: video_url}
    final_video: str

    # 流程控制
    current_step: str
    user_confirmations: Annotated[dict, operator.or_]
    pending_tasks: dict[str, str]  # {task_key: task_id}

    # 错误处理
    errors: list[dict]
    retry_count: dict[str, int]
```

### 3.2 Graph 构建

```python
from langgraph.graph import StateGraph, END
from langgraph.checkpoint.postgres import PostgresSaver

# 构建工作流
workflow = StateGraph(DramaState)

# ==================== 添加节点 ====================

# 阶段1: 解析与角色
workflow.add_node("parse_outline", parse_outline_node)
workflow.add_node("generate_characters", generate_characters_node)
workflow.add_node("confirm_characters", confirm_characters_node)  # interrupt

# 阶段2: 章节与分镜
workflow.add_node("generate_episodes", generate_episodes_node)
workflow.add_node("generate_storyboards", generate_storyboards_node)
workflow.add_node("confirm_storyboards", confirm_storyboards_node)  # interrupt

# 阶段3: 媒体生成
workflow.add_node("submit_image_tasks", submit_image_tasks_node)
workflow.add_node("wait_for_images", wait_for_images_node)  # interrupt
workflow.add_node("confirm_images", confirm_images_node)  # interrupt

workflow.add_node("submit_video_tasks", submit_video_tasks_node)
workflow.add_node("wait_for_videos", wait_for_videos_node)  # interrupt

# 阶段4: 合成输出
workflow.add_node("merge_videos", merge_videos_node)
workflow.add_node("confirm_final", confirm_final_node)  # interrupt

# ==================== 定义边 ====================

workflow.set_entry_point("parse_outline")

# 阶段1
workflow.add_edge("parse_outline", "generate_characters")
workflow.add_edge("generate_characters", "confirm_characters")

# 阶段2
workflow.add_conditional_edges(
    "confirm_characters",
    lambda state: "proceed" if state["user_confirmations"].get("characters") else "regenerate",
    {
        "proceed": "generate_episodes",
        "regenerate": "generate_characters"
    }
)
workflow.add_edge("generate_episodes", "generate_storyboards")
workflow.add_edge("generate_storyboards", "confirm_storyboards")

# 阶段3
workflow.add_conditional_edges(
    "confirm_storyboards",
    lambda state: "proceed" if state["user_confirmations"].get("storyboards") else "regenerate",
    {
        "proceed": "submit_image_tasks",
        "regenerate": "generate_storyboards"
    }
)
workflow.add_edge("submit_image_tasks", "wait_for_images")
workflow.add_edge("wait_for_images", "confirm_images")
workflow.add_edge("confirm_images", "submit_video_tasks")
workflow.add_edge("submit_video_tasks", "wait_for_videos")

# 阶段4
workflow.add_edge("wait_for_videos", "merge_videos")
workflow.add_edge("merge_videos", "confirm_final")
workflow.add_edge("confirm_final", END)

# ==================== 编译 ====================

# 使用 PostgreSQL 作为检查点存储
checkpointer = PostgresSaver(connection_string)

# 编译图，指定中断点
app = workflow.compile(
    checkpointer=checkpointer,
    interrupt_before=[
        "confirm_characters",
        "confirm_storyboards",
        "confirm_images",
        "confirm_final"
    ],
    interrupt_after=[
        "wait_for_images",
        "wait_for_videos"
    ]
)
```

### 3.3 中断点设计

| 中断点 | 触发时机 | 用户操作 |
|--------|----------|----------|
| `confirm_characters` | 角色生成完成后 | 确认/修改/重新生成 |
| `confirm_storyboards` | 分镜生成完成后 | 确认/修改/重新生成 |
| `wait_for_images` | 图片任务提交后 | 等待完成/查看进度 |
| `confirm_images` | 图片生成完成后 | 确认/重新生成不满意 |
| `wait_for_videos` | 视频任务提交后 | 等待完成/查看进度 |
| `confirm_final` | 最终合成后 | 确认/重新合成 |

---

## 四、异步任务新形式

### 4.1 方案 1: LangGraph 内置 Tool 等待（推荐简单场景）

```python
from langgraph.prebuilt import ToolNode
from langchain_core.tools import tool

@tool
def generate_image(prompt: str, negative_prompt: str = "") -> str:
    """生成图片，返回图片URL"""
    # 调用 Go 后端同步执行
    result = go_backend_client.generate_image(
        prompt=prompt,
        negative_prompt=negative_prompt
    )
    return result.url

@tool
def generate_video(image_url: str, prompt: str, duration: int = 5) -> str:
    """生成视频，返回视频URL"""
    result = go_backend_client.generate_video(
        image_url=image_url,
        prompt=prompt,
        duration=duration
    )
    return result.url

# 添加工具节点
workflow.add_node("tools", ToolNode([generate_image, generate_video]))
```

**优点**: 简单，LangGraph 自动管理
**缺点**: 长时间阻塞，不适合批量操作

### 4.2 方案 2: Human-in-the-loop + 外部回调（推荐生产环境）

```python
# 节点：提交异步任务
def submit_image_tasks_node(state: DramaState) -> DramaState:
    """提交所有图片生成任务"""
    tasks = {}

    for storyboard in state["storyboards"]:
        # 调用 Go 后端异步接口
        task_id = go_backend_client.submit_image_task(
            drama_id=state["drama_id"],
            storyboard_id=storyboard["id"],
            prompt=storyboard["image_prompt"],
            negative_prompt=storyboard.get("negative_prompt", "")
        )
        tasks[f"image_{storyboard['id']}"] = task_id

    return {
        **state,
        "pending_tasks": tasks,
        "current_step": "waiting_images"
    }

# 节点：等待任务完成（会触发 interrupt）
def wait_for_images_node(state: DramaState) -> DramaState | None:
    """检查图片任务是否全部完成"""
    completed = {}
    still_pending = {}

    for task_key, task_id in state["pending_tasks"].items():
        status = go_backend_client.get_task_status(task_id)

        if status.status == "completed":
            completed[task_key] = status.result_url
        elif status.status == "failed":
            # 记录错误，可以重试
            return {
                **state,
                "errors": state["errors"] + [{
                    "task_key": task_key,
                    "error": status.error_message
                }]
            }
        else:
            still_pending[task_key] = task_id

    if not still_pending:
        # 所有任务完成
        return {
            **state,
            "images": {**state["images"], **completed},
            "pending_tasks": {},
            "current_step": "images_completed"
        }

    # 仍有未完成任务，返回 None 触发 interrupt
    # LangGraph 会保存当前状态并暂停
    return None

# 外部回调：任务完成后恢复执行
async def on_image_task_completed(task_id: str, result_url: str, thread_id: str):
    """Go 后端通过 webhook 调用此函数"""
    # 更新 Go 后端的任务状态（已完成）
    # ...

    # 恢复 LangGraph 执行
    app.invoke(
        None,  # 不传入新状态，使用保存的状态
        config={"configurable": {"thread_id": thread_id}}
    )
```

**优点**: 支持批量并行，不阻塞
**缺点**: 需要额外的回调机制

### 4.3 方案 3: 流式 + 异步混合（最佳用户体验）

```python
import asyncio
from typing import AsyncIterator

async def generate_storyboards_node(state: DramaState) -> AsyncIterator[dict]:
    """流式生成分镜"""
    storyboards = []

    # 流式调用 LLM
    async for chunk in llm.astream(state["outline"]):
        # 解析部分结果
        partial = parse_storyboard_chunk(chunk)
        storyboards.append(partial)

        # 流式输出给前端
        yield {
            "type": "storyboard_partial",
            "content": partial
        }

    return {**state, "storyboards": storyboards}

async def wait_for_images_node(state: DramaState) -> AsyncIterator[dict]:
    """并行等待图片任务，实时反馈进度"""
    images = {}
    tasks = list(state["pending_tasks"].items())

    # 使用 asyncio 并行等待
    async def wait_single_task(task_key: str, task_id: str):
        while True:
            status = await go_backend_client.async_get_task_status(task_id)
            yield {
                "type": "image_progress",
                "task_key": task_key,
                "status": status.status,
                "progress": status.progress
            }

            if status.status == "completed":
                return task_key, status.result_url
            elif status.status == "failed":
                raise Exception(f"Task {task_key} failed: {status.error_message}")

            await asyncio.sleep(2)

    # 并行执行所有等待
    async for result in async_parallel_wait(tasks, wait_single_task):
        if isinstance(result, Exception):
            yield {
                "type": "image_error",
                "error": str(result)
            }
        else:
            task_key, url = result
            images[task_key] = url
            yield {
                "type": "image_completed",
                "task_key": task_key,
                "url": url
            }

    return {**state, "images": images, "pending_tasks": {}}
```

**优点**: 实时反馈，用户体验最好
**缺点**: 实现复杂度较高

---

## 五、推荐架构

### 5.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                    LangGraph + 混合异步模式                          │
└─────────────────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────────────────┐
│                           前端 (Vue 3)                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
│  │  对话界面   │  │  确认对话框  │  │  进度展示   │  │  预览播放   │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  │
│                              │                                        │
│                    SSE / WebSocket                                   │
└──────────────────────────────│────────────────────────────────────────┘
                               ↓
┌───────────────────────────────────────────────────────────────────────┐
│                    Python LangGraph 服务                               │
│  ┌───────────────────────────────────────────────────────────────┐   │
│  │                    StateGraph (DramaAgent)                     │   │
│  │                                                                │   │
│  │  ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐        │   │
│  │  │ 解析    │ → │ 生成    │ → │⏸ 确认  │ → │ 生成    │        │   │
│  │  │ 大纲    │   │ 角色    │   │ 角色    │   │ 章节    │        │   │
│  │  └─────────┘   └─────────┘   └─────────┘   └─────────┘        │   │
│  │        ↑                                       │              │   │
│  │        │                                       ↓              │   │
│  │  ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐        │   │
│  │  │⏸ 确认  │ ← │ 合成    │ ← │⏸ 等待  │ ← │ 生成    │        │   │
│  │  │ 最终    │   │ 视频    │   │ 视频    │   │ 分镜    │        │   │
│  │  └─────────┘   └─────────┘   └─────────┘   └─────────┘        │   │
│  │                                    ↑                          │   │
│  │                                    │                          │   │
│  │                            ┌───────┴───────┐                  │   │
│  │                            │ ⏸ 等待图片    │                  │   │
│  │                            └───────────────┘                  │   │
│  └───────────────────────────────────────────────────────────────┘   │
│                              │                                        │
│                    PostgreSQL Checkpointer                           │
└──────────────────────────────│────────────────────────────────────────┘
                               │ gRPC / HTTP
                               ↓
┌───────────────────────────────────────────────────────────────────────┐
│                    Go 后端 (任务执行器)                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
│  │ 图片生成    │  │ 视频生成    │  │ 视频合成    │  │ 文件存储    │  │
│  │ Service     │  │ Service     │  │ Service     │  │ Service     │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  │
│                              │                                        │
│                    Webhook 回调                                       │
└──────────────────────────────│────────────────────────────────────────┘
                               ↓
                    ┌─────────────────────┐
                    │  外部 AI 服务       │
                    │  - OpenAI / DALL-E  │
                    │  - 豆包 / MiniMax   │
                    │  - Stable Diffusion │
                    └─────────────────────┘
```

### 5.2 职责划分

| 层级 | 职责 | 技术 |
|------|------|------|
| **前端** | 对话交互、确认对话框、进度展示、预览播放 | Vue 3 + SSE |
| **LangGraph** | 流程编排、状态管理、中断恢复、LLM 调用 | Python + LangGraph |
| **Go 后端** | 图片/视频生成、文件存储、Webhook 回调 | Go + Gin |
| **外部 AI** | 图像生成、视频生成 | OpenAI/豆包/MiniMax |

### 5.3 关键设计决策

| 决策 | 选择 | 理由 |
|------|------|------|
| LLM 调用 | LangGraph 直接调用 | 支持流式输出，用户体验好 |
| 图片/视频生成 | Go 后端执行 | 已有成熟实现，计算密集型适合 Go |
| 状态持久化 | PostgreSQL Checkpointer | 生产级可靠性，支持分布式 |
| 前后端通信 | SSE + WebSocket | 流式输出 + 实时进度 |
| 回调机制 | Webhook | 解耦，支持异步任务完成通知 |

---

## 六、具体实现建议

### 6.1 Go 后端改造

Go 后端从"流程控制者"变为"纯执行器"：

```go
// 保留现有的生成服务，但移除任务轮询机制
// 改为同步执行 + Webhook 回调

// 同步接口（简单场景）
func (h *ImageHandler) GenerateImageSync(c *gin.Context) {
    var req GenerateImageRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    // 同步执行
    result, err := h.imageService.GenerateImageSync(req)
    if err != nil {
        response.InternalError(c, err.Error())
        return
    }

    response.Success(c, result)
}

// 异步接口 + 回调（生产场景）
func (h *ImageHandler) GenerateImageAsync(c *gin.Context) {
    var req GenerateImageAsyncRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    // 创建任务
    taskID := h.taskService.CreateTask(req)

    // 启动后台处理
    go func() {
        result, err := h.imageService.GenerateImageSync(req.GenerateImageRequest)

        // 更新任务状态
        h.taskService.UpdateTaskStatus(taskID, result, err)

        // Webhook 回调 LangGraph
        if req.WebhookURL != "" {
            h.webhookClient.Notify(req.WebhookURL, taskID, result, err)
        }
    }()

    response.Success(c, gin.H{
        "task_id": taskID,
    })
}
```

### 6.2 Python LangGraph 服务

```python
# services/drama_agent.py
from langgraph.graph import StateGraph
from langgraph.checkpoint.postgres import PostgresSaver

class DramaAgent:
    def __init__(self, db_url: str, go_backend_url: str):
        self.go_backend = GoBackendClient(go_backend_url)
        self.checkpointer = PostgresSaver(db_url)
        self.graph = self._build_graph()

    def _build_graph(self) -> StateGraph:
        workflow = StateGraph(DramaState)
        # ... 构建 graph (见上文)
        return workflow.compile(
            checkpointer=self.checkpointer,
            interrupt_before=[...],
            interrupt_after=[...]
        )

    async def start(self, outline: str, thread_id: str) -> AsyncIterator[dict]:
        """开始新创作"""
        async for event in self.graph.astream(
            {"outline": outline, "drama_id": generate_id()},
            config={"configurable": {"thread_id": thread_id}}
        ):
            yield self._format_event(event)

    async def resume(
        self,
        thread_id: str,
        confirmation: dict
    ) -> AsyncIterator[dict]:
        """用户确认后恢复"""
        async for event in self.graph.astream(
            {"user_confirmations": confirmation},
            config={"configurable": {"thread_id": thread_id}}
        ):
            yield self._format_event(event)

    async def get_state(self, thread_id: str) -> dict:
        """获取当前状态"""
        state = await self.graph.aget_state(
            config={"configurable": {"thread_id": thread_id}}
        )
        return state.values

    def _format_event(self, event: dict) -> dict:
        """格式化事件给前端"""
        # 转换为前端可理解的格式
        return {
            "type": event.get("__type__", "unknown"),
            "data": event
        }
```

### 6.3 前端交互

```typescript
// api/agent.ts
export class DramaAgent {
    private threadId: string
    private eventSource: EventSource | null = null

    async start(outline: string): Promise<void> {
        this.threadId = generateThreadId()

        // 使用 SSE 接收流式事件
        this.eventSource = new EventSource(
            `/api/agent/start?thread_id=${this.threadId}&outline=${encodeURIComponent(outline)}`
        )

        return new Promise((resolve, reject) => {
            this.eventSource!.onmessage = (event) => {
                const data = JSON.parse(event.data)
                this.handleEvent(data)

                if (data.type === 'interrupt') {
                    resolve()
                }
            }
            this.eventSource!.onerror = reject
        })
    }

    async resume(confirmation: dict): Promise<void> {
        // 关闭旧的 SSE
        this.eventSource?.close()

        // 发送确认并建立新的 SSE
        const response = await fetch('/api/agent/resume', {
            method: 'POST',
            body: JSON.stringify({
                thread_id: this.threadId,
                confirmation
            })
        })

        // 建立新的 SSE 连接
        this.eventSource = new EventSource(
            `/api/agent/stream?thread_id=${this.threadId}`
        )

        return new Promise((resolve, reject) => {
            this.eventSource!.onmessage = (event) => {
                const data = JSON.parse(event.data)
                this.handleEvent(data)

                if (data.type === 'interrupt' || data.type === 'end') {
                    resolve()
                }
            }
            this.eventSource!.onerror = reject
        })
    }

    private handleEvent(event: dict): void {
        switch (event.type) {
            case 'character_partial':
                this.updateCharacterPreview(event.data)
                break
            case 'storyboard_partial':
                this.updateStoryboardPreview(event.data)
                break
            case 'image_progress':
                this.updateImageProgress(event.task_key, event.progress)
                break
            case 'image_completed':
                this.updateImageResult(event.task_key, event.url)
                break
            case 'interrupt':
                this.showConfirmDialog(event.data)
                break
            case 'error':
                this.showError(event.message)
                break
        }
    }
}
```

---

## 七、迁移路径

### 7.1 分阶段迁移

```
Phase 1: 基础设施
├── 部署 Python LangGraph 服务
├── 配置 PostgreSQL Checkpointer
└── 建立 Python ↔ Go 通信

Phase 2: 单节点迁移
├── 迁移角色生成为 LangGraph 节点
├── 添加流式输出
└── 测试中断/恢复

Phase 3: 完整流程
├── 迁移所有节点到 LangGraph
├── 实现异步任务等待
└── 集成 Webhook 回调

Phase 4: 前端适配
├── 实现对话式交互
├── 实现确认对话框
└── 实现进度展示
```

### 7.2 兼容性策略

| 阶段 | 策略 |
|------|------|
| Phase 1-2 | 新旧并存，新功能走 LangGraph，旧功能保持不变 |
| Phase 3 | 灰度发布，部分用户使用新流程 |
| Phase 4 | 完全切换，下线旧接口 |

### 7.3 回滚方案

每个阶段都保留回滚能力：

1. **Phase 1**: LangGraph 服务独立部署，不影响现有功能
2. **Phase 2**: 通过配置开关切换新旧流程
3. **Phase 3**: 保留 Go 后端异步任务接口作为 fallback
4. **Phase 4**: 数据库兼容，可快速回滚

---

## 八、总结

| 方面 | 现有模式 | LangGraph 模式 |
|------|----------|----------------|
| **流程控制** | 分散在各 Service | StateGraph 统一编排 |
| **用户交互** | 轮询 + 手动刷新 | 流式 + 实时确认 |
| **状态管理** | 多表分散 | Checkpointer 统一 |
| **错误处理** | Task 表记录 | State.errors + 重试节点 |
| **可恢复性** | 困难 | 天然支持 |
| **LLM 调用** | 异步任务 | 同步 + 流式 |
| **媒体生成** | 异步任务 | 保留异步 + 回调 |

**核心变化**：异步任务从"流程控制者"变成"纯执行器"，LangGraph 接管所有流程编排和状态管理。

---

## 九、硬编码优化案例：帧提示词生成

### 9.1 问题分析

现有 `frame_prompt_service.go` 存在以下硬编码问题：

| 问题 | 代码位置 | 影响 |
|------|----------|------|
| 帧类型常量 | L38-44 | 新增类型需修改代码 |
| switch-case 分支 | L130-166 | 耦合度高，难以扩展 |
| 分镜板逻辑 | L358-373 | 3格/4格逻辑硬编码 |
| 降级方案 | 多处 | 每种类型独立实现 |
| Prompt 调用 | 各 generate 方法 | 重复模式 |

### 9.2 LangGraph 动态图方案

```python
from typing import TypedDict, Literal
from langgraph.graph import StateGraph
from langchain_core.tools import tool

# ==================== 配置驱动的帧类型定义 ====================

FRAME_CONFIGS = {
    "first": {
        "name": "首帧",
        "description": "镜头开始的静态画面",
        "prompt_template": "first_frame",
        "fallback_suffix": "first frame, static shot",
        "is_multi_frame": False
    },
    "key": {
        "name": "关键帧",
        "description": "动作高潮瞬间",
        "prompt_template": "key_frame",
        "fallback_suffix": "key frame, dynamic action",
        "is_multi_frame": False
    },
    "last": {
        "name": "尾帧",
        "description": "镜头结束画面",
        "prompt_template": "last_frame",
        "fallback_suffix": "last frame, final state",
        "is_multi_frame": False
    },
    "panel": {
        "name": "分镜板",
        "description": "多格组合",
        "is_multi_frame": True,
        "frame_sequence": ["first", "key", "last"],  # 动态配置序列
        "layout_template": "horizontal_{count}"
    },
    "action": {
        "name": "动作序列",
        "description": "3x3宫格",
        "prompt_template": "action_sequence",
        "is_multi_frame": True,
        "layout": "grid_3x3"
    }
}

# ==================== 状态定义 ====================

class FramePromptState(TypedDict):
    storyboard_id: str
    frame_type: str
    panel_count: int
    model: str

    # 上下文（由节点填充）
    storyboard_context: dict
    scene_context: dict
    drama_style: str

    # 输出
    prompts: list[dict]
    layout: str
    error: str | None

# ==================== 通用节点（消除重复代码） ====================

async def load_context_node(state: FramePromptState) -> FramePromptState:
    """加载分镜上下文 - 替代每个方法重复的数据库查询"""
    storyboard = await db.get_storyboard(state["storyboard_id"])
    scene = await db.get_scene(storyboard.scene_id) if storyboard.scene_id else None
    drama = await db.get_drama_from_episode(storyboard.episode_id)

    return {
        **state,
        "storyboard_context": storyboard.to_dict(),
        "scene_context": scene.to_dict() if scene else None,
        "drama_style": drama.style
    }

async def generate_single_frame_node(state: FramePromptState) -> FramePromptState:
    """通用单帧生成 - 替代 generateFirstFrame/generateKeyFrame/generateLastFrame"""
    config = FRAME_CONFIGS[state["frame_type"]]

    # 构建 prompt（统一逻辑）
    system_prompt = prompt_i18n.get(config["prompt_template"], state["drama_style"])
    user_prompt = build_context_prompt(state["storyboard_context"], state["scene_context"])

    # 调用 AI（统一调用）
    try:
        response = await ai_client.generate(user_prompt, system_prompt, model=state["model"])
        prompt = parse_json_response(response)
    except Exception as e:
        # 统一降级逻辑
        prompt = build_fallback(
            state["storyboard_context"],
            state["scene_context"],
            config["fallback_suffix"]
        )

    return {
        **state,
        "prompts": [{"prompt": prompt, "description": config["description"]}],
        "layout": "single"
    }

async def generate_multi_frame_node(state: FramePromptState) -> FramePromptState:
    """通用多帧生成 - 替代 generatePanelFrames/generateActionSequence"""
    config = FRAME_CONFIGS[state["frame_type"]]
    prompts = []

    if "frame_sequence" in config:
        # 分镜板模式：按序列生成多个单帧
        for i, frame_type in enumerate(config["frame_sequence"]):
            # 递归调用子图（或直接调用单帧节点）
            single_result = await generate_single_frame_node({
                **state,
                "frame_type": frame_type
            })
            prompts.append({
                **single_result["prompts"][0],
                "description": f"第{i+1}格：{FRAME_CONFIGS[frame_type]['name']}"
            })

        layout = config["layout_template"].format(count=len(prompts))
    else:
        # 动作序列模式：一次性生成
        system_prompt = prompt_i18n.get(config["prompt_template"], state["drama_style"])
        user_prompt = build_context_prompt(state["storyboard_context"], state["scene_context"])

        try:
            response = await ai_client.generate(user_prompt, system_prompt, model=state["model"])
            prompt = parse_json_response(response)
        except Exception:
            prompt = build_fallback(state["storyboard_context"], state["scene_context"], "action sequence")

        prompts = [{"prompt": prompt, "description": config["description"]}]
        layout = config["layout"]

    return {
        **state,
        "prompts": prompts,
        "layout": layout
    }

# ==================== 动态路由 ====================

def route_by_frame_type(state: FramePromptState) -> Literal["single", "multi"]:
    """根据配置决定路由，而非 switch-case"""
    config = FRAME_CONFIGS[state["frame_type"]]
    return "multi" if config["is_multi_frame"] else "single"

# ==================== 构建动态图 ====================

def build_frame_prompt_graph():
    workflow = StateGraph(FramePromptState)

    # 添加节点
    workflow.add_node("load_context", load_context_node)
    workflow.add_node("generate_single", generate_single_frame_node)
    workflow.add_node("generate_multi", generate_multi_frame_node)
    workflow.add_node("save_result", save_result_node)

    # 定义边
    workflow.set_entry_point("load_context")
    workflow.add_conditional_edges(
        "load_context",
        route_by_frame_type,
        {
            "single": "generate_single",
            "multi": "generate_multi"
        }
    )
    workflow.add_edge("generate_single", "save_result")
    workflow.add_edge("generate_multi", "save_result")
    workflow.add_edge("save_result", END)

    return workflow.compile()

# ==================== 使用示例 ====================

async def generate_frame_prompt(storyboard_id: str, frame_type: str, model: str = ""):
    graph = build_frame_prompt_graph()

    result = await graph.ainvoke({
        "storyboard_id": storyboard_id,
        "frame_type": frame_type,
        "model": model,
        "panel_count": 3
    })

    return result["prompts"], result["layout"]
```

### 9.3 扩展性对比

| 场景 | 现有代码 | LangGraph 方案 |
|------|----------|----------------|
| 新增帧类型 | 修改常量 + 新增方法 + 修改 switch | 只需在 `FRAME_CONFIGS` 添加配置 |
| 修改分镜板序列 | 修改硬编码 if-else | 修改 `frame_sequence` 配置 |
| 新增 5 格分镜板 | 不支持，需大量代码 | `"frame_sequence": ["first", "key", "key", "key", "last"]` |
| 调整降级逻辑 | 修改多处 | 修改 `build_fallback` 一处 |
| 添加中间件（日志、监控） | 每个方法单独添加 | LangGraph 节点自动继承 |

### 9.4 更进一步：LLM 驱动的动态生成

```python
# 让 LLM 决定如何生成分镜，而非硬编码

@tool
def generate_frame_prompt(
    storyboard_context: str,
    frame_type: str,
    style: str
) -> dict:
    """根据分镜上下文和帧类型生成图像提示词"""
    # 由 LLM 决定调用参数
    pass

# Agent 决策
FRAME_AGENT_PROMPT = """
你是一个分镜提示词生成专家。
根据用户提供的分镜信息，决定：
1. 需要生成几个帧？
2. 每个帧应该是什么类型（首帧/关键帧/尾帧）？
3. 使用什么布局？

输出格式：
{
    "frames": [
        {"type": "first", "description": "..."},
        {"type": "key", "description": "..."},
        {"type": "last", "description": "..."}
    ],
    "layout": "horizontal_3"
}
"""

async def ai_driven_frame_generation(state: FramePromptState) -> FramePromptState:
    """让 AI 决定帧序列，完全消除硬编码"""
    # 1. 让 AI 分析分镜，决定帧序列
    frame_plan = await llm.invoke(
        FRAME_AGENT_PROMPT,
        context=state["storyboard_context"]
    )

    # 2. 按计划生成每个帧
    prompts = []
    for frame in frame_plan["frames"]:
        result = await generate_single_frame_node({
            **state,
            "frame_type": frame["type"]
        })
        prompts.append(result["prompts"][0])

    return {
        **state,
        "prompts": prompts,
        "layout": frame_plan["layout"]
    }
```

### 9.5 总结

| 优化点 | 现有方案 | LangGraph 方案 |
|--------|----------|----------------|
| **帧类型** | 常量 + switch | 配置驱动 |
| **生成逻辑** | 每类型一个方法 | 通用节点 + 配置 |
| **分镜板序列** | 硬编码 if-else | 配置数组 |
| **降级逻辑** | 分散在各方法 | 统一处理 |
| **扩展性** | 需修改代码 | 只需配置 |
| **AI 辅助** | 无 | 可让 LLM 决定帧序列 |

**核心思想**：将硬编码的业务规则转换为**可配置的数据**，让 Graph 根据配置动态路由。

---

## 十、Eino 框架替代方案分析

### 10.1 Eino 框架概述

**Eino** 是字节跳动开源的 AI Agent 开发框架，基于 Go 语言实现。

| 特性 | 说明 |
|------|------|
| **语言** | Go（与现有项目技术栈一致） |
| **Human-in-the-Loop** | ✅ 原生支持多种模式 |
| **工具包装器** | ✅ `InvokableApprovableTool`, `InvokableReviewEditTool` |
| **检查点** | ✅ `CheckPointStore` |
| **流式输出** | ✅ `AsyncIterator` |
| **预构建 Agent** | ✅ `planexecute`, `supervisor` 等 |

### 10.2 Human-in-the-Loop 模式对比

| 模式 | 说明 | 适用场景 |
|------|------|----------|
| **Approval** | 简单 Y/N 确认 | 敏感操作（资金转账、视频合成） |
| **Review-Edit** | 可修改参数后执行 | 需要修正的操作（角色、分镜生成） |
| **Feedback-Loop** | 执行后收集反馈优化 | 迭代改进场景 |
| **Follow-Up** | 执行后追问 | 需要补充信息 |

### 10.3 多智能体模式对比

| 模式 | 结构 | 适用场景 |
|------|------|----------|
| **Supervisor** | 主管 → 子智能体 | 层级任务分配 |
| **Plan-Execute-Replan** | 规划器 → 执行器 → 重规划器 | 多步骤复杂任务 |
| **Supervisor + Plan-Execute** | 主管 + 嵌套 Plan-Execute | 复杂项目级任务 |

### 10.4 推荐模式：Plan-Execute-Replan + Review-Edit

**最适合短剧创作项目的模式组合**：

```
┌─────────────────────────────────────────────────────────────────────┐
│              Plan-Execute-Replan + Review-Edit 模式                 │
└─────────────────────────────────────────────────────────────────────┘

                    ┌─────────────┐
                    │   用户输入   │
                    │ "创作短剧"  │
                    └──────┬──────┘
                           ↓
                    ┌─────────────┐
                    │   Planner   │ ← 规划器：生成创作计划
                    │  生成计划   │
                    └──────┬──────┘
                           ↓
         ┌─────────────────────────────────────┐
         │            执行循环                  │
         │  ┌─────────┐                        │
         │  │ Executor│ ← 执行器：执行每个步骤  │
         │  └────┬────┘                        │
         │       ↓                             │
         │  ┌─────────────┐                    │
         │  │ Tool 调用   │                    │
         │  │ - 生成角色  │                    │
         │  │ - 生成分镜  │                    │
         │  │ - 生成图片  │                    │
         │  └──────┬──────┘                    │
         │         ↓                           │
         │  ┌─────────────────┐                │
         │  │ ⏸ Interrupt    │ ← Review-Edit  │
         │  │ 等待用户审阅    │   中断机制      │
         │  │ - ok: 批准     │                │
         │  │ - edit: 修改   │                │
         │  │ - n: 拒绝      │                │
         │  └────────┬────────┘                │
         │           ↓                         │
         │  ┌─────────────────┐                │
         │  │ Resume 恢复执行  │                │
         │  └────────┬────────┘                │
         │           ↓                         │
         │  ┌─────────────────┐                │
         │  │ 步骤完成？       │                │
         │  └────────┬────────┘                │
         │      是 ↙   ↘ 否                   │
         │     ↓         ↓                     │
         │  ┌──────┐  ┌───────────┐            │
         │  │ END  │  │ Replanner │ ← 重规划   │
         │  └──────┘  └───────────┘            │
         └─────────────────────────────────────┘
```

### 10.5 代码实现示例

```go
package main

import (
    "context"
    "fmt"

    "github.com/cloudwego/eino/adk"
    "github.com/cloudwego/eino/adk/prebuilt/planexecute"
    "github.com/cloudwego/eino/components/model"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"
)

// ==================== 短剧创作 Agent ====================

func NewDramaPlanner(ctx context.Context) (adk.Agent, error) {
    return planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
        ToolCallingChatModel: newChatModel(),
        SystemPrompt: `你是短剧创作规划专家。
根据用户需求，创建详细的短剧制作计划，包括：
1. 生成角色
2. 创建章节
3. 生成分镜
4. 生成场景图片
5. 生成角色图片
6. 生成分镜视频
7. 视频合成
8. 最终确认`,
    })
}

func NewDramaExecutor(ctx context.Context) (adk.Agent, error) {
    // 获取所有短剧创作工具
    dramaTools, err := GetDramaTools(ctx)
    if err != nil {
        return nil, err
    }

    return planexecute.NewExecutor(ctx, &planexecute.ExecutorConfig{
        Model: newChatModel(),
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: dramaTools,
            },
        },
    })
}

func NewDramaReplanner(ctx context.Context) (adk.Agent, error) {
    return planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
        ChatModel: newChatModel(),
    })
}

func NewDramaProductionAgent(ctx context.Context) (adk.Agent, error) {
    planner, _ := NewDramaPlanner(ctx)
    executor, _ := NewDramaExecutor(ctx)
    replanner, _ := NewDramaReplanner(ctx)

    return planexecute.New(ctx, &planexecute.Config{
        Planner:       planner,
        Executor:      executor,
        Replanner:     replanner,
        MaxIterations: 50,  // 短剧创作可能需要更多步骤
    })
}

// ==================== 工具定义（带审阅） ====================

// GenerateCharactersTool 生成角色工具 - 需要审阅编辑
func GenerateCharactersTool(ctx context.Context) (adk.Tool, error) {
    // 基础工具
    baseTool := schema.NewToolInfo(
        "generate_characters",
        "根据剧本大纲生成角色列表",
        map[string]*schema.ParameterInfo{
            "drama_id": {Desc: "剧本ID"},
            "outline":  {Desc: "剧本大纲"},
        },
    )

    // 包装为可审阅编辑的工具
    return adk.NewInvokableReviewEditTool(baseTool, func(ctx context.Context, args string) (string, error) {
        // 调用现有的 Go 服务
        return characterService.GenerateCharacters(args)
    }), nil
}

// GenerateStoryboardTool 生成分镜工具 - 需要审阅编辑
func GenerateStoryboardTool(ctx context.Context) (adk.Tool, error) {
    baseTool := schema.NewToolInfo(
        "generate_storyboard",
        "根据章节内容生成分镜",
        map[string]*schema.ParameterInfo{
            "episode_id": {Desc: "章节ID"},
            "script":     {Desc: "章节剧本"},
        },
    )

    return adk.NewInvokableReviewEditTool(baseTool, func(ctx context.Context, args string) (string, error) {
        return storyboardService.GenerateStoryboard(args)
    }), nil
}

// MergeVideosTool 合成视频工具 - 需要审批（耗时操作）
func MergeVideosTool(ctx context.Context) (adk.Tool, error) {
    baseTool := schema.NewToolInfo(
        "merge_videos",
        "合成多个视频片段",
        map[string]*schema.ParameterInfo{
            "video_clips": {Desc: "视频片段列表"},
            "transition":  {Desc: "转场效果", Optional: true},
        },
    )

    // 使用审批模式（只能批准/拒绝，不能修改）
    return adk.NewInvokableApprovableTool(baseTool, func(ctx context.Context, args string) (string, error) {
        return videoService.MergeVideos(args)
    }), nil
}

// ==================== 主流程 ====================

func main() {
    ctx := context.Background()

    // 创建短剧创作 Agent
    agent, _ := NewDramaProductionAgent(ctx)

    // 创建 Runner（带检查点）
    runner := adk.NewRunner(ctx, adk.RunnerConfig{
        EnableStreaming: true,
        Agent:           agent,
        CheckPointStore: store.NewInMemoryStore(), // 生产环境用 PostgreSQL
    })

    // 用户请求
    query := `创作一部古装短剧《霸道总裁爱上我》，共3集。
    风格：古风言情
    主要角色：霸道王爷、穿越女主、绿茶妹妹`

    // 执行查询
    iter := runner.Query(ctx, query, adk.WithCheckPointID("drama-001"))

    // 处理事件和中断
    for {
        lastEvent, interrupted := processEvents(iter)
        if !interrupted {
            break
        }

        // 处理中断
        interruptCtx := lastEvent.Action.Interrupted.InterruptContexts[0]

        // 根据中断类型处理
        switch info := interruptCtx.Info.(type) {
        case *ReviewEditInfo:
            // Review-Edit 模式：可以修改参数
            handleReviewEdit(info)
        case *ApprovalInfo:
            // Approval 模式：只能批准/拒绝
            handleApproval(info)
        }

        // 恢复执行
        iter, _ = runner.ResumeWithParams(ctx, "drama-001", &adk.ResumeParams{
            Targets: map[string]any{
                interruptCtx.ID: interruptCtx.Info,
            },
        })
    }
}

func handleReviewEdit(info *ReviewEditInfo) {
    fmt.Printf("工具 %s 即将执行，参数:\n%s\n", info.ToolName, info.ArgumentsInJSON)
    fmt.Println("选项: 'ok' 批准 / 'n' 拒绝 / 输入修改后的 JSON")

    var input string
    fmt.Scanln(&input)

    result := &ReviewEditResult{}
    switch input {
    case "ok":
        result.NoNeedToEdit = true
    case "n":
        result.Disapproved = true
    default:
        result.EditedArgumentsInJSON = &input
    }
    info.ReviewResult = result
}

func handleApproval(info *ApprovalInfo) {
    fmt.Printf("敏感操作 %s 需要审批:\n%s\n", info.ToolName, info.ArgumentsInJSON)
    fmt.Print("批准? (Y/N): ")

    var input string
    fmt.Scanln(&input)

    info.Approved = (input == "Y" || input == "y")
}
```

### 10.6 Eino vs LangGraph 对比

| 维度 | Eino | LangGraph |
|------|------|-----------|
| **语言** | **Go**（与现有项目一致） | Python |
| **集成成本** | **低**（直接集成） | 高（需新服务） |
| **Human-in-the-Loop** | ✅ 原生多种模式 | ✅ `interrupt()` |
| **状态持久化** | ✅ `CheckPointStore` | ✅ `Checkpointer` |
| **可视化** | ⚠️ 有限 | ✅ LangGraph Studio |
| **调试工具** | ⚠️ 有限 | ✅ LangSmith |
| **社区生态** | 新兴 | 成熟 |
| **学习曲线** | 中等 | 较陡 |
| **与 Go 服务集成** | **无缝** | 需要跨服务调用 |

### 10.7 短剧创作工具与中断模式映射

| 工具 | 中断模式 | 原因 |
|------|----------|------|
| `generate_characters` | **Review-Edit** | 用户可能需要修改角色设定 |
| `generate_episodes` | **Review-Edit** | 用户可能需要修改章节内容 |
| `generate_storyboard` | **Review-Edit** | 用户可能需要调整分镜 |
| `generate_image` | **Approval** | 图片生成耗资源，确认后执行 |
| `generate_video` | **Approval** | 视频生成耗资源，确认后执行 |
| `merge_videos` | **Approval** | 合成耗时，确认后执行 |

### 10.8 推荐选择

| 场景 | 推荐方案 |
|------|----------|
| **快速集成到现有 Go 项目** | **Eino** ✅ |
| **需要完整可视化调试** | LangGraph |
| **团队熟悉 Python** | LangGraph |
| **团队熟悉 Go** | **Eino** ✅ |
| **需要复杂图结构** | LangGraph |
| **需要标准工作流模式** | **Eino** ✅ |
| **希望保持单体架构** | **Eino** ✅ |

### 10.9 Eino 迁移路径

```
Phase 1: 引入 Eino ADK (1周)
├── 添加 eino 依赖
├── 定义短剧创作工具（Tools）
├── 使用 Plan-Execute-Replan 模式
└── 跑通基础流程

Phase 2: 集成现有服务 (1-2周)
├── 将现有 Go Service 封装为工具
├── 添加 Review-Edit 审阅机制
├── 实现 CheckPoint 持久化
└── 替换现有异步任务机制

Phase 3: 前端适配 (1周)
├── SSE 实时输出
├── 中断确认对话框
└── 进度展示

Phase 4: 优化迭代 (持续)
├── 调整 Prompt 模板
├── 优化中断时机
├── 添加重试机制
└── 监控和日志
```

### 10.10 结论

| 选择 | 适用情况 |
|------|----------|
| **Eino** | 保持 Go 技术栈、快速集成、单体架构 |
| **LangGraph** | 需要 Python 生态、复杂图结构、可视化调试 |

**对于 huobao-drama 项目，Eino 的 Plan-Execute-Replan + Review-Edit 模式是最佳选择**，原因：

1. **技术栈一致**：Go 语言，无需引入新服务
2. **集成成本低**：直接调用现有 Service
3. **功能满足需求**：原生支持多阶段人工确认
4. **学习曲线平缓**：API 设计简洁

---

## 十一、LangGraph 多 Agent + Graph 架构设计

### 11.1 为什么适合多 Agent 架构？

| 特点 | 分析 |
|------|------|
| **多阶段流程** | 8+ 个阶段，每个阶段有不同的专业需求 |
| **能力差异** | LLM 创意、图片生成、视频生成、视频合成需要不同能力 |
| **并行机会** | 场景图片、角色图片、分镜图片可并行生成 |
| **成本优化** | 简单任务用 Haiku，复杂创作用 Sonnet |
| **人工确认** | 每个阶段完成后需要确认 |

### 11.2 整体架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    LangGraph 多 Agent 架构                               │
└─────────────────────────────────────────────────────────────────────────┘

                              ┌─────────────┐
                              │   用户输入   │
                              └──────┬──────┘
                                     ↓
                        ┌────────────────────────┐
                        │    Supervisor Agent    │
                        │      (主管智能体)        │
                        │                        │
                        │  - 理解用户意图         │
                        │  - 分配任务给专业Agent  │
                        │  - 汇总结果            │
                        │  - 协调流程            │
                        └───────────┬────────────┘
                                    │
            ┌───────────────────────┼───────────────────────┐
            │                       │                       │
            ↓                       ↓                       ↓
    ┌───────────────┐       ┌───────────────┐       ┌───────────────┐
    │ Script Agent  │       │ Design Agent  │       │ Render Agent  │
    │  (编剧智能体)  │       │  (设计智能体)  │       │  (渲染智能体)  │
    │               │       │               │       │               │
    │ - 角色生成    │       │ - 场景图片    │       │ - 分镜视频    │
    │ - 章节生成    │       │ - 角色图片    │       │ - 视频合成    │
    │ - 分镜生成    │       │ - 分镜图片    │       │               │
    └───────┬───────┘       └───────┬───────┘       └───────┬───────┘
            │                       │                       │
            │                       │                       │
    ┌───────┴───────┐       ┌───────┴───────┐       ┌───────┴───────┐
    │   子图/工具    │       │   子图/工具    │       │   子图/工具    │
    │               │       │               │       │               │
    │ • 角色工具    │       │ • 图片生成    │       │ • 视频生成    │
    │ • 章节工具    │       │   (并行)      │       │ • FFmpeg合成  │
    │ • 分镜工具    │       │               │       │               │
    └───────────────┘       └───────────────┘       └───────────────┘

            ↑                       ↑                       ↑
            │                       │                       │
            └───────────────────────┴───────────────────────┘
                              共享状态 + 中断确认
```

### 11.3 详细代码实现

```python
from typing import TypedDict, Annotated, Literal
from langgraph.graph import StateGraph, END
from langgraph.checkpoint.postgres import PostgresSaver
from langgraph.types import interrupt, Command
from langchain_anthropic import ChatAnthropic
from langchain_core.messages import HumanMessage, AIMessage
import operator

# ==================== 全局状态 ====================

class DramaProductionState(TypedDict):
    """短剧创作全局状态"""
    # 基本信息
    drama_id: str
    title: str
    outline: str
    style: str

    # 创作内容
    characters: list[dict]
    episodes: list[dict]
    storyboards: list[dict]

    # 媒体资源
    scene_images: dict[str, str]      # {scene_id: image_url}
    character_images: dict[str, str]  # {character_id: image_url}
    storyboard_images: dict[str, str] # {storyboard_id: image_url}
    storyboard_videos: dict[str, str] # {storyboard_id: video_url}
    final_video: str

    # 流程控制
    current_phase: str
    completed_phases: list[str]
    pending_tasks: dict[str, str]

    # 人工确认
    confirmations: Annotated[dict, operator.or_]

    # 消息历史
    messages: Annotated[list, operator.add]

    # 错误处理
    errors: list[dict]

# ==================== Supervisor Agent ====================

class SupervisorAgent:
    """主管智能体 - 协调整个创作流程"""

    def __init__(self):
        self.model = ChatAnthropic(model="claude-sonnet-4-5-20250514")

    def __call__(self, state: DramaProductionState) -> Command:
        # 分析当前状态，决定下一步
        if not state.get("characters"):
            return Command(goto="script_agent", update={"current_phase": "script"})

        if not state.get("scene_images") or not state.get("character_images"):
            return Command(goto="design_agent", update={"current_phase": "design"})

        if not state.get("storyboard_videos"):
            return Command(goto="render_agent", update={"current_phase": "render"})

        if not state.get("final_video"):
            return Command(goto="merge_agent", update={"current_phase": "merge"})

        return Command(goto=END)

# ==================== Script Agent (编剧智能体) ====================

class ScriptAgent:
    """编剧智能体 - 负责角色、章节、分镜生成"""

    def __init__(self):
        # 使用 Sonnet 处理创意任务
        self.model = ChatAnthropic(model="claude-sonnet-4-5-20250514")
        self.sub_graph = self._build_sub_graph()

    def _build_sub_graph(self) -> StateGraph:
        """构建编剧子图"""
        workflow = StateGraph(DramaProductionState)

        # 子节点
        workflow.add_node("generate_characters", self._generate_characters)
        workflow.add_node("confirm_characters", self._confirm_characters)
        workflow.add_node("generate_episodes", self._generate_episodes)
        workflow.add_node("confirm_episodes", self._confirm_episodes)
        workflow.add_node("generate_storyboards", self._generate_storyboards)
        workflow.add_node("confirm_storyboards", self._confirm_storyboards)

        # 边
        workflow.set_entry_point("generate_characters")
        workflow.add_edge("generate_characters", "confirm_characters")
        workflow.add_conditional_edges(
            "confirm_characters",
            lambda s: "regenerate" if s.get("needs_regenerate") else "proceed",
            {"regenerate": "generate_characters", "proceed": "generate_episodes"}
        )
        workflow.add_edge("generate_episodes", "confirm_episodes")
        workflow.add_edge("confirm_episodes", "generate_storyboards")
        workflow.add_edge("generate_storyboards", "confirm_storyboards")
        workflow.add_edge("confirm_storyboards", END)

        return workflow.compile()

    async def _generate_characters(self, state: DramaProductionState):
        """生成角色"""
        prompt = f"""根据以下剧本大纲，生成角色列表：
        标题：{state['title']}
        大纲：{state['outline']}
        风格：{state.get('style', '现代')}

        返回 JSON 格式的角色列表。"""

        response = await self.model.ainvoke(prompt)
        characters = parse_json(response.content)

        return {"characters": characters}

    async def _confirm_characters(self, state: DramaProductionState):
        """确认角色 - 中断等待用户确认"""
        decision = interrupt({
            "type": "character_confirmation",
            "characters": state["characters"],
            "message": "请确认生成的角色是否满意",
            "options": ["approve", "edit", "regenerate"]
        })

        if decision["action"] == "approve":
            return {"completed_phases": state["completed_phases"] + ["characters"]}
        elif decision["action"] == "edit":
            return {"characters": decision["edited_characters"]}
        else:
            return {"needs_regenerate": True}

    def __call__(self, state: DramaProductionState):
        """执行编剧子图"""
        return self.sub_graph.invoke(state)

# ==================== Design Agent (设计智能体) ====================

class DesignAgent:
    """设计智能体 - 负责图片生成（可并行）"""

    def __init__(self):
        self.model = ChatAnthropic(model="claude-haiku-4-5-20250219")  # 用 Haiku 降成本
        self.sub_graph = self._build_sub_graph()

    def _build_sub_graph(self) -> StateGraph:
        """构建设计子图 - 支持并行"""
        workflow = StateGraph(DramaProductionState)

        # 并行生成节点
        workflow.add_node("generate_all_images", self._generate_all_images_parallel)
        workflow.add_node("confirm_images", self._confirm_images)

        workflow.set_entry_point("generate_all_images")
        workflow.add_edge("generate_all_images", "confirm_images")
        workflow.add_edge("confirm_images", END)

        return workflow.compile()

    async def _generate_all_images_parallel(self, state: DramaProductionState):
        """并行生成所有图片"""
        import asyncio

        # 构建并行任务
        tasks = []

        # 场景图片任务
        for scene in state.get("scenes", []):
            tasks.append(self._generate_scene_image(scene, state))

        # 角色图片任务
        for character in state.get("characters", []):
            tasks.append(self._generate_character_image(character, state))

        # 分镜图片任务
        for storyboard in state.get("storyboards", []):
            tasks.append(self._generate_storyboard_image(storyboard, state))

        # 并行执行
        results = await asyncio.gather(*tasks, return_exceptions=True)

        # 汇总结果
        scene_images = {}
        character_images = {}
        storyboard_images = {}
        errors = []

        for result in results:
            if isinstance(result, Exception):
                errors.append({"error": str(result)})
            elif result["type"] == "scene":
                scene_images[result["id"]] = result["url"]
            elif result["type"] == "character":
                character_images[result["id"]] = result["url"]
            elif result["type"] == "storyboard":
                storyboard_images[result["id"]] = result["url"]

        return {
            "scene_images": scene_images,
            "character_images": character_images,
            "storyboard_images": storyboard_images,
            "errors": state.get("errors", []) + errors
        }

    async def _generate_scene_image(self, scene: dict, state: DramaProductionState):
        """生成单个场景图片 - 调用 Go 后端"""
        # 调用 Go 后端的图片生成服务
        result = await go_backend_client.generate_image(
            prompt=scene["prompt"],
            negative_prompt=state.get("negative_prompt", ""),
            image_type="scene"
        )
        return {"type": "scene", "id": scene["id"], "url": result["url"]}

    async def _confirm_images(self, state: DramaProductionState):
        """确认图片 - 中断等待用户确认"""
        decision = interrupt({
            "type": "image_confirmation",
            "scene_images": state["scene_images"],
            "character_images": state["character_images"],
            "message": "请确认生成的图片是否满意",
            "options": ["approve", "regenerate_selected", "regenerate_all"]
        })

        if decision["action"] == "approve":
            return {"completed_phases": state["completed_phases"] + ["design"]}
        elif decision["action"] == "regenerate_selected":
            return {"pending_regenerate": decision["selected_ids"]}
        else:
            return {"scene_images": {}, "character_images": {}, "storyboard_images": {}}

    def __call__(self, state: DramaProductionState):
        return self.sub_graph.invoke(state)

# ==================== Render Agent (渲染智能体) ====================

class RenderAgent:
    """渲染智能体 - 负责视频生成和合成"""

    def __init__(self):
        self.model = ChatAnthropic(model="claude-haiku-4-5-20250219")
        self.sub_graph = self._build_sub_graph()

    def _build_sub_graph(self) -> StateGraph:
        workflow = StateGraph(DramaProductionState)

        workflow.add_node("generate_videos", self._generate_videos)
        workflow.add_node("confirm_videos", self._confirm_videos)

        workflow.set_entry_point("generate_videos")
        workflow.add_edge("generate_videos", "confirm_videos")
        workflow.add_edge("confirm_videos", END)

        return workflow.compile()

    async def _generate_videos(self, state: DramaProductionState):
        """生成分镜视频"""
        import asyncio

        tasks = []
        for storyboard in state.get("storyboards", []):
            image_url = state["storyboard_images"].get(storyboard["id"])
            tasks.append(self._generate_storyboard_video(storyboard, image_url, state))

        results = await asyncio.gather(*tasks, return_exceptions=True)

        storyboard_videos = {}
        errors = []

        for result in results:
            if isinstance(result, Exception):
                errors.append({"error": str(result)})
            else:
                storyboard_videos[result["id"]] = result["url"]

        return {
            "storyboard_videos": storyboard_videos,
            "errors": state.get("errors", []) + errors
        }

    async def _generate_storyboard_video(self, storyboard: dict, image_url: str, state: DramaProductionState):
        """生成单个分镜视频 - 调用 Go 后端"""
        result = await go_backend_client.generate_video(
            image_url=image_url,
            prompt=storyboard["video_prompt"],
            duration=storyboard.get("duration", 5)
        )
        return {"id": storyboard["id"], "url": result["url"]}

    async def _confirm_videos(self, state: DramaProductionState):
        decision = interrupt({
            "type": "video_confirmation",
            "videos": state["storyboard_videos"],
            "message": "请确认生成的视频是否满意",
            "options": ["approve", "regenerate_selected", "regenerate_all"]
        })

        if decision["action"] == "approve":
            return {"completed_phases": state["completed_phases"] + ["render"]}
        else:
            return {"storyboard_videos": {}}

    def __call__(self, state: DramaProductionState):
        return self.sub_graph.invoke(state)

# ==================== Merge Agent (合成智能体) ====================

class MergeAgent:
    """合成智能体 - 负责最终视频合成"""

    async def __call__(self, state: DramaProductionState):
        # 确认合成
        decision = interrupt({
            "type": "merge_confirmation",
            "video_count": len(state["storyboard_videos"]),
            "message": "即将合成所有视频片段，确认执行？",
            "options": ["approve", "cancel"]
        })

        if decision["action"] == "cancel":
            return Command(goto=END)

        # 调用 Go 后端合成
        result = await go_backend_client.merge_videos(
            video_urls=list(state["storyboard_videos"].values()),
            transition="fade"
        )

        return {
            "final_video": result["url"],
            "completed_phases": state["completed_phases"] + ["merge"]
        }

# ==================== 构建主图 ====================

def build_drama_production_graph():
    """构建短剧创作主图"""

    workflow = StateGraph(DramaProductionState)

    # 添加节点
    workflow.add_node("supervisor", SupervisorAgent())
    workflow.add_node("script_agent", ScriptAgent())
    workflow.add_node("design_agent", DesignAgent())
    workflow.add_node("render_agent", RenderAgent())
    workflow.add_node("merge_agent", MergeAgent())

    # 设置入口
    workflow.set_entry_point("supervisor")

    # Supervisor 路由到各个 Agent
    workflow.add_conditional_edges(
        "supervisor",
        lambda s: s.get("current_phase", "script"),
        {
            "script": "script_agent",
            "design": "design_agent",
            "render": "render_agent",
            "merge": "merge_agent",
        }
    )

    # Agent 完成后返回 Supervisor
    workflow.add_edge("script_agent", "supervisor")
    workflow.add_edge("design_agent", "supervisor")
    workflow.add_edge("render_agent", "supervisor")
    workflow.add_edge("merge_agent", END)

    # 编译
    checkpointer = PostgresSaver("postgresql://...")
    return workflow.compile(
        checkpointer=checkpointer,
        interrupt_before=["confirm_characters", "confirm_images", "confirm_videos", "merge_agent"]
    )

# ==================== 使用示例 ====================

async def main():
    graph = build_drama_production_graph()

    # 初始状态
    initial_state = {
        "drama_id": "drama-001",
        "title": "霸道总裁爱上我",
        "outline": "一个关于...",
        "style": "古风言情",
        "completed_phases": [],
        "messages": [],
        "errors": []
    }

    config = {"configurable": {"thread_id": "production-001"}}

    # 执行
    async for event in graph.astream(initial_state, config):
        print(f"Event: {event}")

        # 处理中断
        if event.get("__interrupt__"):
            interrupt_data = event["__interrupt__"]
            user_response = await get_user_response(interrupt_data)

            # 恢复执行
            async for resume_event in graph.astream(
                Command(resume=user_response),
                config
            ):
                print(f"Resume Event: {resume_event}")
```

### 11.4 Agent 职责划分

| Agent | 职责 | 推荐模型 | 原因 |
|-------|------|----------|------|
| **Supervisor** | 协调、路由、汇总 | Sonnet 4.5 | 需要理解复杂状态 |
| **Script Agent** | 角色生成、章节、分镜 | Sonnet 4.5 | 需要创意能力 |
| **Design Agent** | 图片生成（并行） | Haiku 4.5 | 简单指令，降成本 |
| **Render Agent** | 视频生成 | Haiku 4.5 | 简单指令，降成本 |
| **Merge Agent** | 视频合成 | Haiku 4.5 | 简单指令，降成本 |

### 11.5 成本优化策略

```
┌─────────────────────────────────────────────────────────────────────┐
│                        模型使用策略                                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Sonnet 4.5 (高成本，高能力)                                        │
│  ├── Supervisor: 理解复杂状态、路由决策                             │
│  └── Script Agent: 创意内容生成（角色、章节、分镜）                  │
│                                                                      │
│  Haiku 4.5 (低成本，够用)                                           │
│  ├── Design Agent: 图片生成指令（调用 Go 后端）                     │
│  ├── Render Agent: 视频生成指令（调用 Go 后端）                     │
│  └── Merge Agent: 合成指令（调用 Go 后端）                          │
│                                                                      │
│  预估成本节省: 60-70%                                               │
└─────────────────────────────────────────────────────────────────────┘
```

### 11.6 并行执行设计

```python
# Design Agent 内部并行
async def _generate_all_images_parallel(self, state):
    tasks = [
        self._generate_scene_image(scene, state),      # 场景1
        self._generate_scene_image(scene2, state),     # 场景2
        self._generate_character_image(char, state),   # 角色1
        self._generate_character_image(char2, state),  # 角色2
        self._generate_storyboard_image(sb, state),    # 分镜1
        # ...
    ]
    results = await asyncio.gather(*tasks)
```

### 11.7 中断点设计

| 中断点 | 位置 | 目的 |
|--------|------|------|
| `confirm_characters` | Script Agent | 确认角色设定 |
| `confirm_episodes` | Script Agent | 确认章节内容 |
| `confirm_storyboards` | Script Agent | 确认分镜脚本 |
| `confirm_images` | Design Agent | 确认图片质量 |
| `confirm_videos` | Render Agent | 确认视频效果 |
| `merge_agent` | Merge Agent | 最终合成确认 |

### 11.8 总结

| 方面 | 单 Agent | 多 Agent + Graph |
|------|----------|------------------|
| **复杂度** | 低 | 中高 |
| **可维护性** | 一般 | **好**（职责分离） |
| **成本优化** | 难 | **容易**（不同模型） |
| **并行能力** | 有限 | **强**（子图并行） |
| **调试能力** | 一般 | **好**（独立调试） |
| **扩展性** | 一般 | **强**（添加新 Agent） |

**结论**：对于短剧创作这类多阶段、多能力的复杂任务，**多 Agent + Graph 模式非常适合**，特别是：
1. 职责清晰分离
2. 可以使用不同模型优化成本
3. 支持并行执行提升效率
4. 易于扩展和维护

---

## 十二、Eino 多 Agent 架构设计

### 12.1 Eino 多 Agent 能力

Eino **完全支持**多 Agent 架构，提供以下预构建模式：

| 模式 | 包路径 | 说明 |
|------|--------|------|
| **Supervisor** | `adk/prebuilt/supervisor` | 主管协调多个子 Agent |
| **Plan-Execute-Replan** | `adk/prebuilt/planexecute` | 规划-执行-重规划循环 |
| **Layered Supervisor** | 自定义 | 嵌套主管（主管的子 Agent 也是主管） |
| **Supervisor + Plan-Execute** | 组合使用 | 主管协调 Plan-Execute 子 Agent |

### 12.2 短剧创作多 Agent 架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Eino 多 Agent 架构                                    │
└─────────────────────────────────────────────────────────────────────────┘

                              ┌─────────────┐
                              │   用户输入   │
                              └──────┬──────┘
                                     ↓
                        ┌────────────────────────┐
                        │    Drama Supervisor    │
                        │     (短剧主管智能体)     │
                        │                        │
                        │  - 理解创作需求         │
                        │  - 分配任务给专业Agent  │
                        │  - 汇总结果            │
                        │  - 协调流程            │
                        └───────────┬────────────┘
                                    │
            ┌───────────────────────┼───────────────────────┐
            │                       │                       │
            ↓                       ↓                       ↓
    ┌───────────────┐       ┌───────────────┐       ┌───────────────┐
    │ Script Agent  │       │ Design Agent  │       │ Render Agent  │
    │  (编剧智能体)  │       │  (设计智能体)  │       │  (渲染智能体)  │
    │               │       │               │       │               │
    │ • 角色生成    │       │ • 场景图片    │       │ • 分镜视频    │
    │ • 章节生成    │       │ • 角色图片    │       │ • 视频合成    │
    │ • 分镜生成    │       │ • 分镜图片    │       │               │
    │               │       │               │       │               │
    │ Plan-Execute  │       │  并行工具调用  │       │  并行工具调用  │
    │   模式        │       │               │       │               │
    └───────────────┘       └───────────────┘       └───────────────┘
            │                       │                       │
            ↓                       ↓                       ↓
    ┌───────────────────────────────────────────────────────────────┐
    │                      共享 Go 后端服务                          │
    │  • CharacterService  • ImageGenerationService                 │
    │  • EpisodeService    • VideoGenerationService                 │
    │  • StoryboardService • VideoMergeService                      │
    └───────────────────────────────────────────────────────────────┘
```

### 12.3 代码实现

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/cloudwego/eino/adk"
    "github.com/cloudwego/eino/adk/prebuilt/supervisor"
    "github.com/cloudwego/eino/adk/prebuilt/planexecute"
    "github.com/cloudwego/eino/components/model"
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/components/tool/utils"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"

    // 项目内部服务
    "github.com/drama-generator/backend/application/services"
)

// ==================== 模型配置 ====================

func newSonnetModel() model.ToolCallingChatModel {
    // 创意任务使用 Sonnet
    return openai.NewChatModel(ctx, &openai.ChatModelConfig{
        Model: "claude-sonnet-4-5-20250514",
    })
}

func newHaikuModel() model.ToolCallingChatModel {
    // 简单指令使用 Haiku（降成本）
    return openai.NewChatModel(ctx, &openai.ChatModelConfig{
        Model: "claude-haiku-4-5-20250219",
    })
}

// ==================== 工具定义 ====================

// 角色生成工具（带审阅）
func NewGenerateCharactersTool() (tool.BaseTool, error) {
    type generateCharactersReq struct {
        DramaID string `json:"drama_id" jsonschema_description:"剧本ID"`
        Outline string `json:"outline" jsonschema_description:"剧本大纲"`
        Style   string `json:"style" jsonschema_description:"风格"`
    }

    generateCharacters := func(ctx context.Context, req *generateCharactersReq) (string, error) {
        // 调用现有 Go 服务
        characters, err := services.CharacterService.GenerateCharacters(ctx, req.DramaID, req.Outline, req.Style)
        if err != nil {
            return "", err
        }
        return toJSON(characters), nil
    }

    baseTool, _ := utils.InferTool(
        "generate_characters",
        "根据剧本大纲生成角色列表，包括姓名、性格、外貌、声音风格等",
        generateCharacters,
    )

    // 包装为可审阅编辑的工具
    return &InvokableReviewEditTool{InvokableTool: baseTool}, nil
}

// 图片生成工具（带审批）
func NewGenerateImageTool() (tool.BaseTool, error) {
    type generateImageReq struct {
        Prompt         string `json:"prompt" jsonschema_description:"图片描述"`
        NegativePrompt string `json:"negative_prompt" jsonschema_description:"负面提示词"`
        ImageType      string `json:"image_type" jsonschema_description:"图片类型: scene/character/storyboard"`
        ReferenceURL   string `json:"reference_url" jsonschema_description:"参考图片URL"`
    }

    generateImage := func(ctx context.Context, req *generateImageReq) (string, error) {
        // 调用现有 Go 服务
        result, err := services.ImageGenerationService.Generate(ctx, &services.GenerateImageRequest{
            Prompt:         req.Prompt,
            NegativePrompt: req.NegativePrompt,
            ImageType:      req.ImageType,
            ReferenceURL:   req.ReferenceURL,
        })
        if err != nil {
            return "", err
        }
        return result.ImageURL, nil
    }

    baseTool, _ := utils.InferTool(
        "generate_image",
        "生成图片（场景/角色/分镜），返回图片URL",
        generateImage,
    )

    // 包装为可审批的工具
    return &InvokableApprovableTool{InvokableTool: baseTool}, nil
}

// 视频生成工具（带审批）
func NewGenerateVideoTool() (tool.BaseTool, error) {
    type generateVideoReq struct {
        ImageURL string `json:"image_url" jsonschema_description:"参考图片URL"`
        Prompt   string `json:"prompt" jsonschema_description:"视频描述"`
        Duration int    `json:"duration" jsonschema_description:"视频时长（秒）"`
    }

    generateVideo := func(ctx context.Context, req *generateVideoReq) (string, error) {
        result, err := services.VideoGenerationService.Generate(ctx, &services.GenerateVideoRequest{
            ImageURL: req.ImageURL,
            Prompt:   req.Prompt,
            Duration: req.Duration,
        })
        if err != nil {
            return "", err
        }
        return result.VideoURL, nil
    }

    baseTool, _ := utils.InferTool(
        "generate_video",
        "生成分镜视频，返回视频URL",
        generateVideo,
    )

    return &InvokableApprovableTool{InvokableTool: baseTool}, nil
}

// 视频合成工具（带审批）
func NewMergeVideosTool() (tool.BaseTool, error) {
    type mergeVideosReq struct {
        VideoURLs         []string `json:"video_urls" jsonschema_description:"视频URL列表"`
        Transition        string   `json:"transition" jsonschema_description:"转场效果"`
        TransitionDuratio float64  `json:"transition_duration" jsonschema_description:"转场时长"`
    }

    mergeVideos := func(ctx context.Context, req *mergeVideosReq) (string, error) {
        result, err := services.VideoMergeService.Merge(ctx, &services.MergeVideosRequest{
            VideoURLs:         req.VideoURLs,
            Transition:        req.Transition,
            TransitionDuration: req.TransitionDuration,
        })
        if err != nil {
            return "", err
        }
        return result.OutputURL, nil
    }

    baseTool, _ := utils.InferTool(
        "merge_videos",
        "合成多个视频片段为一个完整视频",
        mergeVideos,
    )

    return &InvokableApprovableTool{InvokableTool: baseTool}, nil
}

// ==================== Script Agent (编剧智能体) ====================

func NewScriptAgent(ctx context.Context) (adk.Agent, error) {
    // 获取工具
    charactersTool, _ := NewGenerateCharactersTool()
    episodesTool, _ := NewGenerateEpisodesTool()
    storyboardsTool, _ := NewGenerateStoryboardsTool()

    // 使用 Plan-Execute-Replan 模式
    planner, _ := planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
        ToolCallingChatModel: newSonnetModel(),
    })

    executor, _ := planexecute.NewExecutor(ctx, &planexecute.ExecutorConfig{
        Model: newSonnetModel(),
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: []tool.BaseTool{charactersTool, episodesTool, storyboardsTool},
            },
        },
    })

    replanner, _ := planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
        ChatModel: newSonnetModel(),
    })

    return planexecute.New(ctx, &planexecute.Config{
        Planner:       planner,
        Executor:      executor,
        Replanner:     replanner,
        MaxIterations: 20,
    })
}

// ==================== Design Agent (设计智能体) ====================

func NewDesignAgent(ctx context.Context) (adk.Agent, error) {
    // 获取图片生成工具
    imageTool, _ := NewGenerateImageTool()

    return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
        Name:        "design_agent",
        Description: "负责生成所有图片（场景、角色、分镜），支持并行生成",
        Instruction: `你是一个专业的短剧设计智能体。

任务：
1. 根据分镜信息生成场景图片
2. 根据角色设定生成角色图片
3. 根据分镜描述生成分镜参考图

规则：
- 使用 generate_image 工具生成图片
- 可以同时发起多个图片生成请求（并行）
- 生成完成后，汇总所有图片URL返回给主管

输出格式：
{
    "scene_images": {"scene_1": "url1", ...},
    "character_images": {"char_1": "url1", ...},
    "storyboard_images": {"sb_1": "url1", ...}
}`,
        Model: newHaikuModel(), // 简单指令用 Haiku
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: []tool.BaseTool{imageTool},
            },
        },
    })
}

// ==================== Render Agent (渲染智能体) ====================

func NewRenderAgent(ctx context.Context) (adk.Agent, error) {
    // 获取工具
    videoTool, _ := NewGenerateVideoTool()
    mergeTool, _ := NewMergeVideosTool()

    return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
        Name:        "render_agent",
        Description: "负责视频生成和最终合成",
        Instruction: `你是一个专业的短剧渲染智能体。

任务：
1. 使用分镜图片生成分镜视频
2. 将所有视频片段合成为最终视频

规则：
- 使用 generate_video 工具生成视频
- 使用 merge_videos 工具合成视频
- 视频生成可以并行执行
- 合成前需要用户确认

输出格式：
{
    "storyboard_videos": {"sb_1": "url1", ...},
    "final_video": "url"
}`,
        Model: newHaikuModel(),
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: []tool.BaseTool{videoTool, mergeTool},
            },
        },
    })
}

// ==================== Drama Supervisor (主管智能体) ====================

func NewDramaSupervisor(ctx context.Context) (adk.Agent, error) {
    return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
        Name:        "drama_supervisor",
        Description: "短剧创作主管，负责协调整个创作流程",
        Instruction: `你是短剧创作的主管智能体，管理三个专业智能体：

1. script_agent: 负责创意内容（角色、章节、分镜）
2. design_agent: 负责图片生成（场景、角色、分镜图片）
3. render_agent: 负责视频生成和合成

工作流程：
1. 首先委托 script_agent 生成角色、章节、分镜
2. 等待 script_agent 完成后，委托 design_agent 生成图片
3. 等待 design_agent 完成后，委托 render_agent 生成视频并合成
4. 汇总所有结果，返回给用户

规则：
- 按顺序委托任务，一个 Agent 完成后再委托下一个
- 每个 Agent 完成后，检查结果质量
- 如果某个阶段失败，可以重新委托
- 最后汇总所有结果给用户`,
        Model:  newSonnetModel(),
        Exit:   &adk.ExitTool{},
    })
}

// ==================== 构建多 Agent 系统 ====================

func BuildDramaProductionSystem(ctx context.Context) (adk.Agent, error) {
    // 创建主管
    supervisorAgent, err := NewDramaSupervisor(ctx)
    if err != nil {
        return nil, err
    }

    // 创建子 Agent
    scriptAgent, err := NewScriptAgent(ctx)
    if err != nil {
        return nil, err
    }

    designAgent, err := NewDesignAgent(ctx)
    if err != nil {
        return nil, err
    }

    renderAgent, err := NewRenderAgent(ctx)
    if err != nil {
        return nil, err
    }

    // 构建 Supervisor 模式
    return supervisor.New(ctx, &supervisor.Config{
        Supervisor: supervisorAgent,
        SubAgents:  []adk.Agent{scriptAgent, designAgent, renderAgent},
    })
}

// ==================== 主流程 ====================

func main() {
    ctx := context.Background()

    // 构建多 Agent 系统
    agent, err := BuildDramaProductionSystem(ctx)
    if err != nil {
        log.Fatalf("构建系统失败: %v", err)
    }

    // 创建 Runner（带检查点）
    runner := adk.NewRunner(ctx, adk.RunnerConfig{
        EnableStreaming: true,
        Agent:           agent,
        CheckPointStore: NewPostgresCheckPointStore(), // 生产环境用 PostgreSQL
    })

    // 用户请求
    query := `创作一部古装短剧《霸道总裁爱上我》，共3集。
    风格：古风言情
    主要角色：霸道王爷、穿越女主、绿茶妹妹
    场景：王府、御花园、集市`

    checkpointID := "drama-production-001"

    // 执行查询
    iter := runner.Query(ctx, query, adk.WithCheckPointID(checkpointID))

    // 处理事件和中断
    for {
        event, hasEvent := iter.Next()
        if !hasEvent {
            break
        }

        if event.Err != nil {
            log.Printf("错误: %v", event.Err)
            continue
        }

        // 打印事件
        printEvent(event)

        // 处理中断（人工确认）
        if event.Action != nil && event.Action.Interrupted != nil {
            interruptCtx := event.Action.Interrupted.InterruptContexts[0]

            // 根据中断类型处理
            switch info := interruptCtx.Info.(type) {
            case *ReviewEditInfo:
                // Review-Edit 模式
                handleReviewEdit(info)
            case *ApprovalInfo:
                // Approval 模式
                handleApproval(info)
            }

            // 恢复执行
            iter, err = runner.ResumeWithParams(ctx, checkpointID, &adk.ResumeParams{
                Targets: map[string]any{
                    interruptCtx.ID: interruptCtx.Info,
                },
            })
            if err != nil {
                log.Fatalf("恢复执行失败: %v", err)
            }
        }
    }

    fmt.Println("\n短剧创作完成！")
}

// ==================== 中断处理 ====================

func handleReviewEdit(info *ReviewEditInfo) {
    fmt.Printf("\n========================================\n")
    fmt.Printf("工具 %s 需要审阅:\n", info.ToolName)
    fmt.Printf("参数: %s\n", info.ArgumentsInJSON)
    fmt.Println("----------------------------------------")
    fmt.Println("选项:")
    fmt.Println("  - 输入 'ok' 批准")
    fmt.Println("  - 输入 'n' 拒绝")
    fmt.Println("  - 输入修改后的 JSON 参数")
    fmt.Println("----------------------------------------")

    var input string
    fmt.Print("你的选择: ")
    fmt.Scanln(&input)

    result := &ReviewEditResult{}
    switch strings.ToLower(strings.TrimSpace(input)) {
    case "ok", "y", "yes":
        result.NoNeedToEdit = true
    case "n", "no":
        result.Disapproved = true
        fmt.Print("拒绝原因: ")
        fmt.Scanln(&result.DisapproveReason)
    default:
        result.EditedArgumentsInJSON = &input
    }

    info.ReviewResult = result
}

func handleApproval(info *ApprovalInfo) {
    fmt.Printf("\n========================================\n")
    fmt.Printf("敏感操作 %s 需要审批:\n", info.ToolName)
    fmt.Printf("参数: %s\n", info.ArgumentsInJSON)
    fmt.Println("----------------------------------------")

    var input string
    fmt.Print("批准? (Y/N): ")
    fmt.Scanln(&input)

    info.Approved = (strings.ToLower(input) == "y" || strings.ToLower(input) == "yes")
}
```

### 12.4 Eino 多 Agent 模式对比

| 模式 | 适用场景 | 短剧创作中的角色 |
|------|----------|------------------|
| **Supervisor** | 协调多个专业 Agent | Drama Supervisor |
| **Plan-Execute-Replan** | 复杂多步骤任务 | Script Agent（角色→章节→分镜） |
| **Supervisor + Plan-Execute** | 嵌套复杂任务 | 主管 + 编剧子 Agent |

### 12.5 Agent 与工具映射

| Agent | 工具 | 中断模式 | 模型 |
|-------|------|----------|------|
| **Script Agent** | `generate_characters` | Review-Edit | Sonnet |
| | `generate_episodes` | Review-Edit | Sonnet |
| | `generate_storyboards` | Review-Edit | Sonnet |
| **Design Agent** | `generate_image` | Approval | Haiku |
| **Render Agent** | `generate_video` | Approval | Haiku |
| | `merge_videos` | Approval | Haiku |

### 12.6 Eino vs LangGraph 多 Agent 对比

| 维度 | Eino | LangGraph |
|------|------|-----------|
| **语言** | **Go**（与现有项目一致） | Python |
| **Supervisor 模式** | ✅ `adk/prebuilt/supervisor` | ✅ 需自己实现 |
| **Plan-Execute** | ✅ `adk/prebuilt/planexecute` | ✅ 需自己实现 |
| **嵌套 Agent** | ✅ 原生支持 | ✅ 子图 |
| **Human-in-the-Loop** | ✅ 多种模式 | ✅ `interrupt()` |
| **并行执行** | ✅ 工具级并行 | ✅ 子图并行 |
| **检查点** | ✅ `CheckPointStore` | ✅ `Checkpointer` |
| **可视化** | ⚠️ 有限 | ✅ Studio |
| **与 Go 服务集成** | **无缝** | 需要 HTTP/gRPC |

### 12.7 推荐架构选择

| 场景 | 推荐方案 |
|------|----------|
| **保持 Go 技术栈** | **Eino Supervisor + Plan-Execute** |
| **需要 Python 生态** | LangGraph 多 Agent |
| **单体架构** | **Eino** |
| **微服务架构** | LangGraph（独立服务） |
| **快速集成** | **Eino** |
| **可视化调试需求强** | LangGraph |

### 12.8 总结

**Eino 完全支持多 Agent 架构**，且更适合短剧创作项目：

| 优势 | 说明 |
|------|------|
| **技术栈一致** | Go 语言，无需引入 Python 服务 |
| **集成成本低** | 直接调用现有 Service |
| **预构建模式** | Supervisor、Plan-Execute 开箱即用 |
| **Human-in-the-Loop** | 原生支持 Review-Edit、Approval 等模式 |
| **成本优化** | 不同 Agent 可使用不同模型 |

**推荐架构**：
```
Drama Supervisor (协调)
    ├── Script Agent (Plan-Execute 模式)
    │   ├── generate_characters (Review-Edit)
    │   ├── generate_episodes (Review-Edit)
    │   └── generate_storyboards (Review-Edit)
    ├── Design Agent (并行工具调用)
    │   └── generate_image (Approval)
    └── Render Agent (并行工具调用)
        ├── generate_video (Approval)
        └── merge_videos (Approval)
```

---

## 十三、openOii 项目多 Agent 设计对比

### 13.1 openOii 项目概述

**openOii** 是一个开源的 AI 漫剧生成平台，与 huobao-drama 功能高度相似。

| 维度 | openOii | huobao-drama |
|------|---------|--------------|
| **语言** | Python + FastAPI | Go + Gin |
| **Agent 框架** | Claude Agent SDK | 可选 Eino/LangGraph |
| **数据库** | PostgreSQL + Redis | SQLite/PostgreSQL |
| **实时通信** | WebSocket | SSE/WebSocket |
| **LLM** | Claude (via SDK) | 多模型支持 |
| **开源** | ✅ MIT | ✅ |

### 13.2 openOii 多 Agent 架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    openOii 多 Agent 架构                                 │
└─────────────────────────────────────────────────────────────────────────┘

                    ┌─────────────────────────────────────┐
                    │      GenerationOrchestrator         │
                    │           (编排器)                   │
                    │                                     │
                    │  • 顺序执行 Agent                    │
                    │  • Redis 等待用户确认                │
                    │  • Review Agent 路由决策            │
                    │  • 清理下游数据                      │
                    └───────────────┬─────────────────────┘
                                    │
        ┌───────────────────────────┼───────────────────────────┐
        │                           │                           │
        ↓                           ↓                           ↓
┌───────────────┐           ┌───────────────┐           ┌───────────────┐
│ Onboarding    │           │ Director      │           │ Scriptwriter  │
│ Agent         │     →     │ Agent         │     →     │ Agent         │
│               │           │               │           │               │
│ • 需求分析    │           │ • 规划创作方向│           │ • 生成角色    │
│ • 项目初始化  │           │ • 视觉风格    │           │ • 生成分镜    │
│               │           │ • 剧情大纲    │           │               │
└───────────────┘           └───────────────┘           └───────────────┘
        │                           │                           │
        ↓                           ↓                           ↓
┌───────────────┐           ┌───────────────┐           ┌───────────────┐
│ Character     │           │ Storyboard    │           │ Video         │
│ Artist Agent  │     →     │ Artist Agent  │     →     │ Generator     │
│               │           │               │           │               │
│ • 角色图片    │           │ • 分镜首帧    │           │ • 分镜视频    │
│               │           │               │           │               │
└───────────────┘           └───────────────┘           └───────────────┘
        │                           │                           │
        ↓                           ↓                           ↓
┌───────────────┐           ┌───────────────┐
│ Video Merger  │           │ Review Agent  │
│ Agent         │     →     │ (路由器)       │
│               │           │               │
│ • 视频拼接    │           │ • 分析用户反馈│
│               │           │ • 决定重跑点  │
│               │           │ • 增量/全量   │
└───────────────┘           └───────────────┘
```

### 13.3 openOii Agent 职责

| Agent | 职责 | 输入 | 输出 |
|-------|------|------|------|
| **OnboardingAgent** | 需求分析、项目初始化 | 用户故事 | 项目设置 |
| **DirectorAgent** | 规划创作方向、视觉风格 | 项目信息 | 导演笔记、风格定义 |
| **ScriptwriterAgent** | 生成角色、分镜脚本 | 导演规划 | 角色、分镜描述 |
| **CharacterArtistAgent** | 生成角色图片 | 角色描述 | 角色图片 URL |
| **StoryboardArtistAgent** | 生成分镜首帧 | 分镜描述 | 分镜图片 URL |
| **VideoGeneratorAgent** | 生成分镜视频 | 分镜图片+描述 | 视频片段 URL |
| **VideoMergerAgent** | 拼接完整视频 | 视频片段列表 | 最终视频 URL |
| **ReviewAgent** | 分析反馈、路由决策 | 用户反馈 | 起始 Agent、模式 |

### 13.4 openOii 核心设计模式

#### 1. 编排器模式 (Orchestrator)

```python
class GenerationOrchestrator:
    def __init__(self):
        self.agents = [
            OnboardingAgent(),
            DirectorAgent(),
            ScriptwriterAgent(),
            CharacterArtistAgent(),
            StoryboardArtistAgent(),
            VideoGeneratorAgent(),
            VideoMergerAgent(),
            ReviewAgent(),  # 不参与正常流程
        ]

    async def run(self, project_id, run_id, request):
        # 顺序执行
        for agent in self.agents:
            if agent.name == "review":
                continue  # Review 不参与正常流程

            await agent.run(ctx)

            # 等待用户确认（通过 Redis）
            if not auto_mode:
                feedback = await self._wait_for_confirm(project_id, run_id, agent.name)
                if feedback:
                    # 有反馈，调用 Review Agent 决定从哪里重新开始
                    routing = await review_agent.run(ctx)
                    start_agent = routing["start_agent"]
                    # 清理下游数据，从 start_agent 重新执行
                    await self._cleanup_for_rerun(project_id, start_agent)
                    # 重新规划执行计划
                    plan = self._rebuild_plan(start_agent)
                    break
```

#### 2. 用户确认机制 (Redis 跨进程)

```python
async def wait_for_confirm_redis(run_id: int, timeout: int = 1800) -> bool:
    """通过 Redis 订阅等待用户确认"""
    r = await get_redis()
    channel = f"openoii:confirm_channel:{run_id}"

    pubsub = r.pubsub()
    await pubsub.subscribe(channel)

    try:
        while True:
            msg = await pubsub.get_message(timeout=1.0)
            if msg is not None:
                return True
    finally:
        await pubsub.unsubscribe(channel)

async def trigger_confirm_redis(run_id: int) -> bool:
    """用户确认时调用，发布信号"""
    r = await get_redis()
    await r.set(f"openoii:confirm:{run_id}", "1", ex=3600)
    await r.publish(f"openoii:confirm_channel:{run_id}", "confirm")
    return True
```

#### 3. Review Agent 路由决策

```python
class ReviewAgent(BaseAgent):
    async def run(self, ctx: AgentContext) -> dict:
        # 1. 获取用户反馈
        feedback = await self._get_latest_feedback(ctx)

        # 2. 获取项目状态
        state = await self._get_project_state(ctx)

        # 3. 调用 LLM 分析反馈
        prompt = json.dumps({"feedback": feedback, "state": state})
        resp = await self.call_llm(ctx, system_prompt=SYSTEM_PROMPT, user_prompt=prompt)
        data = extract_json(resp.text)

        # 4. 返回路由决策
        return {
            "start_agent": data["routing"]["start_agent"],  # 从哪个 Agent 重新开始
            "mode": data["routing"]["mode"],                 # "full" 或 "incremental"
            "target_ids": data.get("target_ids"),            # 精细化控制目标
        }
```

#### 4. 清理机制

```python
async def _cleanup_for_rerun(self, project_id, start_agent, mode):
    """根据起始 Agent 和模式清理数据"""

    if start_agent == "scriptwriter":
        # 从编剧重新开始，删除角色和分镜
        await self._delete_project_characters(project_id)
        await self._delete_project_shots(project_id)
    elif start_agent == "character_artist":
        # 重新生成角色图片，清空下游
        await self._clear_character_images(project_id)
        await self._clear_shot_images(project_id)
        await self._clear_shot_videos(project_id)
    elif start_agent == "storyboard_artist":
        # 重新生成分镜图片，清空下游
        await self._clear_shot_images(project_id)
        await self._clear_shot_videos(project_id)
    # ...
```

### 13.5 架构对比

| 维度 | openOii | huobao-drama (Eino) | huobao-drama (LangGraph) |
|------|---------|---------------------|--------------------------|
| **编排模式** | Orchestrator 顺序执行 | Supervisor 协调 | Supervisor + 子图 |
| **Agent 数量** | 8 个 | 3-4 个（更聚合） | 3-4 个（更聚合） |
| **路由决策** | Review Agent (LLM) | Supervisor (LLM) | Supervisor (LLM) |
| **用户确认** | Redis Pub/Sub | CheckPoint + Resume | interrupt() + Command |
| **重跑机制** | 清理数据 + 重新执行 | 同左 | 从任意检查点恢复 |
| **并行能力** | 单 Agent 内并行 | 工具级并行 | 子图级并行 |
| **状态持久化** | DB + Redis | CheckPointStore | PostgresSaver |
| **可视化** | 无 | 有限 | LangGraph Studio |

### 13.6 openOii 可借鉴的设计

| 设计 | 说明 | 借鉴价值 |
|------|------|----------|
| **Review Agent** | 专门处理用户反馈，决定路由 | ⭐⭐⭐⭐⭐ 非常值得借鉴 |
| **增量模式** | 只清理产物，保留数据结构 | ⭐⭐⭐⭐ 节省资源 |
| **精细化控制** | target_ids 指定重新生成的具体对象 | ⭐⭐⭐⭐ 提升用户体验 |
| **Agent Handoff 消息** | "@导演 邀请 @编剧 加入了群聊" | ⭐⭐⭐ 体验友好 |
| **阶段映射** | Agent → Stage (ideate/visualize/animate/deploy) | ⭐⭐⭐ 进度展示清晰 |
| **Redis 跨进程确认** | 支持多 Worker 共享确认状态 | ⭐⭐⭐ 分布式友好 |
| **完成信息模板** | AGENT_COMPLETION_INFO 统一管理 | ⭐⭐ 代码整洁 |

### 13.7 推荐借鉴的实现

#### 1. 添加 Review Agent（强烈推荐）

```go
// Eino 版本
func NewReviewAgent(ctx context.Context) (adk.Agent, error) {
    return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
        Name:        "review_agent",
        Description: "分析用户反馈，决定从哪个 Agent 重新执行",
        Instruction: `你是短剧创作的审核智能体。

任务：
1. 分析用户反馈的内容类型（角色/分镜/视频/风格/剧情）
2. 决定应该从哪个 Agent 开始重新执行
3. 决定是全量重新生成还是增量更新
4. 精确指定需要重新生成的目标 ID

输出格式：
{
    "start_agent": "script_agent" | "design_agent" | "render_agent",
    "mode": "full" | "incremental",
    "target_ids": {
        "character_ids": [1, 2],
        "storyboard_ids": [3, 4, 5]
    }
}`,
        Model: newSonnetModel(),
    })
}
```

#### 2. 增量模式支持

```go
type RerunMode string

const (
    RerunModeFull        RerunMode = "full"        // 全量重新生成
    RerunModeIncremental RerunMode = "incremental" // 增量更新（只清理产物）
)

type CleanupStrategy struct {
    Mode        RerunMode
    TargetIDs   *TargetIDs  // 精细化控制
    ClearImages bool
    ClearVideos bool
    ClearData   bool
}

func (s *DramaService) CleanupForRerun(projectID uint, startAgent string, mode RerunMode) error {
    switch startAgent {
    case "script_agent":
        if mode == RerunModeFull {
            // 全量：删除角色和分镜数据
            s.db.Where("project_id = ?", projectID).Delete(&models.Character{})
            s.db.Where("project_id = ?", projectID).Delete(&models.Storyboard{})
        }
        // 都要清理下游产物
        s.clearAllImages(projectID)
        s.clearAllVideos(projectID)
    case "design_agent":
        s.clearImages(projectID)
        s.clearVideos(projectID)
    case "render_agent":
        s.clearVideos(projectID)
    }
    return nil
}
```

#### 3. Agent 完成信息模板

```go
var AgentCompletionInfo = map[string]AgentCompletion{
    "script_agent": {
        Completed: "已完成剧本创作",
        Details:   "生成了角色设定和分镜脚本",
        Next:      "接下来将为角色生成参考图片",
        Question:  "剧本内容和角色设定是否满意？",
    },
    "design_agent": {
        Completed: "已完成图片生成",
        Details:   "生成了角色和分镜图片",
        Next:      "接下来将生成分镜视频",
        Question:  "图片效果是否满意？",
    },
    "render_agent": {
        Completed: "已完成视频生成",
        Details:   "生成了所有分镜视频",
        Next:      "接下来将拼接成完整视频",
        Question:  "视频效果如何？",
    },
}
```

### 13.8 总结：哪个设计更好？

| 维度 | openOii | huobao-drama 推荐 |
|------|---------|-------------------|
| **技术栈** | Python（AI 生态好） | Go（性能好，与现有代码一致） |
| **编排复杂度** | 中（Orchestrator 模式） | 中（Supervisor 模式） |
| **用户反馈处理** | **优**（专门 Review Agent） | 建议借鉴 |
| **增量更新** | **优**（支持精细化控制） | 建议借鉴 |
| **Agent 粒度** | 细（8 个） | 粗（3-4 个） |
| **可维护性** | 中 | **优**（职责更聚合） |

**推荐策略**：

1. **保留 Eino/Go 技术栈** - 与现有项目一致
2. **借鉴 openOii 的 Review Agent** - 专门处理用户反馈路由
3. **借鉴增量模式** - 节省资源，提升体验
4. **保持 Agent 粒度适中** - 3-4 个比 8 个更易维护
5. **借鉴完成信息模板** - 统一管理 Agent 输出
