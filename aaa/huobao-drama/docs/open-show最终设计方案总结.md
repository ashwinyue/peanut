# open-show 最终设计方案总结

> **版本**: v4.0 Final
> **创建日期**: 2026-02-22
> **项目目标**: 基于 Eino Graph 重构 huobao-drama，功能100%对等 + 新增HITL + 批量处理优化

---

## 📋 执行摘要

### 核心决策

| 决策项 | 选择方案 | 理由 |
|-------|---------|------|
| **数据库** | SQLite | 与 huobao 一致，零配置，纯 Go 驱动 |
| **ORM** | GORM + GORM Gen | 自动生成 Model 和 Query，提高开发效率 |
| **编排框架** | Eino Graph | 灵活的工作流编排，支持 HITL |
| **人工审核** | HITL 节点 | 关键节点可暂停审核，质量可控 |
| **批量处理** | alice 模式 | 串行+延迟，避免 API 限流 |
| **前端** | 复用 huobao Vue 3 | 节省开发时间，UI 一致 |

### 项目对比

| 特性 | huobao-drama | open-show |
|------|-------------|-----------|
| **代码行数** | ~1.9万行（后端） | ~2.5万行（后端） |
| **编排方式** | 固定流程 | Eino Graph（灵活） |
| **人工审核** | ❌ 无 | ✅ HITL 节点 |
| **批量处理** | 部分 | ✅ alice 模式 |
| **批量审核** | ❌ 无 | ✅ 支持 |
| **数据库** | SQLite | SQLite（一致） |
| **ORM** | GORM | GORM + GORM Gen |

---

## 🎯 功能完整性（100%对等）

### 核心功能模块（12个）

#### 1. 剧本管理（Script Management）
- ✅ AI 剧本生成（一句话生成剧本）
- ✅ 剧本结构化（提取角色、场景、分镜）
- ✅ 多类型支持（爱情/悬疑/动作等）
- ✅ 时长控制（自动规划分镜数量）
- ✅ 剧本编辑（手动编辑角色、场景、分镜、台词）
- ✅ 剧本导入（从小说生成剧本）

#### 2. 角色管理（Character Management）
- ✅ AI 角色生成（一句话生成角色）
- ✅ 批量生成角色（串行+延迟）
- ✅ 一致性保障（同一角色多图一致性）
- ✅ 角色编辑（手动编辑提示词）
- ✅ 角色变体（衣橱：日常/战斗/受伤）
- ✅ 角色转圈（360度角色展示）
- ✅ 角色库（保存到素材库，复用）
- ✅ 上传角色图（本地图片上传）

#### 3. 场景管理（Scene Management）
- ✅ AI 场景生成（一句话生成场景）
- ✅ 批量生成场景（串行+延迟）
- ✅ 纯环境场景（无人物干扰的纯背景）
- ✅ 场景编辑（手动编辑提示词）
- ✅ 场景库（保存到素材库，复用）
- ✅ 上传场景图（本地图片上传）

#### 4. 分镜管理（Storyboard Management）
- ✅ AI 分镜提取（自动提取分镜信息）
- ✅ 镜头类型选择（远景/中景/特写等）
- ✅ 运镜设计（推/拉/摇/移等）
- ✅ 九宫格分镜（一键生成9视角预览）
- ✅ 裁剪单格（裁剪为单独分镜）
- ✅ 整图作为首帧
- ✅ 手动分镜（添加/编辑/删除分镜）

#### 5. 帧管理（Frame Management）
- ✅ 首帧（Start Frame）
- ✅ 关键帧（Key Frame）
- ✅ 尾帧（End Frame）
- ✅ 分镜板（Panel，漫画风格）
- ✅ 自动生成帧提示词
- ✅ 手动编辑帧提示词
- ✅ 帧图片生成（文生图）
- ✅ 批量生成帧（串行+延迟）

#### 6. 图像生成（Image Generation）
- ✅ 多模型支持（OpenAI DALL-E / Midjourney / Stable Diffusion / Gemini）
- ✅ 横竖屏选择（16:9 / 9:16 / 1:1）
- ✅ 批量生成角色（串行+延迟）
- ✅ 批量生成场景（串行+延迟）
- ✅ 批量生成帧（串行+延迟）

#### 7. 视频生成（Video Generation）
- ✅ 图生视频（从首帧生成视频）
- ✅ 首尾帧插值（从首尾帧生成过渡视频）
- ✅ 多平台支持（OpenAI Sora / 豆包 / MiniMax / ChatFire 自研）
- ✅ 批量生成视频（串行+延迟）
- ✅ 进度追踪（实时显示生成进度）

#### 8. 视频合成（Video Merge）
- ✅ FFmpeg 合成（视频片段拼接）
- ✅ 转场效果（淡入淡出等）
- ✅ 音频合成（添加背景音乐）
- ✅ 成片导出（导出 MP4）
- ✅ 源素材导出（导出所有源素材 ZIP）

#### 9. 素材管理（Asset Management）
- ✅ 本地存储（可配置存储路径）
- ✅ 素材导入（从 ZIP 导入素材）
- ✅ 素材导出（导出为 ZIP）
- ✅ 项目数据导出（导出项目 JSON）
- ✅ 项目数据导入（导入项目 JSON）
- ✅ 素材库（角色/场景/道具素材库）
- ✅ 素材复用（从素材库导入到项目）

#### 10. 提示词管理（Prompt Management）
- ✅ 提示词国际化（中文/英文/日文）
- ✅ 地域特征前缀（中国人/日本人特征自动添加）
- ✅ 场景负面提示词（"无人物"指令）
- ✅ 提示词模板（角色/场景/分镜预设风格）

#### 11. 任务管理（Task Management）
- ✅ 实时进度追踪（显示当前任务进度）
- ✅ 任务状态（pending/processing/waiting/completed/failed）
- ✅ 错误处理（显示错误信息）
- ✅ 任务历史（查看历史任务）
- ✅ 任务重试（重新执行失败任务）

#### 12. 上传服务（Upload Service）
- ✅ 图片上传（上传角色/场景图片）
- ✅ 视频上传（上传视频片段）
- ✅ 批量上传（批量上传多张图片）

### 新增功能模块（2个）

#### 13. HITL 人工审核（NEW）
- ✅ 剧本审核（剧本生成后自动暂停，审核后继续）
- ✅ 角色审核（角色生成后自动暂停，审核后继续）
- ✅ 分镜审核（分镜提取后自动暂停，审核后继续）
- ✅ 帧审核（帧生成后自动暂停，审核后继续）
- ✅ 批量审核（批量通过/拒绝/修改）
- ✅ 审核记录（保存所有审核操作）

#### 14. 批量处理（Batch Processing - NEW）
- ✅ 批量创建任务（一次性创建多个短剧任务）
- ✅ 批量生成角色（串行+延迟，alice 模式）
- ✅ 批量生成场景（串行+延迟，alice 模式）
- ✅ 批量生成帧（串行+延迟，alice 模式）
- ✅ 批量生成视频（串行+延迟，alice 模式）
- ✅ 批量审核（批量通过/拒绝/修改）
- ✅ 批量导出（批量导出视频/素材）

---

## 🏗️ 技术架构

### 1. 整体架构图

```
┌─────────────────────────────────────────────────────────┐
│                    前端（Vue 3 + alice UI）               │
│  - 复用 huobao-drama 完整前端                           │
│  - 新增 HITL 审核工作台                                    │
│  - 新增批量操作面板                                        │
└─────────────────────────────────────────────────────────┘
                        ▲ HTTP + SSE
                        │
┌─────────────────────────────────────────────────────────┐
│                   API Layer (Gin)                         │
│  POST /api/v1/projects                    创建项目       │
│  GET  /api/v1/projects/:id                获取项目       │
│  POST /api/v1/projects/:id/generate      生成短剧       │
│  POST /api/v1/projects/:id/cancel       取消生成       │
│  POST /api/v1/characters/batch          批量生成角色     │
│  POST /api/v1/scenes/batch             批量生成场景     │
│  POST /api/v1/frames/batch              批量生成帧       │
│  POST /api/v1/videos/batch              批量生成视频     │
│  GET  /api/v1/review/pending            获取待审核列表   │
│  POST /api/v1/review/:id/approve        审核通过         │
│  POST /api/v1/review/batch/approve     批量审核通过     │
│  GET  /api/v1/projects/:id/progress     生成进度（SSE）  │
│  POST /api/v1/assets/library           保存到素材库     │
│  GET  /api/v1/assets/library           获取素材库       │
│  POST /api/v1/assets/import             导入素材         │
│  POST /api/v1/assets/export             导出素材         │
└─────────────────────────────────────────────────────────┘
                        │
┌─────────────────────────────────────────────────────────┐
│              Graph Orchestration (Eino)                  │
│                                                         │
│  ┌────────────────────────────────────────────────┐     │
│  │  START                                         │     │
│  │    │ (可选：HITL 审核)                           │     │
│  │    ▼                                           │     │
│  │  [剧本生成 Node]                                │     │
│  │    │ (可选：HITL 审核)                           │     │
│  │    ├─→ [HITL 节点] ⏸️  ←──┐                   │     │
│  │    │                    │                   │     │
│  │    ▼                    ▼                   │     │
│  │  [角色生成 Node ←───────┘ (approved)       │     │
│  │    │ (并行) (可选：HITL 审核)                  │     │
│  │    ├─→ [场景生成 Node]                         │     │
│  │    │   (可选：HITL 审核)                        │     │
│  │    ├─→ [道具生成 Node]                          │     │
│  │    │   (可选：HITL 审核)                        │     │
│  │    └─→ [HITL 节点] ⏸️  ←──┐                   │     │
│  │    │                    │                   │     │
│  │    ▼                    ▼                   │     │
│  │  [分镜提取 Node]                                │     │
│  │    │ (可选：HITL 审核)                          │     │
│  │    ├─→ [HITL 节点] ⏸️  ←──┐                   │     │
│  │    │                    │                   │     │
│  │    ▼                    ▼                   │     │
│  │  [帧生成 Node] (批量生成 - alice 模式)        │     │
│  │    │ (可选：HITL 审核)                          │     │
│  │    ├─→ [HITL 节点] ⏸️  ←──┐                   │     │
│  │    │                    │                   │     │
│  │    ▼                    ▼                   │     │
│  │  [视频生成 Node] (批量生成 - alice 模式)        │     │
│  │    │                                           │     │
│  │    ▼                                           │     │
│  │  [视频合成 Node]                                │     │
│  │    │                                           │     │
│  │    ▼                                           │     │
│ │   END                                          │     │
│  └────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────┘
                        │
┌─────────────────────────────────────────────────────────┐
│              Service Layer (12个核心服务)                │
│  ScriptService | CharacterService | SceneService       │
│  StoryboardService | FrameService | ImageGenService    │
│  VideoGenService | MergeService | PromptService        │
│  BatchService | ReviewService | TaskService           │
│  AssetService | UploadService                         │
└─────────────────────────────────────────────────────────┘
                        │
┌─────────────────────────────────────────────────────────┐
│              Repository Layer (GORM Gen 生成)            │
│  ProjectRepo | ScriptRepo | CharacterRepo | ...        │
└─────────────────────────────────────────────────────────┘
                        │
┌─────────────────────────────────────────────────────────┐
│              Tool Layer                                │
│  LLM Tool | Image Tool | Video Tool | FFmpeg Tool       │
└─────────────────────────────────────────────────────────┘
                        │
┌─────────────────────────────────────────────────────────┐
│              Database (SQLite)                        │
│  projects | scripts | characters | scenes | storyboards   │
│  frames | asset_library | hitl_records | tasks          │
└─────────────────────────────────────────────────────────┘
```

### 2. 服务层设计（12个核心服务）

```go
// 1. ScriptService 剧本服务
type ScriptService interface {
    GenerateScript(ctx context.Context, input *ScriptInput) (*Script, error)
    ParseScript(ctx context.Context, content string) (*Script, error)
    UpdateScript(ctx context.Context, scriptID string, updates *ScriptUpdate) (*Script, error)
    ExtractStoryboards(ctx context.Context, scriptID string) ([]*Storyboard, error)
}

// 2. CharacterService 角色服务
type CharacterService interface {
    GenerateCharacter(ctx context.Context, input *CharacterInput) (*Character, error)
    BatchGenerateCharacters(ctx context.Context, characterIDs []string, progressCallback func(*BatchProgress)) error
    AddVariation(ctx context.Context, characterID string, variation *Variation) error
    UploadCharacter(ctx context.Context, characterID string, image []byte) error
    SaveToLibrary(ctx context.Context, characterID string) (*AssetLibraryItem, error)
    LoadFromLibrary(ctx context.Context, libraryItemID string) (*Character, error)
}

// 3. SceneService 场景服务
type SceneService interface {
    GenerateScene(ctx context.Context, input *SceneInput) (*Scene, error)
    BatchGenerateScenes(ctx context.Context, sceneIDs []string, progressCallback func(*BatchProgress)) error
    UploadScene(ctx context.Context, sceneID string, image []byte) error
    SaveToLibrary(ctx context.Context, sceneID string) (*AssetLibraryItem, error)
    LoadFromLibrary(ctx context.Context, libraryItemID string) (*Scene, error)
}

// 4. StoryboardService 分镜服务
type StoryboardService interface {
    ExtractStoryboards(ctx context.Context, scriptID string) ([]*Storyboard, error)
    GenerateStoryboardScript(ctx context.Context, storyboardID string) error
    UpdateStoryboard(ctx context.Context, storyboardID string, updates *StoryboardUpdate) error
    DeleteStoryboard(ctx context.Context, storyboardID string) error
    ComposeNineGrid(ctx context.Context, storyboardID string) (*NineGrid, error)
    ComposeManual(ctx context.Context, storyboardID string, layout string) error
}

// 5. FrameService 帧服务
type FrameService interface {
    GenerateFramePrompt(ctx context.Context, frameID string) (string, error)
    GenerateFrame(ctx context.Context, frameID string) (*Frame, error)
    BatchGenerateFrames(ctx context.Context, frameIDs []string, progressCallback func(*BatchProgress)) error
    SetFrameType(ctx context.Context, frameID string, frameType FrameType) error
}

// 6. ImageGenerationService 图像生成服务
type ImageGenerationService interface {
    GenerateImage(ctx context.Context, prompt string, options *ImageOptions) (string, error)
    GenerateCharacterImage(ctx context.Context, character *Character, style string) (string, error)
    GenerateSceneImage(ctx context.Context, scene *Scene) (string, error)
    GenerateFrameImage(ctx context.Context, frame *Frame) (string, error)
}

// 7. VideoGenerationService 视频生成服务
type VideoGenerationService interface {
    GenerateVideo(ctx context.Context, imageURL string, options *VideoOptions) (string, error)
    GenerateVideoFromFrames(ctx context.Context, startFrame, endFrame string, options *VideoOptions) (string, error)
    BatchGenerateVideos(ctx context.Context, storyboardIDs []string, progressCallback func(*BatchProgress)) error
}

// 8. VideoMergeService 视频合成服务
type VideoMergeService interface {
    MergeVideos(ctx context.Context, storyboardIDs []string, options *MergeOptions) (string, error)
    AddTransition(ctx context.Context, videoID string, transitionType string) error
    ExportMasterVideo(ctx context.Context, projectID string) (string, error)
    ExportSourceAssets(ctx context.Context, projectID string) (string, error)
}

// 9. PromptService 提示词服务
type PromptService interface {
    GenerateCharacterPrompt(ctx context.Context, character *Character, language, visualStyle, genre string) (*VisualPrompts, error)
    GenerateScenePrompt(ctx context.Context, scene *Scene, language, visualStyle, genre string) (*VisualPrompts, error)
    GenerateFramePrompt(ctx context.Context, frame *Frame, language, visualStyle, genre string) (*VisualPrompts, error)
    GetRegionalPrefix(language string, assetType string) string
}

// 10. BatchService 批量处理服务
type BatchService interface {
    BatchGenerate(ctx context.Context, itemType string, itemIDs []string, generator func(id string) error, progressCallback func(*BatchProgress)) error
    GetBatchProgress(ctx context.Context, batchID string) (*BatchProgress, error)
}

// 11. ReviewService 审核服务
type ReviewService interface {
    ReviewTask(ctx context.Context, taskID, stage string, action string, feedback string, modifyData any) error
    BatchReview(ctx context.Context, taskIDs []string, action string, feedback string, modifyData map[string]any, progressCallback func(*BatchProgress)) error
    GetPendingReviews(ctx context.Context, stage string) ([]*HITLRecord, error)
}

// 12. AssetService 素材服务
type AssetService interface {
    SaveAsset(ctx context.Context, assetType string, data []byte) (*AssetLibraryItem, error)
    GetAssets(ctx context.Context, assetType string) ([]*AssetLibraryItem, error)
    ApplyAsset(ctx context.Context, projectID, assetID string) error
    ImportAssets(ctx context.Context, projectID string, zipFile []byte) error
    ExportAssets(ctx context.Context, projectID string) (string, error)
    ExportProjectData(ctx context.Context, projectID string) ([]byte, error)
    ImportProjectData(ctx context.Context, data []byte) (string, error)
}

// 13. TaskService 任务服务
type TaskService interface {
    CreateTask(ctx context.Context, input *DramaInput) (*Task, error)
    GetTask(ctx context.Context, taskID string) (*Task, error)
    UpdateTaskStatus(ctx context.Context, taskID string, status TaskStatus) error
    UpdateTaskProgress(ctx context.Context, taskID string, progress float64, stage string) error
    CancelTask(ctx context.Context, taskID string) error
    GetTaskProgress(ctx context.Context, taskID string) (<-chan *Progress, error)
}

// 14. UploadService 上传服务
type UploadService interface {
    UploadImage(ctx context.Context, file []byte, filename string) (string, error)
    UploadVideo(ctx context.Context, file []byte, filename string) (string, error)
    BatchUploadImages(ctx context.Context, files []byte[], progressCallback func(*UploadProgress)) error
}
```

---

## 📊 数据库设计

### 数据库选型

```
数据库: SQLite (与 huobao 一致)
驱动: modernc.org/sqlite (纯 Go 实现，无 CGO 依赖)
ORM: GORM v1.30.0
代码生成: GORM Gen
```

### GORM Gen 配置

```yaml
# gen.conf
database:
  dsn: "sqlite:data/drama.db"

# 生成配置
output:
  model: ./internal/models
  query: ./internal/repository/query
  filer: ./internal/repository

# 字段标签
field_taggable: true

# 模型配置
model:
  table_prefix: true
  naming_style: snake_case

# Query 配置
query:
  naming_style: snake_case
```

### 数据表结构

```sql
-- 项目表
CREATE TABLE projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    language VARCHAR(50) DEFAULT '中文',
    visual_style VARCHAR(50) DEFAULT 'live-action',
    genre VARCHAR(50),
    status VARCHAR(20) DEFAULT 'draft',
    progress REAL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 剧本表
CREATE TABLE scripts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    synopsis TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- 集数表
CREATE TABLE episodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    script_id INTEGER NOT NULL,
    episode_num INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (script_id) REFERENCES scripts(id) ON DELETE CASCADE
);

-- 场景表
CREATE TABLE scenes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    script_id INTEGER,
    episode_id INTEGER,
    scene_num INTEGER NOT NULL,
    location VARCHAR(255),
    time VARCHAR(50),
    mood VARCHAR(50),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (script_id) REFERENCES scripts(id) ON DELETE SET NULL,
    FOREIGN KEY (episode_id) REFERENCES episodes(id) ON DELETE SET NULL
);

-- 角色表
CREATE TABLE characters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    appearance TEXT,
    personality TEXT,
    visual_prompt TEXT,
    negative_prompt TEXT,
    reference_image VARCHAR(1024),
    status VARCHAR(20) DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- 角色变体表
CREATE TABLE character_variations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    reference_image VARCHAR(1024),
    status VARCHAR(20) DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);

-- 分镜表
CREATE TABLE storyboards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    scene_id INTEGER,
    shot_num INTEGER NOT NULL,
    shot_type VARCHAR(50),
    camera_move VARCHAR(50),
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (scene_id) REFERENCES scenes(id) ON DELETE SET NULL
);

-- 帧表
CREATE TABLE frames (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    storyboard_id INTEGER NOT NULL,
    frame_type VARCHAR(50) NOT NULL,
    prompt TEXT,
    image_url VARCHAR(1024),
    status VARCHAR(20) DEFAULT 'pending',
    duration REAL,
    video_url VARCHAR(1024),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (storyboard_id) REFERENCES storyboards(id) ON DELETE CASCADE
);

-- 素材库表
CREATE TABLE asset_library (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    data JSON NOT NULL,
    image_url VARCHAR(1024),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- HITL 审核记录表
CREATE TABLE hitl_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id VARCHAR(255) NOT NULL,
    stage VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    data JSON,
    feedback TEXT,
    reviewed_by VARCHAR(255),
    reviewed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 任务表
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL,
    current_stage VARCHAR(50),
    progress REAL DEFAULT 0,
    error TEXT,
    result JSON,
    completed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- 索引
CREATE INDEX idx_hitl_pending ON hitl_records(status, stage) WHERE status = 'pending';
CREATE INDEX idx_tasks_user ON tasks(project_id, status);
CREATE INDEX idx_tasks_status ON tasks(status, created_at);
```

---

## 🔄 Graph 编排流程

### 完整工作流（含 HITL）

```
START
  │
  ├─→ [剧本生成 Node]
  │   │ (HITL 可选)
  │   └─→ [HITL 节点 ⏸️] ─→ approved → 继续
  │                              rejected → 结束
  │
  ├─→ [角色生成 Node] (批量生成 - alice 模式)
  │   │ (HITL 可选)
  │   └─→ [HITL 节点 ⏸️] ─→ approved → 继续
  │                              rejected → 结束
  │
  ├─→ [场景生成 Node] (批量生成 - alice 模式)
  │   │ (HITL 可选)
  │   └─→ [HITL 节点 ⏸️] ─→ approved → 继续
  │                              rejected → 结束
  │
  ├─→ [道具生成 Node] (批量生成 - alice 模式)
  │   │ (HITL 可选)
  │   └─→ [HITL 节点 ⏸️] ─→ approved → 并行汇聚
  │                              rejected → 结束
  │
  ├─→ [分镜提取 Node]
  │   │ (HITL 可选)
  │   └─→ [HITL 节点 ⏸️] ─→ approved → 继续
  │                              rejected → 结束
  │
  ├─→ [帧生成 Node] (批量生成 - alice 模式)
  │   │ (HITL 可选)
  │   └─→ [HITL 节点 ⏸️] ─→ approved → 继续
  │                              rejected → 结束
  │
  ├─→ [视频生成 Node] (批量生成 - alice 模式)
  │   │ (HITL 可选)
  │   └─→ [HITL 节点 ⏸️] ─→ approved → 继续
  │                              rejected → 结束
  │
  └─→ [视频合成 Node]
      │
      └→ END
```

### HITL 状态机

```
pending ─┬─→ approved ─→ 继续执行
         ├─→ modified ─→ 回退重新生成
         ├─→ rejected ─→ 终止流程
         └─→ timeout ─→ rejected (24小时超时)
```

---

## 🚀 批量处理（alice 模式）

### 批量生成实现

```go
type BatchProgress struct {
    Current int         `json:"current"`
    Total   int         `json:"total"`
    Message string      `json:"message"`
    Errors  []string    `json:"errors,omitempty"`
}

func (s *BatchService) BatchGenerate(
    ctx context.Context,
    itemIDs []string,
    generator func(id string) error,
    progressCallback func(*BatchProgress),
) error {
    progress := &BatchProgress{
        Current: 0,
        Total:   len(itemIDs),
        Errors:  make([]string, 0),
    }

    for i, id := range itemIDs {
        // 更新进度
        progress.Current = i + 1
        progress.Message = fmt.Sprintf("正在生成 %d/%d", i+1, len(itemIDs))
        progressCallback(progress)

        // 执行生成
        if err := generator(id); err != nil {
            // 记录错误，继续执行（容错）
            progress.Errors = append(progress.Errors, fmt.Sprintf("%s: %s", id, err.Error()))
        }

        // 延迟（alice 默认 3秒，避免 API 限流）
        if i < len(itemIDs)-1 {
            time.Sleep(3 * time.Second)
        }
    }

    progress.Message = "完成"
    progressCallback(progress)
    return nil
}
```

### 批量审核实现

```go
func (s *ReviewService) BatchReview(
    ctx context.Context,
    taskIDs []string,
    action string,
    feedback string,
    progressCallback func(*BatchProgress),
) error {
    return s.batchSvc.BatchGenerate(ctx, taskIDs, func(taskID string) error {
        return s.reviewTask(ctx, taskID, action, feedback)
    }, progressCallback)
}
```

---

## 📁 项目结构

```
open-show/
├── cmd/
│   └── server/
│       └── main.go                      # 入口
├── internal/
│   ├── graph/                             # Graph 编排
│   │   ├── drama_graph.go                # 主图编排
│   │   ├── hitl_node.go                  # HITL 节点
│   │   ├── pipeline_state.go             # 状态管理
│   │   └── nodes/                         # 节点实现
│   │       ├── script_node.go
│   │       ├── character_node.go
│   │       ├── scene_node.go
│   │       ├── storyboard_node.go
│   │       ├── frame_node.go
│   │       ├── video_node.go
│   │       └── merge_node.go
│   ├── services/                          # 业务服务（12个）
│   │   ├── script_service.go
│   │   ├── character_service.go
│   │   ├── scene_service.go
│   │   ├── storyboard_service.go
│   │   ├── frame_service.go
│   │   ├── image_generation_service.go
│   │   ├── video_generation_service.go
│   │   ├── merge_service.go
│   │   ├── prompt_service.go              # 提示词 i18n
│   │   ├── batch_service.go               # 批量处理
│   │   ├── review_service.go              # 审核服务
│   │   ├── task_service.go                # 任务管理
│   │   ├── asset_service.go               # 素材管理
│   │   └── upload_service.go              # 上传服务
│   ├── handlers/                          # HTTP 处理
│   │   ├── project.go
│   │   ├── script.go
│   │   ├── character.go
│   │   ├── scene.go
│   │   ├── storyboard.go
│   │   ├── frame.go
│   │   ├── video.go
│   │   ├── asset.go
│   │   ├── batch.go                       # 批量操作
│   │   ├── review.go                      # 审核操作
│   │   └── upload.go
│   ├── models/                            # 数据模型（GORM Gen）
│   │   ├── base.go                         # 基础模型
│   │   ├── project.generated.go           # 自动生成
│   │   ├── script.generated.go
│   │   ├── character.generated.go
│   │   ├── scene.generated.go
│   │   ├── storyboard.generated.go
│   │   ├── frame.generated.go
│   │   ├── asset_library.generated.go
│   │   ├── hitl_record.generated.go
│   │   └── task.generated.go
│   ├── repository/                        # 数据访问
│   │   ├── query/                         # GORM Gen 生成
│   │   │   └── *.gen.go
│   │   └── project_repo.go               # 可选：手动实现
│   ├── tools/                             # 工具集成
│   │   ├── llm/
│   │   │   ├── openai_client.go
│   │   │   ├── gemini_client.go
│   │   │   └── doubao_client.go
│   │   ├── image/
│   │   │   ├── dalle_client.go
│   │   │   ├── midjourney_client.go
│   │   │   └── sd_client.go
│   │   ├── video/
│   │   │   ├── sora_client.go
│   │   │   ├── doubao_client.go
│   │   │   └── minimax_client.go
│   │   └── ffmpeg/
│   │       └── composer.go
│   ├── config/                            # 配置
│   │   └── config.go
│   └── sse/                               # SSE
│       └── writer.go
├── web/                                  # 前端（复用 huobao）
│   └── src/
│       ├── views/
│       │   ├── ProjectDashboard.vue      # 项目管理
│       │   ├── ScriptEditor.vue          # 剧本编辑
│       │   ├── CharacterManager.vue      # 角色管理
│       │   ├── SceneManager.vue         # 场景管理
│       │   ├── StoryboardComposer.vue   # 分镜制作
│       │   ├── FrameManager.vue         # 帧管理
│       │   ├── StageExport.vue          # 成片导出
│       │   ├── ReviewWorkbench.vue      # HITL 审核工作台（NEW）
│       │   └── BatchActions.vue         # 批量操作面板（NEW）
│       ├── components/
│       │   ├── BatchProgress.vue        # 批量进度显示
│       │   └── ...
│       └── services/
│           ├── api.ts
│           └── sse.ts
├── data/
│   └── drama.db                         # SQLite 数据库
├── configs/
│   └── config.yaml
├── gen.conf                              # GORM Gen 配置
├── go.mod
├── go.sum
└── README.md
```

---

## 🎯 实施计划

### Phase 1: 项目初始化（1周）

| 任务 | 说明 |
|------|------|
| 创建项目结构 | 按照 `open-show/` 结构创建目录 |
| 配置 GORM Gen | 安装 GORM Gen，配置 `gen.conf` |
| 定义基础模型 | 定义 `BaseModel` 和核心表结构 |
| 数据库初始化 | SQLite 连接测试，自动迁移 |
| 基础 API | 实现 Project CRUD API |

### Phase 2: 核心服务（6-8周）

| 服务 | 时间 | 说明 |
|------|------|------|
| ScriptService | 1周 | 剧本生成、解析、编辑 |
| CharacterService | 1周 | 角色生成、变体、素材库 |
| SceneService | 1周 | 场景生成、素材库 |
| StoryboardService | 1.5周 | 分镜提取、九宫格、拼接 |
| FrameService | 1周 | 帧提示词、生成 |
| ImageGenerationService | 1周 | 图像生成（多模型） |
| VideoGenerationService | 1周 | 视频生成（多平台） |
| MergeService | 1周 | 视频合成、转场、导出 |
| PromptService | 1.5周 | 提示词 i18n、地域特征 |

### Phase 3: Graph 编排（3-4周）

| 任务 | 时间 | 说明 |
|------|------|------|
| Graph 定义 | 1周 | 主图编排、节点定义 |
| 节点实现 | 2周 | 12个节点的具体实现 |
| HITL 节点 | 1周 | 人工审核节点实现 |
| 状态管理 | 0.5周 | PipelineState 和进度追踪 |

### Phase 4: 新增功能（3-4周）

| 任务 | 时间 | 说明 |
|------|------|------|
| BatchService | 1周 | 批量生成（alice 模式） |
| ReviewService | 1周 | 审核服务 |
| TaskService | 1周 | 任务管理 |
| 批量 API | 1周 | 批量操作 API |
| HITL 前端 | 2周 | 审核工作台、批量操作面板 |

### Phase 5: 前端适配（2-3周）

| 任务 | 时间 | 说明 |
|------|------|------|
| 复用 huobao 前端 | - | 完整的 Vue 3 UI |
| 新增审核界面 | 1周 | ReviewWorkbench.vue |
| 新增批量操作 | 1周 | BatchActions.vue |
| SSE 集成 | 1周 | 实时进度推送 |

### Phase 6: 测试与优化（2周）

| 任务 | 时间 | 说明 |
|------|------|------|
| 单元测试 | 1周 | 服务层测试 |
| 集成测试 | 0.5周 | API 测试 |
| 性能优化 | 0.5周 | 批量处理优化 |

---

## 📚 配置文件

### gen.conf（GORM Gen）

```yaml
database:
  dsn: "sqlite:data/drama.db"

output:
  model: ./internal/models
  query: ./internal/repository/query
  filer: ./internal/repository

field_taggable: true

model:
  table_prefix: true
  naming_style: snakecase

query:
  naming_style: snake_code
```

### config.yaml（应用配置）

```yaml
app:
  name: "open-show"
  version: "1.0.0"
  debug: true

server:
  port: 5678
  host: "0.0.0.0"

database:
  type: "sqlite"
  path: "./data/drama.db"
  max_idle: 10
  max_open: 100

storage:
  type: "local"
  local_path: "./data/storage"
  base_url: "http://localhost:5678/static"

# 批量处理配置（参考 alice）
batch:
  delay: 3000              # 批量延迟（毫秒）
  concurrency: 1           # 串行执行
  retry_on_error: false    # 错误不重试

# HITL 配置
hitl:
  timeout: 24              # 审核超时（小时）
  poll_interval: 2         # 轮询间隔（秒）
  auto_approve: false      # 不自动批准
  enable_stages:         # 启用审核的阶段
    - script
    - character
    - storyboard
    - frame

# AI 模型配置
ai:
  default_text_provider: "openai"
  default_image_provider: "openai"
  default_video_provider: "doubao"

# 多模型配置
models:
  openai:
    base_url: "https://api.openai.com/v1"
    api_key: "${OPENAI_API_KEY}"
    models:
      text: "gpt-4"
      image: "dall-e-3"
      video: "sora"

  gemini:
    base_url: "${GEMINI_API_URL}"
    api_key: "${GEMINI_API_KEY}"
    models:
      text: "gemini-2.5-pro"
      image: "gemini-2.0-flash-exp"

  doubao:
    base_url: "https://ark.cn/blue"
    api_key: "${DOUBAO_API_KEY}"
    models:
      video: "doubao-pro"
```

---

## 📊 代码量估算

| 模块 | 代码行数 | 说明 |
|------|---------|------|
| **Graph 编排** | ~500 | Eino Graph + 节点 |
| **HITL 节点** | ~150 | 人工审核逻辑 |
| **服务层** | ~8,000 | 12个核心服务（对等 huobao） |
| **API 层** | ~1,500 | REST API + 批量 API |
| **模型层** | ~2,000 | GORM Gen 自动生成 |
| **工具层** | ~3,000 | LLM/图像/视频/FFmpeg 客户端 |
| **前端** | ~15,000 | 复用 huobao + 新增界面 |
| **配置/其他** | ~1,000 | 配置、SSE 等 |
| **总计** | **~31,150 行** | **功能完整对等** |

---

## ✅ 验收标准

### 功能完整性

- [ ] 剧本管理：100% 对等 huobao
- [ ] 角色管理：100% 对等 huobao（含变体、转身、素材库）
- [ ] 场景管理：100% 对等 huobao（含素材库）
- [ ] 分镜管理：100% 对等 huobao（含九宫格、拼接）
- [ ] 帧管理：100% 对等 huobao（含首帧/关键帧/尾帧）
- [ ] 图像生成：100% 对等 huobao（多模型、横竖屏）
- [ ] 视频生成：100% 对等 huobao（多平台、首尾帧插值）
- [ ] 视频合成：100% 对等 huobao（FFmpeg、转场、导出）
- [ ] 素材管理：100% 对等 huobao（导入导出、素材库）
- [ ] 提示词 i18n：100% 对等 huobao（中/英/日）
- [ ] 上传服务：100% 对等 huobao
- [ ] **HITL 审核**：新增功能
- [ ] **批量处理**：优化（alice 模式）
- [ ] **批量审核**：新增功能

### 技术指标

- [ ] SQLite 数据库（与 huobao 一致）
- [ ] GORM + GORM Gen（自动生成）
- [ ] Eino Graph 编排（灵活工作流）
- [ ] SSE 实时进度推送
- [ ] 批量生成（串行+延迟，alice 模式）
- [ ] HITL 人工审核（可配置开关）

---

## 📝 文档清单

1. ✅ open-show完整功能设计-v3.md（本文档）
2. ✅ Eino改造方案-next-show基础.md（参考）
3. ✅ open-show架构设计-基于alice.md（参考）
4. 📄 gen.conf（GORM Gen 配置）
5. 📄 config.yaml（应用配置）
6. 📄 README.md（项目说明）

---

## 🚀 下一步行动

### 立即开始

1. **创建项目脚手架**
   ```bash
   mkdir -p open-show
   cd open-show
   go mod init github.com/open-show
   ```

2. **安装依赖**
   ```bash
   go get github.com/cloudwego/eino@latest
   go get github.com/cloudwego/eino-ext@latest
   go get gorm.io/gorm@latest
   go get gorm.io/driver/sqlite@latest
   go get modernc.org/sqlite@latest
   ```

3. **安装 GORM Gen**
   ```bash
   go install -u gorm.io/gen/tools/genum@latest
   go install -u gorm.io/gen/tools/gen@latest
   ```

4. **创建 gen.conf**
   ```bash
   # 按照上面的 gen.conf 配置创建
   ```

5. **定义基础模型**
   ```bash
   # 创建 internal/models/base.go
   ```

6. **生成 Model 和 Query**
   ```bash
   gen
   ```

---

**是否开始创建项目脚手架？**

**需要我详细展开某个模块的实现吗？**

**有其他问题或需要调整的地方吗？**
