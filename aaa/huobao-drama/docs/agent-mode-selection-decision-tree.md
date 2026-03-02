# Agent Mode Selection Decision Tree

**Extracted:** 2025-02-01
**Context:** 选择合适的 Agent 模式（Plan-Execute vs Supervisor vs Deep）时

## Problem

随着 Agent 应用发展，开发者面临三种主流模式：
- **Plan-Execute**: 计划执行模式
- **Supervisor**: 监督者模式（多专家协作）
- **Deep**: 深度研究模式

不知道该选择哪种模式来解决具体问题。

## Solution

使用决策树根据任务特性选择合适的 Agent 模式：

### 1. Plan-Execute Agent（计划执行）

**判断标准：**
- ✅ 任务有明确的执行顺序
- ✅ 需要展示进度和步骤
- ✅ 失败后需要调整计划继续执行

**适用场景：**
```
- 多步骤的自动化流程（如：数据清洗 → 分析 → 报告）
- 需要中间检查点的长任务
- 确定性的生产流程（如：剧本 → 分镜 → 图片 → 视频）
```

**特点：**
- 透明性：展示已完成/进行中/待执行步骤
- 有序性：按步骤顺序执行
- 可调整：失败时可以 replan

### 2. Supervisor Agent（监督者）

**判断标准：**
- ✅ 任务需要多个领域专家协作
- ✅ 不确定具体需要哪些专家
- ✅ 子任务相对独立

**适用场景：**
```
- 知识库检索 + 数据分析组合
- 多领域专家协作（如：法律 + 财务 + 技术）
- 需要动态选择专家的开放任务
```

**特点：**
- 协作性强：擅长多专家协调
- 灵活委派：动态选择专家
- 无显式计划：直接委派执行

### 3. Deep Agent（深度研究）

**判断标准：**
- ✅ 需要深入调研复杂问题
- ✅ 需要多角度分析和交叉验证
- ✅ 需要综合多个来源的信息

**适用场景：**
```
- 需要多轮迭代的复杂问题研究
- 需要交叉验证的分析任务
- 综合多个来源的调研报告
```

**特点：**
- 研究导向：多轮迭代检索
- 交叉验证：对比多个来源
- 反思机制：检查答案完整性

### 决策流程图

```
                   ┌─────────────────┐
                   │   用户任务需求   │
                   └────────┬────────┘
                            │
                            ▼
                   ┌─────────────────┐
                   │ 是否有明确顺序？ │
                   └────────┬────────┘
                     是 │         │ 否
                        │         ▼
                        │    ┌─────────────────┐
                        │    │ 需要多专家协作？ │
                        │    └────────┬────────┘
                        │      是 │         │ 否
                        │         │         ▼
                        │         │    ┌─────────────────┐
                        │         │    │需要深度研究？   │
                        │         │    └────────┬────────┘
                        │         │      是 │         │ 否
                        ▼         ▼         │         │
                ┌───────────┐ ┌─────────┐  │         ▼
                │Plan-Execute│ │Supervisor│ │       ┌─────────┐
                │   Agent    │ │  Agent  │  │       │  Deep   │
                └───────────┘ └─────────┘  │       │  Agent  │
                                          │       └─────────┘
                                          │
                                    (考虑混合模式)
```

## Example

### 短剧制作项目的混合应用

```go
// 主控：Plan-Execute Agent
MainAgent = Plan-Execute

// 子任务分工：
1. 剧本大纲生成 → Deep Agent (多角度创意研究)
2. 场景提取     → Tool (直接调用)
3. 分镜生成     → Plan-Execute Agent (有序生成)
4. 图片生成     → Supervisor Agent (委派给多个图片服务商)
5. 视频生成     → Supervisor Agent (委派给多个视频服务商)
6. 视频合成     → Tool (FFmpeg 直接调用)
```

### 代码示例（Eino Framework）

```go
// Plan-Execute 配置
planExecuteAgent := &Agent{
    AgentType: AgentTypePlanExecute,
    MaxIterations: 50,
    Config: map[string]any{
        "replan_enabled": true,
        "allowed_tools": []string{
            "transfer_task",
            "thinking",
            "todo_write",
        },
    },
}

// Supervisor 配置
supervisorAgent := &Agent{
    AgentType: AgentTypeSupervisor,
    MaxIterations: 20,
    Config: map[string]any{
        "default_sub_agents": []string{
            "rag_agent",
            "data_analyst",
        },
    },
}

// Deep 配置
deepAgent := &Agent{
    AgentType: AgentTypeDeep,
    MaxIterations: 50,
    Config: map[string]any{
        "reflection_enabled": true,
        "max_research_rounds": 5,
    },
}
```

## When to Use

在以下情况触发此技能：
- 🤔 需要设计 Agent 解决方案
- 🤔 不确定选择 Plan-Execute/Supervisor/Deep
- 🤔 评估现有 Agent 架构是否合理
- 🤔 需要将多个 Agent 组合成系统

## Additional Resources

- CloudWeGo Eino 文档：https://www.cloudwego.io/zh/docs/eino/
- Next-Show 项目：docs/next-show/ (Eino ADK 应用脚手架)
