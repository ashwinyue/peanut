# AI 视频创作工具对话式 Agent 实施指南

> 基于 `docs/agent.desing` 设计方案，详细说明当前项目代码需要的调整
>
> **版本**: v1.0
> **更新日期**: 2026-02-02

---

## 目录

- [一、架构调整总览](#一架构调整总览)
- [二、后端代码调整](#二后端代码调整)
- [三、前端代码调整](#三前端代码调整)
- [四、新增文件清单](#四新增文件清单)
- [五、实施步骤](#五实施步骤)
- [六、测试验证](#六测试验证)

---

## 一、架构调整总览

### 1.1 当前架构 vs 目标架构

```
当前架构（多步骤页面流程）:
┌─────────────────────────────────────────────────────────────┐
│  用户需要在多个页面间切换：                                  │
│  剧本编辑 → 分镜拆解 → 角色图片 → 场景图片 → 视频生成        │
│                                                             │
│  问题：                                                      │
│  ❌ 流程割裂，步骤多                                         │
│  ❌ 需要在6个页面间反复切换                                   │
│  ❌ 无法批量操作                                             │
│  ❌ 学习成本高                                               │
└─────────────────────────────────────────────────────────────┘

目标架构（Cursor 模式）:
┌─────────────────────────────────────────────────────────────┐
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ 分镜列表      │  │  媒体预览区    │  │  参数设置      │     │  ← 主界面（保留）
│  │ （可选中）    │  │              │  │  （完全保留）   │     │
│  │              │  │ [时间轴编辑器] │  │              │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐ │
│  │ 🤖 Agent 侧边栏（新增）                                │ │  ← 对话区域
│  │ [对话流 + 快捷指令]                                    │ │
│  │ 输入框...                                   [发送]    │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 核心设计原则

| 原则 | 说明 | 代码体现 |
|-----|------|---------|
| **主界面不变** | 保留原有专业视频编辑功能 | 无需修改现有组件 |
| **侧边栏增强** | 新增 Agent 对话区 | 新增 `AgentSidebar.vue` |
| **双向联动** | 对话 ↔ 主界面同步 | 事件总线 + 状态管理 |
| **消息驱动** | 用对话卡片承载状态 | 统一的消息类型定义 |

---

## 二、后端代码调整

### 2.1 新增 API 端点

**文件位置**: `api/handlers/agent.go` (新建)

```go
package handlers

import (
    "github.com/gin-gonic/gin"
    "github.com/drama-generator/backend/pkg/logger"
)

type AgentHandler struct {
    log *logger.Logger
}

func NewAgentHandler(log *logger.Logger) *AgentHandler {
    return &AgentHandler{log: log}
}

// ChatRequest Agent 对话请求
type ChatRequest struct {
    Query     string                 `json:"query" binding:"required"`
    ContextID  string                 `json:"context_id,omitempty"`  // 会话上下文ID
    DramaID   uint                   `json:"drama_id,omitempty"`    // 当前剧本ID
    EpisodeID uint                   `json:"episode_id,omitempty"`  // 当前章节ID
    Metadata  map[string]interface{} `json:"metadata,omitempty"`   // 扩展元数据
}

// SSE 流式响应
func (h *AgentHandler) ChatStream(c *gin.Context) {
    // 设置 SSE 响应头
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    c.Header("Access-Control-Allow-Origin", "*")

    var req ChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.sendError(c, "INVALID_PARAMS", err.Error())
        return
    }

    // 1. 解析用户自然语言指令
    intent := h.parseIntent(req.Query)

    // 2. 根据意图执行对应操作
    switch intent.Type {
    case "batch_modify":
        h.handleBatchModify(c, req, intent)
    case "generate":
        h.handleGenerate(c, req, intent)
    case "query":
        h.handleQuery(c, req, intent)
    case "suggest":
        h.handleSuggest(c, req, intent)
    default:
        h.handleUnknown(c, req)
    }
}

// Intent 解析结果
type Intent struct {
    Type       string                 `json:"type"`        // 操作类型
    Entities   map[string]string      `json:"entities"`    // 提取的实体
    Parameters map[string]interface{} `json:"parameters"`  // 参数
}

// parseIntent 解析用户意图
func (h *AgentHandler) parseIntent(query string) Intent {
    // TODO: 集成 LLM 进行意图识别
    // 当前版本使用规则匹配
    return Intent{
        Type: "unknown",
        Entities: make(map[string]string),
        Parameters: make(map[string]interface{}),
    }
}

// sendSSE 发送 SSE 事件
func (h *AgentHandler) sendSSE(c *gin.Context, eventType string, data interface{}) {
    c.SSEvent(eventType, data)
    c.Writer.Flush()
}

// sendError 发送错误
func (h *AgentHandler) sendError(c *gin.Context, code, message string) {
    h.sendSSE(c, "error", gin.H{
        "code":    code,
        "message": message,
    })
}
```

### 2.2 路由注册调整

**文件位置**: `api/routes/routes.go`

```go
// 在 api := r.Group("/api/v1") 内添加
agent := api.Group("/agent")
{
    agent.POST("/chat", agentHandler.ChatStream)
    agent.GET("/context/:context_id", agentHandler.GetContext)
    agent.POST("/feedback", agentHandler.SubmitFeedback)
}
```

### 2.3 意图处理实现

**文件位置**: `application/services/agent/intent_service.go` (新建)

```go
package agent

import (
    "context"
    "fmt"
)

type IntentService struct {
    aiService AIService
}

type Intent struct {
    Type       string                 `json:"type"`
    Entities   map[string]string      `json:"entities"`
    Parameters map[string]interface{} `json:"parameters"`
    Confidence float64                `json:"confidence"`
}

// ParseIntent 使用 LLM 解析用户意图
func (s *IntentService) ParseIntent(ctx context.Context, query string) (*Intent, error) {
    systemPrompt := `你是一个视频编辑助手的意图识别系统。
解析用户指令，返回 JSON 格式：

{
  "type": "操作类型（batch_modify/generate/query/suggest）",
  "entities": {
    "target": "操作对象（镜头/角色/场景等）",
    "scope": "操作范围（单个/批量/全部）"
  },
  "parameters": {
    "具体参数"
  }
}

示例：
用户: "所有德古拉镜头改成冷色调"
→ {"type":"batch_modify", "entities":{"target":"镜头","scope":"德古拉"}, "parameters":{"style":"冷色调"}}

用户: "生成德古拉睁眼特写"
→ {"type":"generate", "entities":{"target":"镜头"}, "parameters":{"character":"德古拉","action":"睁眼","shot":"特写"}}
`

    result, err := s.aiService.GenerateText(ctx, query, systemPrompt)
    if err != nil {
        return nil, err
    }

    // 解析 LLM 返回的 JSON
    var intent Intent
    if err := json.Unmarshal([]byte(result), &intent); err != nil {
        return nil, err
    }

    return &intent, nil
}
```

### 2.4 批量操作工具

**文件位置**: `application/services/agent/batch_tools.go` (新建)

```go
package agent

import (
    "context"
    "fmt"
)

type BatchTools struct {
    db    *gorm.DB
    log   *logger.Logger
    drama *DramaService
}

// BatchModifyStoryboards 批量修改分镜
func (t *BatchTools) BatchModifyStoryboards(
    ctx context.Context,
    dramaID uint,
    filter StoryboardFilter,
    updates StoryboardUpdates,
) (*BatchResult, error) {
    // 1. 查询符合条件的分镜
    var storyboards []models.Storyboard
    err := t.db.Where("drama_id = ?", dramaID).
        Where(filter.ToSQL()).
        Find(&storyboards).Error

    if err != nil {
        return nil, err
    }

    // 2. 批量更新
    results := make([]UpdateResult, 0, len(storyboards))
    for _, sb := range storyboards {
        updated := &models.Storyboard{
            ID:           sb.ID,
            ImagePrompt:  updates.ImagePrompt,
            VideoPrompt:  updates.VideoPrompt,
            Duration:     updates.Duration,
        }

        err := t.db.Model(updated).Updates(updated).Error
        results = append(results, UpdateResult{
            StoryboardID: sb.ID,
            Success:      err == nil,
            Error:        err,
        })
    }

    return &BatchResult{
        Total:     len(storyboards),
        Updated:   len(results),
        Results:   results,
    }, nil
}
```

---

## 三、前端代码调整

### 3.1 新增 Agent 侧边栏组件

**文件位置**: `web/src/components/agent/AgentSidebar.vue` (新建)

```vue
<template>
  <div class="agent-sidebar" :class="{ collapsed: isCollapsed }">
    <!-- 侧边栏头部 -->
    <div class="sidebar-header">
      <h3>🤖 AI 助手</h3>
      <el-button
        text
        @click="toggleCollapse"
        class="collapse-btn"
      >
        <el-icon>
          <component :is="isCollapsed ? ArrowRight : ArrowLeft" />
        </el-icon>
      </el-button>
    </div>

    <!-- 对话消息区域 -->
    <div class="messages-container" ref="messagesContainer">
      <div
        v-for="(message, index) in messages"
        :key="index"
        class="message-item"
        :class="message.role"
      >
        <!-- 用户指令 -->
        <div v-if="message.role === 'user'" class="user-message">
          <div class="message-content">{{ message.content }}</div>
        </div>

        <!-- Agent 回复卡片 -->
        <AgentMessageCard
          v-else
          :message="message"
          @action="handleCardAction"
        />
      </div>

      <!-- 加载中提示 -->
      <div v-if="isLoading" class="loading-message">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>思考中...</span>
      </div>
    </div>

    <!-- 快捷操作栏 -->
    <div class="quick-actions" v-if="!isCollapsed">
      <div class="action-tabs">
        <el-button
          v-for="action in quickActions"
          :key="action.key"
          size="small"
          @click="handleQuickAction(action)"
        >
          <el-icon><component :is="action.icon" /></el-icon>
          {{ action.label }}
        </el-button>
      </div>
    </div>

    <!-- 输入区域 -->
    <div class="input-area" v-if="!isCollapsed">
      <el-input
        v-model="inputText"
        type="textarea"
        :rows="2"
        placeholder="输入指令或按 ⌘K 唤醒..."
        @keydown.enter.exact="handleSend"
        @keydown.enter.shift.prevent="inputText += '\n'"
      />

      <div class="input-actions">
        <el-button-group>
          <el-button
            :icon="Microphone"
            size="small"
            @click="startVoiceInput"
          />
          <el-button
            type="primary"
            :loading="isLoading"
            :disabled="!inputText.trim()"
            @click="handleSend"
          >
            发送
          </el-button>
        </el-button-group>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import {
  ArrowLeft,
  ArrowRight,
  Loading,
  Microphone
} from '@element-plus/icons-vue'
import AgentMessageCard from './AgentMessageCard.vue'
import { agentAPI } from '@/api/agent'
import { useAgentStore } from '@/stores/agent'

// Props
const props = defineProps<{
  dramaId?: string | number
  episodeId?: string | number
}>()

// Emits
const emit = defineEmits<{
  highlight: [entity: { type: string; id: string | number }]
  navigate: [path: string]
}>()

// State
const isCollapsed = ref(false)
const inputText = ref('')
const isLoading = ref(false)
const messagesContainer = ref<HTMLElement>()

const agentStore = useAgentStore()
const { messages, addMessage, updateMessage } = agentStore

// Quick Actions
const quickActions = ref([
  { key: 'recent', label: '最近', icon: 'Clock' },
  { key: 'generate', label: '生成', icon: 'MagicStick' },
  { key: 'modify', label: '修改', icon: 'Edit' },
  { key: 'query', label: '查询', icon: 'Search' },
  { key: 'help', label: '帮助', icon: 'QuestionFilled' },
])

// Methods
const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
}

const handleSend = async () => {
  if (!inputText.value.trim() || isLoading.value) return

  const userMessage = inputText.value
  inputText.value = ''

  // 添加用户消息
  addMessage({
    role: 'user',
    content: userMessage,
    timestamp: new Date()
  })

  // 滚动到底部
  await nextTick()
  scrollToBottom()

  isLoading.value = true

  try {
    // 调用 Agent API (SSE 流式)
    const eventSource = new EventSource(
      `/api/v1/agent/chat?query=${encodeURIComponent(userMessage)}&drama_id=${props.dramaId}`
    )

    eventSource.onmessage = (event) => {
      const data = JSON.parse(event.data)

      switch (data.type) {
        case 'thinking':
          updateMessage({
            role: 'agent',
            type: 'thinking',
            content: data.content
          })
          break

        case 'tool_call':
          updateMessage({
            role: 'agent',
            type: 'tool_call',
            tool: data.tool,
            content: data.content
          })
          break

        case 'tool_result':
          updateMessage({
            role: 'agent',
            type: 'tool_result',
            tool: data.tool,
            content: data.content,
            data: data.result
          })

          // 触发主界面联动
          if (data.highlight) {
            emit('highlight', data.highlight)
          }
          break

        case 'progress':
          updateMessage({
            role: 'agent',
            type: 'progress',
            current: data.current,
            total: data.total,
            message: data.message
          })
          break

        case 'complete':
          isLoading.value = false
          eventSource.close()

          updateMessage({
            role: 'agent',
            type: 'complete',
            content: data.content
          })
          break

        case 'error':
          isLoading.value = false
          eventSource.close()

          ElMessage.error(data.message)
          break
      }

      scrollToBottom()
    }

    eventSource.onerror = () => {
      isLoading.value = false
      eventSource.close()
      ElMessage.error('连接中断')
    }

  } catch (error: any) {
    isLoading.value = false
    ElMessage.error(error.message || '发送失败')
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

const handleCardAction = (action: any) => {
  switch (action.type) {
    case 'confirm':
      // 确认操作
      executeAction(action.data)
      break
    case 'preview':
      // 预览效果
      emit('navigate', action.data.path)
      break
    case 'cancel':
      // 取消操作
      ElMessage.info('操作已取消')
      break
  }
}

const handleQuickAction = (action: any) => {
  inputText.value = action.prompt || `帮我${action.label}`
  handleSend()
}

const startVoiceInput = () => {
  // TODO: 实现语音输入
  ElMessage.info('语音输入功能开发中')
}
</script>

<style scoped>
.agent-sidebar {
  width: 360px;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-card);
  border-left: 1px solid var(--border-primary);
  transition: width 0.3s ease;
}

.agent-sidebar.collapsed {
  width: 48px;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid var(--border-primary);
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.message-item {
  margin-bottom: 16px;
}

.user-message {
  display: flex;
  justify-content: flex-end;
}

.user-message .message-content {
  max-width: 80%;
  padding: 10px 16px;
  background: var(--accent);
  color: white;
  border-radius: 12px 12px 0 12px;
}

.quick-actions {
  padding: 12px 16px;
  border-top: 1px solid var(--border-primary);
}

.action-tabs {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.input-area {
  padding: 16px;
  border-top: 1px solid var(--border-primary);
}

.input-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}
</style>
```

### 3.2 消息卡片组件

**文件位置**: `web/src/components/agent/AgentMessageCard.vue` (新建)

```vue
<template>
  <div class="agent-message-card" :class="`type-${message.type}`">
    <!-- 确认结果卡片 -->
    <div v-if="message.type === 'tool_result'" class="result-card">
      <div class="card-header">
        <el-icon color="#67c23a"><SuccessFilled /></el-icon>
        <span class="card-title">{{ getTitle(message) }}</span>
      </div>
      <div class="card-content">
        <div v-if="message.changes" class="changes-list">
          <div v-for="(change, key) in message.changes" :key="key" class="change-item">
            <span class="change-label">{{ key }}:</span>
            <span class="change-value">{{ change }}</span>
          </div>
        </div>
        <div v-if="message.preview" class="preview-area">
          <img :src="message.preview" alt="预览图" />
        </div>
      </div>
      <div class="card-actions">
        <el-button size="small" @click="$emit('action', { type: 'preview' })">
          预览
        </el-button>
        <el-button size="small" @click="$emit('action', { type: 'cancel' })">
          撤销
        </el-button>
      </div>
    </div>

    <!-- 进度卡片 -->
    <div v-else-if="message.type === 'progress'" class="progress-card">
      <div class="progress-header">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>{{ message.message || '处理中...' }}</span>
      </div>
      <el-progress
        :percentage="Math.round((message.current / message.total) * 100)"
        :format="() => `${message.current}/${message.total}`"
      />
      <div class="progress-actions">
        <el-button size="small" text @click="$emit('action', { type: 'cancel' })">
          取消
        </el-button>
      </div>
    </div>

    <!-- 预览确认卡片 -->
    <div v-else-if="message.type === 'preview_confirm'" class="preview-card">
      <div class="preview-image">
        <img :src="message.image" alt="生成结果" />
      </div>
      <p class="preview-question">符合预期吗？</p>
      <div class="preview-actions">
        <el-button type="success" @click="$emit('action', { type: 'confirm' })">
          <el-icon><Check /></el-icon>
          确认
        </el-button>
        <el-button @click="$emit('action', { type: 'regenerate' })">
          <el-icon><Refresh /></el-icon>
          重新生成
        </el-button>
        <el-button @click="$emit('action', { type: 'edit' })">
          <el-icon><Edit /></el-icon>
          编辑
        </el-button>
      </div>
    </div>

    <!-- 主动建议卡片 -->
    <div v-else-if="message.type === 'suggestion'" class="suggestion-card">
      <div class="suggestion-header">
        <el-icon color="#e6a23c"><WarningFilled /></el-icon>
        <span>{{ message.title }}</span>
      </div>
      <p class="suggestion-content">{{ message.content }}</p>
      <div class="suggestion-actions">
        <el-button type="primary" size="small" @click="$emit('action', { type: 'confirm', data: message.data })">
          {{ message.confirmText || '立即处理' }}
        </el-button>
        <el-button size="small" text @click="$emit('action', { type: 'ignore' })">
          忽略
        </el-button>
      </div>
    </div>

    <!-- 思考卡片 -->
    <div v-else-if="message.type === 'thinking'" class="thinking-card">
      <el-icon class="is-loading"><Loading /></el-icon>
      <span>{{ message.content || '思考中...' }}</span>
    </div>

    <!-- 默认文本消息 -->
    <div v-else class="text-message">
      {{ message.content }}
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  SuccessFilled,
  Loading,
  Check,
  Refresh,
  Edit,
  WarningFilled
} from '@element-plus/icons-vue'

interface Message {
  type: string
  content?: string
  [key: string]: any
}

const props = defineProps<{
  message: Message
}>()

const emit = defineEmits<{
  action: [data: any]
}>()

const getTitle = (message: Message) => {
  const titles: Record<string, string> = {
    generate: '生成完成',
    modify: '修改完成',
    batch_modify: '批量修改完成'
  }
  return titles[message.tool] || '操作完成'
}
</script>

<style scoped>
.agent-message-card {
  max-width: 90%;
}

.result-card,
.preview-card,
.suggestion-card {
  background: white;
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  padding: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.card-header,
.suggestion-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  font-weight: 600;
}

.card-actions,
.preview-actions,
.suggestion-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
  justify-content: flex-end;
}

.progress-card {
  padding: 12px;
  background: #f5f7fa;
  border-radius: 8px;
}

.progress-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.thinking-card {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  color: var(--text-secondary);
  font-size: 14px;
}

.preview-image img {
  width: 100%;
  border-radius: 4px;
  margin-bottom: 8px;
}

.preview-question {
  text-align: center;
  margin: 8px 0;
  color: var(--text-primary);
}
</style>
```

### 3.3 主界面集成

**文件位置**: `web/src/views/drama/DramaWorkflow.vue`

需要调整的部分：

```vue
<template>
  <div class="workflow-container">
    <AppHeader :fixed="false" :show-logo="false">
      <!-- 保持现有头部不变 -->
    </AppHeader>

    <!-- 主内容区域调整为两列布局 -->
    <div class="main-content-wrapper">
      <!-- 左侧：现有工作流内容 -->
      <div class="main-area" :class="{ 'with-sidebar': showAgentSidebar }">
        <!-- 保持现有 stage-area 内容不变 -->
        <div class="stage-area">
          <!-- 现有的 3 个阶段卡片 -->
        </div>
      </div>

      <!-- 右侧：Agent 侧边栏 -->
      <AgentSidebar
        v-if="showAgentSidebar"
        :drama-id="drama?.id"
        :episode-id="currentEpisode?.id"
        @highlight="handleAgentHighlight"
        @navigate="handleAgentNavigate"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import AgentSidebar from '@/components/agent/AgentSidebar.vue'

// 新增状态
const showAgentSidebar = ref(true)

// 处理 Agent 高亮联动
const handleAgentHighlight = (entity: { type: string; id: string | number }) => {
  switch (entity.type) {
    case 'storyboard':
      // 高亮对应分镜
      highlightStoryboard(entity.id)
      break
    case 'character':
      // 高亮对应角色
      highlightCharacter(entity.id)
      break
  }
}

// 处理 Agent 导航
const handleAgentNavigate = (path: string) => {
  router.push(path)
}
</script>

<style scoped>
.main-content-wrapper {
  display: flex;
  width: 100%;
  height: calc(100vh - 56px);
}

.main-area {
  flex: 1;
  overflow-y: auto;
  transition: all 0.3s ease;
}

.main-area.with-sidebar {
  max-width: calc(100% - 360px);
}

/* 高亮样式 */
.highlight-storyboard {
  animation: pulse 1s ease-in-out;
  border: 2px solid var(--accent) !important;
}

@keyframes pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(64, 158, 255, 0.4); }
  50% { box-shadow: 0 0 0 8px rgba(64, 158, 255, 0); }
}
</style>
```

### 3.4 事件总线（跨组件通信）

**文件位置**: `web/src/utils/eventBus.ts` (新建)

```typescript
import { EventEmitter } from 'events'

class EventBus extends EventEmitter {
  on(event: string, listener: (...args: any[]) => void): this {
    return super.on(event, listener)
  }

  emit(event: string, ...args: any[]): boolean {
    return super.emit(event, ...args)
  }
}

export const agentEventBus = new EventBus()

// 事件类型定义
export enum AgentEvents {
  HIGHLIGHT_STORYBOARD = 'highlight:storyboard',
  HIGHLIGHT_CHARACTER = 'highlight:character',
  JUMP_TO_TIMELINE = 'jump:timeline',
  PLAY_PREVIEW = 'play:preview',
  REFRESH_DATA = 'refresh:data',
}
```

### 3.5 Pinia Store

**文件位置**: `web/src/stores/agent.ts` (新建)

```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface AgentMessage {
  id?: string
  role: 'user' | 'agent'
  type: string
  content: string
  timestamp: Date
  [key: string]: any
}

export const useAgentStore = defineStore('agent', () => {
  const messages = ref<AgentMessage[]>([])
  const contextId = ref<string>('')
  const isConnected = ref(false)

  function addMessage(message: AgentMessage) {
    messages.value.push({
      ...message,
      id: `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    })
  }

  function updateMessage(partial: Partial<AgentMessage>) {
    const lastMessage = messages.value[messages.value.length - 1]
    if (lastMessage) {
      Object.assign(lastMessage, partial)
    }
  }

  function clearMessages() {
    messages.value = []
  }

  function setContextId(id: string) {
    contextId.value = id
  }

  return {
    messages,
    contextId,
    isConnected,
    addMessage,
    updateMessage,
    clearMessages,
    setContextId
  }
})
```

### 3.6 Agent API 客户端

**文件位置**: `web/src/api/agent.ts` (新建)

```typescript
import request from '@/utils/request'

export interface ChatRequest {
  query: string
  context_id?: string
  drama_id?: string | number
  episode_id?: string | number
  metadata?: Record<string, any>
}

// SSE 流式聊天
export function chatStream(params: ChatRequest): EventSource {
  const queryString = new URLSearchParams({
    query: params.query,
    context_id: params.context_id || '',
    drama_id: String(params.drama_id || ''),
    episode_id: String(params.episode_id || '')
  })

  return new EventSource(`/api/v1/agent/chat?${queryString}`)
}

// 提交反馈
export function submitFeedback(data: {
  message_id: string
  feedback: 'positive' | 'negative'
  comment?: string
}) {
  return request.post('/agent/feedback', data)
}

// 获取对话上下文
export function getContext(contextId: string) {
  return request.get(`/agent/context/${contextId}`)
}
```

---

## 四、新增文件清单

### 4.1 后端新增文件

```
api/handlers/
└── agent.go                          # Agent 处理器

application/services/agent/
├── intent_service.go                  # 意图识别服务
├── batch_tools.go                     # 批量操作工具
├── context_service.go                 # 对话上下文管理
└── suggestion_service.go              # 主动建议服务

pkg/agent/
├── parser/
│   ├── intent_parser.go              # 意图解析器
│   └── entity_extractor.go            # 实体提取器
└── llm/
    └── prompt_template.go             # LLM Prompt 模板
```

### 4.2 前端新增文件

```
web/src/
├── components/agent/
│   ├── AgentSidebar.vue               # Agent 侧边栏主组件
│   ├── AgentMessageCard.vue           # 消息卡片组件
│   ├── QuickActions.vue               # 快捷操作组件
│   └── VoiceInput.vue                 # 语音输入组件
├── stores/
│   └── agent.ts                       # Agent Pinia Store
├── api/
│   └── agent.ts                       # Agent API
├── utils/
│   └── eventBus.ts                    # 事件总线
└── types/
    └── agent.ts                       # Agent 类型定义
```

---

## 五、实施步骤

### Phase 1: 基础设施（1-2 天）

**目标**: 建立 Agent 对话框架

- [ ] 后端：添加 `agent.go` Handler
- [ ] 后端：实现 SSE 流式响应
- [ ] 前端：创建 `AgentSidebar.vue` 组件
- [ ] 前端：集成到 `DramaWorkflow.vue`
- [ ] 前端：实现基本的消息收发

**验证**: 能在侧边栏发送消息并收到响应

### Phase 2: 意图识别（2-3 天）

**目标**: 理解用户自然语言指令

- [ ] 后端：实现 `IntentService.ParseIntent`
- [ ] 后端：定义意图类型和参数格式
- [ ] 后端：实现规则匹配（MVP）
- [ ] 前端：添加快捷指令按钮
- [ ] 前端：实现指令历史记录

**验证**: 能识别"生成"、"修改"、"查询"等基本指令

### Phase 3: 批量操作（2-3 天）

**目标**: 实现批量修改功能

- [ ] 后端：实现 `BatchTools.BatchModifyStoryboards`
- [ ] 后端：添加进度反馈机制
- [ ] 前端：实现进度卡片组件
- [ ] 前端：实现预览和确认功能
- [ ] 前端：支持撤销操作

**验证**: "所有德古拉镜头改成冷色调"能批量执行

### Phase 4: 双向联动（2 天）

**目标**: Agent 对话 ↔ 主界面联动

- [ ] 前端：实现高亮联动
- [ ] 前端：实现跳转联动
- [ ] 前端：实现数据刷新联动
- [ ] 后端：推送实体关联信息

**验证**: Agent 提到"镜头2"时，主界面自动高亮

### Phase 5: 主动建议（2-3 天）

**目标**: Agent 检测问题并主动建议

- [ ] 后端：实现上下文分析
- [ ] 后端：实现建议触发逻辑
- [ ] 前端：实现建议卡片组件
- [ ] 前端：支持一键处理建议

**验证**: 检测到节奏问题时主动提示添加过渡

### Phase 6: 语音输入（1-2 天）

**目标**: 支持语音输入指令

- [ ] 前端：集成 Web Speech API
- [ ] 前端：实现语音识别
- [ ] 前端：添加麦克风按钮
- [ ] 后端：支持语音转文本的意图识别

**验证**: 按住麦克风说话能转为文本指令

---

## 六、测试验证

### 6.1 功能测试清单

| 功能 | 测试用例 | 预期结果 |
|-----|---------|---------|
| **基础对话** | 发送"帮我生成一个分镜" | Agent 理解并执行 |
| **批量修改** | "所有德古拉镜头改成冷色调" | 批量更新并反馈进度 |
| **高亮联动** | Agent 提到"镜头2" | 主界面镜头2高亮 |
| **进度反馈** | 批量生成中 | 实时显示进度条 |
| **撤销操作** | 点击"撤销"按钮 | 操作回退，状态恢复 |
| **快捷指令** | 点击"生成"快捷按钮 | 自动填充并发送指令 |

### 6.2 性能指标

| 指标 | 目标值 | 测试方法 |
|-----|-------|---------|
| 响应延迟 | < 500ms | 计录发送到首字节时间 |
| 意图识别准确率 | > 90% | 测试 100 个常见指令 |
| 批量操作速度 | 50 个分镜 < 10s | 性能测试 |
| 内存占用 | < 100MB | Chrome DevTools |

### 6.3 兼容性测试

| 浏览器 | 版本 | 状态 |
|-------|------|------|
| Chrome | 120+ | ✅ 必须支持 |
| Edge | 120+ | ✅ 支持 |
| Safari | 17+ | ⚠️ 尽力支持 |
| Firefox | 120+ | ⚠️ 尽力支持 |

---

## 附录

### A. SSE 事件格式定义

```typescript
// Agent → 前端的事件类型
interface AgentEvent {
  // 思考事件
  thinking: {
    type: 'thinking'
    content: string
  }

  // 工具调用事件
  tool_call: {
    type: 'tool_call'
    tool: string
    content: string
  }

  // 工具结果事件
  tool_result: {
    type: 'tool_result'
    tool: string
    content: string
    result: any
    highlight?: {
      type: string
      id: string | number
    }
  }

  // 进度事件
  progress: {
    type: 'progress'
    current: number
    total: number
    message: string
  }

  // 完成事件
  complete: {
    type: 'complete'
    content: string
  }

  // 错误事件
  error: {
    type: 'error'
    code: string
    message: string
  }
}
```

### B. 意图类型定义

```typescript
enum IntentType {
  BATCH_MODIFY = 'batch_modify',    // 批量修改
  GENERATE = 'generate',             // 生成内容
  QUERY = 'query',                  // 查询信息
  SUGGEST = 'suggest',              // 主动建议
  MODIFY = 'modify',                // 单个修改
  DELETE = 'delete',                // 删除
  EXPORT = 'export',                // 导出
}

interface Intent {
  type: IntentType
  entities: {
    target: string    // 操作对象：镜头/角色/场景
    scope: string    // 操作范围：单个/批量/全部
    filter: Record<string, any>  // 过滤条件
  }
  parameters: {
    [key: string]: any
  }
  confidence: number  // 识别置信度
}
```

### C. 关键代码位置映射

| 功能模块 | 当前文件 | 调整后文件 | 变更类型 |
|---------|---------|-----------|---------|
| 剧本工作流 | `DramaWorkflow.vue` | `DramaWorkflow.vue` | 修改（集成侧边栏） |
| 分镜编辑 | `StoryboardEdit.vue` | `StoryboardEdit.vue` | 修改（添加高亮） |
| 消息组件 | - | `AgentMessageCard.vue` | 新增 |
| Agent API | - | `agent.go` | 新增 |
| 意图识别 | - | `intent_service.go` | 新增 |
| 批量操作 | `storyboard_service.go` | `batch_tools.go` | 新增 |

---

**文档结束**

如有疑问，请参考设计文档 `docs/agent.desing` 或提交 Issue。
