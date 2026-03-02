# 依赖分析与 Eino 方案评估

## 1. 当前项目依赖

### 核心技术栈

```
github.com/drama-generator/backend
├── Web 框架
│   └── gin-gonic/gin v1.9.1
├── 数据层
│   ├── gorm/gorm v1.30.0
│   ├── gorm/driver/mysql v1.5.2
│   ├── gorm/driver/sqlite v1.6.0
│   └── gorm/datatypes v1.2.0
├── 配置与日志
│   ├── spf13/viper v1.17.0
│   └── uber.org/zap v1.26.0
└── 工具库
    ├── google/uuid v1.6.0
    └── robfig/cron/v3 v3.0.1
```

### 特点

- ✅ **轻量级**：依赖数量少，总间接依赖约 30 个
- ✅ **成熟稳定**：都是 Go 生态中的主流库
- ✅ **简单直接**：无 AI Agent 框架，手动编排业务流程

---

## 2. Eino 方案依赖

### 核心框架 (CloudWeGo Eino)

```
github.com/cloudwego/eino v0.9.0-alpha.2
```

### 扩展组件

```go
github.com/cloudwego/eino-ext/
├── components/model/      # LLM 模型适配
│   ├── openai v0.1.5
│   ├── deepseek
│   ├── ollama v0.1.6
│   ├── ark v0.1.45
│   ├── agenticark v0.1.0-alpha.1
│   └── gemini, qwen, claude 等
├── components/tool/       # 工具集成
│   ├── duckduckgo/v2
│   ├── commandline
│   ├── mcp/officialmcp v0.1.0
│   ├── searxng
│   └── bingsearch, wikipedia 等
├── components/retriever/  # 检索器
│   └── volc_vikingdb
└── callbacks/            # 回调与可观测性
    ├── cozeloop v0.1.6
    └── apmplus
```

### 支持设施

```
├── 会话存储
│   └── redis/go-redis/v9 v9.17.2
├── 文档解析
│   ├── parser/html
│   └── parser/pdf
├── 测试工具
│   ├── alicebob/miniredis/v2 v2.35.0
│   └── stretchr/testify v1.11.1
└── 其他
    ├── chromedp/chromedp v0.9.5         # 浏览器自动化
    ├── modelcontextprotocol/go-sdk      # MCP 协议
    └── wk8/go-ordered-map/v2            # 有序 Map
```

### 特点

- ✅ **AI 原生**：专为 LLM Agent 设计
- ✅ **模型丰富**：支持 10+ 主流 LLM
- ✅ **工具生态**：内置搜索、浏览器、MCP 等工具
- ⚠️ **依赖较多**：总间接依赖约 80+ 个
- ⚠️ **学习曲线**：需要理解 Agent/Tool/Chain 概念

---

## 3. 方案对比

### 架构复杂度

| 指标 | 当前架构 | Eino 方案 |
|------|----------|-----------|
| **代码量** | 100% | ~60-70% |
| **Service 文件数** | 21 个 | ~8 个 Tools |
| **编排方式** | 手动协调 | Agent 自动规划 |
| **错误恢复** | 手动重试 | 自动 Replan |
| **用户体验** | 多次 API 调用 | 自然语言对话 |

### 依赖对比

| 类型 | 当前架构 | Eino 方案 |
|------|----------|-----------|
| **Web 框架** | Gin | 保留 Gin |
| **ORM** | GORM | 保留 GORM |
| **AI 框架** | ❌ 无 | ✅ Eino |
| **LLM 适配** | 自定义封装 | Eino 内置 |
| **工具系统** | ❌ 无 | ✅ Eino Tool |
| **会话管理** | 自定义 | Eino + Redis |
| **总依赖数** | ~30 | ~100 |

---

## 4. 迁移映射

### Service → Tool 映射

| 现有 Service | Eino Tool | 说明 |
|-------------|-----------|------|
| `drama_service.go` | 保留 Biz 层 | 核心业务逻辑保留 |
| `storyboard_service.go` | `generate_storyboard` | 生成分镜 Tool |
| `image_generation_service.go` | `generate_images` | 批量生成图片 Tool |
| `video_generation_service.go` | `generate_videos` | 批量生成视频 Tool |
| `video_merge_service.go` | `merge_video` | 合成视频 Tool |
| `character_service.go` | `extract_characters` | 提取角色 Tool |
| `scene_service.go` | `extract_scenes` | 提取场景 Tool |

### 目录结构变化

```
现有结构                          →  Eino 方案结构
─────────────────────────────────────────────────────────────
api/handlers/                    →  internal/handler/http/
application/services/            →  internal/biz/agent/tools/
domain/models/                   →  internal/model/
pkg/ai/                          →  internal/biz/agent/pkg/ai/
pkg/image/                       →  internal/biz/agent/pkg/image/
pkg/video/                       →  internal/biz/agent/pkg/video/
                                 →  internal/biz/agent/
                                     ├── agent.go      # Agent 定义
                                     └── runtime.go    # 运行时
```

---

## 5. Agent 设计示例

### DramaProduction Agent

```go
const DramaProductionSystemPrompt = `你是短剧制作专家（Drama Production Agent）。

### 角色
- 理解用户的短剧创作需求
- 将短剧制作任务分解为清晰的执行步骤
- 按步骤执行，跟踪进度，动态调整

### 工作流程
1. **规划阶段**：分析需求，分解任务
2. **执行阶段**：调用工具完成各步骤
3. **反思阶段**：失败分析，调整计划

### 可用工具
- generate_script: 生成剧本大纲和角色
- extract_scenes: 提取场景
- generate_storyboard: 生成分镜
- generate_images: 批量生成图片
- generate_videos: 批量生成视频
- merge_video: 合成最终视频
`
```

### Tool 定义示例

```go
// generate_images Tool
type GenerateImagesInput struct {
    EpisodeID       uint   `json:"episode_id"`
    ConcurrentLimit int    `json:"concurrent_limit"` // 并发数
}

type GenerateImagesOutput struct {
    Generated    []ImageGeneration `json:"generated"`
    Failed       []ImageGeneration `json:"failed"`
    TotalCount   int               `json:"total_count"`
    SuccessCount int               `json:"success_count"`
}
```

---

## 6. 实施建议

### 采用 Eino 的优势

1. **自动化编排**：Agent 自动规划执行步骤，无需手动协调
2. **错误恢复**：内置 Replan 机制，失败自动调整策略
3. **用户体验**：从多次 API 调用变为自然语言对话
4. **生态丰富**：内置 LLM、工具、检索器等组件
5. **流式输出**：SSE 实时推送执行进度

### 需要注意的点

1. **学习曲线**：需要理解 Eino 的 Agent/Tool/Chain 概念
2. **依赖增加**：引入约 70 个额外依赖包
3. **状态管理**：需要实现 SSE 流式事件推送
4. **调试复杂度**：Agent 思维链调试比传统代码复杂

### 实施步骤

```
Phase 1: 基础设施 (1-2天)
├── 初始化项目结构
├── 配置 Eino 依赖
└── 复制现有 Model 层

Phase 2: Tools 开发 (3-4天)
├── generate_script
├── extract_scenes
├── generate_storyboard
├── generate_images
├── generate_videos
└── merge_video

Phase 3: Agent 集成 (2-3天)
├── 定义 DramaProduction Agent
├── 实现 Plan-Execute 逻辑
└── 错误处理和 Replan

Phase 4: API 层 (2天)
├── 实现 SSE Chat Handler
├── 保留现有 REST API
└── 事件处理器

Phase 5: 前端适配 (2-3天)
├── Agent 对话界面
├── 进度可视化
└── 完整流程测试
```

---

## 7. 参考资源

### 官方文档

- [Eino GitHub](https://github.com/cloudwego/eino)
- [Eino Examples](https://github.com/cloudwego/eino-examples)
- [Next-Show 迁移方案](./NEXT_SHOW_MIGRATION.md)

### 示例代码位置

```
docs/next-show/docs/eino-examples/
├── adk/
│   ├── intro/               # 入门示例
│   ├── human-in-the-loop/   # 人机协作
│   └── multiagent/          # 多 Agent
└── compose/                 # 组合式 API
```

---

## 8. 用户体验对比：当前架构 vs Eino Agent

### 什么是"手动"？

**"手动"** 指的是前端需要按顺序调用多个 API，用户需要多次点击才能完成一个短剧制作。

### 当前架构：前端需要调用 9+ 个步骤

```
用户想制作一个短剧，需要：

┌─────────────────────────────────────────────────────────────┐
│              前端需要调用的 API（按顺序）                      │
└─────────────────────────────────────────────────────────────┘

1. POST /api/v1/dramas
   └─ 创建剧本

2. PUT /api/v1/dramas/:id/outline
   └─ 保存大纲

3. POST /api/v1/generation/characters
   └─ 生成角色（异步，需要轮询）
   └─ GET /api/v1/tasks/:id  ← 轮询状态

4. POST /api/v1/generation/episodes
   └─ 生成章节

5. POST /api/v1/storyboards/generate
   └─ 为每集生成分镜

6. POST /api/v1/images
   └─ 批量生成场景图（可能几十次调用）

7. POST /api/v1/images
   └─ 批量生成角色图

8. POST /api/v1/videos
   └─ 为每个分镜生成视频（可能几十次调用）
   └─ GET /api/v1/tasks/:id  ← 轮询每个视频状态

9. POST /api/v1/video-merges
   └─ 合成最终视频
   └─ GET /api/v1/video-merges/:id  ← 轮询合成状态

───────────────────────────────────────────────────────────────
总计：**9+ 个步骤，可能 50-100 次 API 调用**
```

### Eino Agent 方案：一次调用完成

```javascript
// 前端只需一次调用
POST /api/v1/chat
{
  "query": "帮我制作一个3集古风仙侠短剧，主角是白浅和夜华"
}

// 后端 Agent 自动完成上述所有步骤
// 并通过 SSE 实时推送进度
```

### 前端代码对比

#### 当前架构（伪代码）

```typescript
// 需要编写复杂的状态管理和步骤编排
async function createDrama() {
  // 1. 创建剧本
  const drama = await api.post('/api/v1/dramas', { title: '我的短剧' });

  // 2. 保存大纲
  await api.put(`/api/v1/dramas/${drama.id}/outline`, { outline: '...' });

  // 3. 生成角色（异步）
  const charTask = await api.post('/api/v1/generation/characters', { drama_id: drama.id });
  await pollTask(charTask.task_id); // 需要轮询

  // 4. 生成章节
  await api.post('/api/v1/generation/episodes', { drama_id: drama.id });

  // 5-9. ... 后续步骤
}
```

#### Eino Agent 方案（伪代码）

```typescript
// 一个聊天框 + SSE 监听
const chatClient = new ChatClient();

chatClient.on('message', (msg) => {
  console.log(msg.content); // 实时进度
});

chatClient.send("帮我制作一个3集古风仙侠短剧");
```

### 对比总结

```
┌─────────────────────────────────────────────────────────────┐
│                        当前架构                              │
├─────────────────────────────────────────────────────────────┤
│  前端代码：需要编写复杂的状态管理和步骤编排                   │
│  用户操作：需要9次点击，等待每步完成                         │
│  API 调用：50-100 次                                         │
│  用户体验：😓 繁琐                                           │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                      Eino Agent 方案                         │
├─────────────────────────────────────────────────────────────┤
│  前端代码：一个聊天框 + SSE 监听                             │
│  用户操作：一句话描述需求                                    │
│  API 调用：1 次                                             │
│  用户体验：😊 便捷                                           │
└─────────────────────────────────────────────────────────────┘
```

### SSE 流式输出示例

```javascript
// Agent 执行过程中实时推送进度
data: {"type":"thinking","content":"分析需求：仙侠恋题材，3集..."}
data: {"type":"tool_call","tool":"generate_script","content":"开始生成剧本"}
data: {"type":"tool_result","tool":"generate_script","content":"已生成剧本大纲"}
data: {"type":"tool_call","tool":"extract_characters","content":"提取角色中..."}
data: {"type":"tool_result","tool":"extract_characters","content":"已提取 5 个角色"}
data: {"type":"tool_call","tool":"generate_storyboard","content":"第1集分镜生成中..."}
data: {"type":"progress","value":"3/12","content":"分镜进度 3/12"}
data: {"type":"tool_call","tool":"generate_images","content":"图片生成中（并发5）"}
data: {"type":"answer","content":"短剧制作完成！视频已生成"}
```

---

## 9. 私域运营场景适用性分析

### 为什么 Agent 方案更适合私域运营？

**私域运营** 指企业通过自有渠道（微信、企微、APP）直接触达和服务用户，Agent 对话方式是此场景的最佳选择。

### 典型私域运营场景

```
┌─────────────────────────────────────────────────────────────┐
│                   私域运营核心场景                            │
├─────────────────────────────────────────────────────────────┤
│  1. 用户主动咨询                                             │
│     └─ "帮我生成一个XX产品的宣传视频"                         │
│                                                             │
│  2. 运营人员使用                                             │
│     └─ "为下周的活动准备10个素材"                            │
│                                                             │
│  3. 批量自动化                                               │
│     └─ "根据这份产品列表生成对应的宣传文案"                   │
│                                                             │
│  4. 实时响应需求                                             │
│     └─ 用户改需求 → Agent 动态调整                           │
└─────────────────────────────────────────────────────────────┘
```

### 场景对比：批量素材生成

```
任务：为 50 个商品生成推广素材

多 API 方案：
┌─────────────────────────────────────────────┐
│  操作员需要：                                │
│  1. 导入商品列表（50个）                     │
│  2. 逐个选择商品                            │
│  3. 点击"生成文案" → 等待                    │
│  4. 点击"生成图片" → 等待                    │
│  5. 点击"合成视频" → 等待                    │
│  6. 重复 2-5 步 50 次                       │
│                                             │
│  耗时：约 2-3 小时                           │
│  体验：😫 疲劳操作                           │
└─────────────────────────────────────────────┘

Agent 方案：
┌─────────────────────────────────────────────┐
│  操作员只需：                                │
│  1. 上传商品列表                             │
│  2. 发送："为这50个商品生成推广视频"          │
│  3. 喝咖啡 ☕                                │
│  4. 等待 Agent 批量完成                      │
│                                             │
│  耗时：约 10-15 分钟                         │
│  体验：😊 自动化处理                          │
└─────────────────────────────────────────────┘
```

### 私域运营核心优势

| 指标 | 多 API 方案 | Agent 方案 |
|-----|-----------|-----------|
| **响应时间** | 5-10 分钟 | 实时反馈 |
| **操作门槛** | 需要培训 | 零门槛 |
| **人效提升** | 1x | **10x+** |
| **可扩展性** | 人力线性增长 | 自动化扩展 |
| **用户满意度** | 😐 一般 | 😊 高 |

### 规模化能力对比

```
私域运营规模扩展：

10 个用户/天    →  两种方案差异不大
100 个用户/天   →  Agent 优势显现
1000 个用户/天  →  Agent 必不可少
10000+ 个用户/天 →  只有 Agent 能支撑
```

### 结论：私域运营评分

| 评估维度 | 评分 |
|---------|------|
| **适用性** | ⭐⭐⭐⭐⭐ (完美匹配) |
| **降本增效** | ⭐⭐⭐⭐⭐ (10倍+提升) |
| **用户体验** | ⭐⭐⭐⭐⭐ (零门槛) |
| **可扩展性** | ⭐⭐⭐⭐⭐ (易规模化) |

**强烈推荐**：私域运营应用采用 Eino Agent 方案。

---

## 10. 产品定位：类似可灵 AI

### 与可灵 AI (Kling AI) 的对比

采用 Eino Agent 后，你的项目就是 **"短剧版可灵 AI"**：

```
┌─────────────────────────────────────────────────────────────┐
│                    可灵 AI (Kling AI)                        │
├─────────────────────────────────────────────────────────────┤
│  用户：帮我生成一个古风仙侠视频                               │
│  Agent：分析需求 → 生成剧本 → 生成视频 → 输出                 │
│  体验：一句话搞定                                            │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│              Huobao Drama (Eino Agent 方案)                  │
├─────────────────────────────────────────────────────────────┤
│  用户：帮我制作一个3集古风仙侠短剧                            │
│  Agent：分析需求 → 生成剧本 → 生成分镜 → 生成图片 → 生成视频   │
│  体验：一句话搞定                                            │
└─────────────────────────────────────────────────────────────┘
```

### 核心差异：你更聚焦 **短剧全流程**

```
可灵 AI：
├─ 单个视频生成
├─ 文生视频 / 图生视频
└─ 适用于：短视频、广告片

你的项目：
├─ 完整短剧制作
├─ 剧本 → 角色 → 分镜 → 场景图 → 角色图 → 视频 → 合成
├─ 多集连续内容
└─ 适用于：系列短剧、连续剧
```

### 功能对比表

| 功能 | 可灵 AI | 你的项目 (Eino) |
|-----|---------|-----------------|
| **对话式交互** | ✅ | ✅ |
| **剧本生成** | ✅ | ✅ |
| **角色管理** | ❌ | ✅ (复用角色) |
| **多集管理** | ❌ | ✅ |
| **分镜制作** | ❌ | ✅ |
| **场景提取** | ❌ | ✅ |
| **批量生成** | 部分 | ✅ (并发生成) |
| **视频合成** | ❌ | ✅ (自动拼接) |

### 你的独特竞争力

```
┌─────────────────────────────────────────────────────────────┐
│                     你独有的竞争力                            │
├─────────────────────────────────────────────────────────────┤
│  1. 完整性                                                   │
│     └─ 从创意到成片的全流程自动化                            │
│                                                             │
│  2. 连续性                                                   │
│     └─ 多集短剧的角色、场景自动复用和连贯性                  │
│                                                             │
│  3. 可控性                                                   │
│     └─ 分镜级别的精细控制（AI辅助+人工调整）                 │
│                                                             │
│  4. 批量化                                                   │
│     └─ 一句话生成完整系列，而非单个视频                      │
└─────────────────────────────────────────────────────────────┘
```

### 用户使用场景对比

#### 可灵 AI 用户

```
用户：生成一个仙侠视频

结果：
├─ 1 个视频片段
├─ 时长 5-10 秒
└─ 独立内容

适用场景：
├─ 短视频创作
├─ 广告片
└─ 社交媒体素材
```

#### 你的项目用户

```
用户：制作一个3集仙侠短剧

结果：
├─ 3 个完整剧集
├─ 每集 2-3 分钟
├─ 连贯的角色和剧情
├─ 完整的分镜脚本
└─ 合成的最终视频

适用场景：
├─ 系列短剧
├─ 连续剧
├─ 内容工作室
└─ 批量内容生产
```

### 产品定位建议

```
┌─────────────────────────────────────────────────────────────┐
│                   产品 Slogan 对比                           │
├─────────────────────────────────────────────────────────────┤
│  可灵 AI：                                                   │
│  "用 AI 创造想象中的视频"                                    │
│                                                             │
│  你的项目：                                                   │
│  "用 AI 一句话制作完整短剧"                                  │
│  或                                                         │
│  "短剧制作，从创作到成片的 AI 全流程"                        │
└─────────────────────────────────────────────────────────────┘
```

### 核心价值主张

```
可灵 AI：    "让每个人都能创造想象中的视频"
你的项目：   "让每个人都能一键制作完整短剧"

差异点：
├─ 可灵 AI 专注：单视频生成
└─ 你的项目专注：完整短剧系列

这构成了差异化竞争优势！
```

---

## 11. 总结

| 场景 | 推荐方案 |
|------|----------|
| 快速验证原型 | 当前架构 |
| 复杂流程自动化 | Eino 方案 |
| 需要自然语言交互 | Eino 方案 |
| 追求最小依赖 | 当前架构 |
| 长期维护性 | Eino 方案 |

**最终建议**：如果你的项目需要复杂的 AI 编排能力和更好的用户体验，采用 Eino 方案是值得的投资。
