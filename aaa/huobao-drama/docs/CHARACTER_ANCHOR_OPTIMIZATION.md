# 角色锚点系统优化方案

> **文档版本**: v1.0
> **创建日期**: 2025-02-22
> **作者**: Claude Code
> **状态**: 设计阶段

---

## 📋 目录

- [一、现状分析](#一现状分析)
- [二、优化方案](#二优化方案)
- [三、实施计划](#三实施计划)
- [四、风险评估](#四风险评估)
- [五、API设计](#五api设计)
- [六、UI设计](#六ui设计)

---

## 一、现状分析

### 1.1 当前问题

| 问题 | 严重程度 | 说明 |
|------|---------|------|
| **单图片限制** | 🔴 高 | 只有 `ImageURL` 单字段，无法存储三视图 |
| **ReferenceImages 未使用** | 🟡 中 | JSON 字段存在但未在业务逻辑中使用 |
| **SeedValue 未利用** | 🟡 中 | 种子值未在生成时使用，无法保证可重现性 |
| **缺少衣橱系统** | 🔴 高 | 没有角色服装/造型变体的数据结构 |
| **锚点关联简陋** | 🔴 高 | 仅标识角色存在，无法描述位置/姿态/视角 |

### 1.2 当前数据模型

```go
// Character 模型 - 功能过于简单
type Character struct {
    ID              uint
    Name            string
    ImageURL        *string        // ❌ 只有一张图片
    ReferenceImages datatypes.JSON // ⚠️ 未使用
    SeedValue       *string        // ⚠️ 未利用
}

// storyboard_characters 关联表 - 信息不足
// ❌ 无法表达：视角、服装、位置、姿态
```

### 1.3 现有功能

- ✅ 角色基础信息管理
- ✅ 角色库系统
- ✅ 分镜-角色多对多关联
- ⚠️ 角色图片作为参考图传递（但只能使用单一图片）

---

## 二、优化方案

### 2.1 数据模型增强

#### 2.1.1 CharacterView 模型（三视图系统）

```go
// CharacterView 角色视图（三视图）
type CharacterView struct {
    ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
    CharacterID uint           `gorm:"not null;index" json:"character_id"`
    ViewType    string         `gorm:"type:varchar(20);not null" json:"view_type"` // front, side, back, three_quarter
    ImageURL    string         `gorm:"type:varchar(500);not null" json:"image_url"`
    LocalPath   *string        `gorm:"type:varchar(500)" json:"local_path,omitempty"`
    Description *string        `gorm:"type:text" json:"description"`
    IsPrimary   bool           `gorm:"default:false" json:"is_primary"` // 是否为主视图
    CreatedAt   time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

    Character Character `gorm:"foreignKey:CharacterID" json:"character,omitempty"`
}

func (c *CharacterView) TableName() string {
    return "character_views"
}
```

**ViewType 枚举值：**
- `front` - 正面视图
- `side` - 侧面视图
- `back` - 背面视图
- `three_quarter` - 三分之四视图
- `top_down` - 俯视图
- `bottom_up` - 仰视图

#### 2.1.2 CharacterOutfit 模型（衣橱系统）

```go
// CharacterOutfit 角色服装/造型变体
type CharacterOutfit struct {
    ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
    CharacterID uint           `gorm:"not null;index" json:"character_id"`
    Name        string         `gorm:"type:varchar(100);not null" json:"name"` // 服装名称，如"校服"、"战斗服"
    Description *string        `gorm:"type:text" json:"description"`
    ImageURL    string         `gorm:"type:varchar(500);not null" json:"image_url"`
    LocalPath   *string        `gorm:"type:varchar(500)" json:"local_path,omitempty"`
    Tags        *string        `gorm:"type:varchar(500)" json:"tags"` // casual, formal, battle, etc.
    IsDefault   bool           `gorm:"default:false" json:"is_default"` // 是否为默认造型
    CreatedAt   time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

    Character Character `gorm:"foreignKey:CharacterID" json:"character,omitempty"`
}

func (c *CharacterOutfit) TableName() string {
    return "character_outfits"
}
```

**Tags 示例：**
- `casual` - 休闲装
- `formal` - 正式装
- `battle` - 战斗装
- `school` - 校服
- `traditional` - 传统服饰

#### 2.1.3 StoryboardCharacterAnchor 模型（锚点关联）

```go
// StoryboardCharacterAnchor 分镜-角色锚点关联表
type StoryboardCharacterAnchor struct {
    ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
    StoryboardID uint           `gorm:"not null;index:idx_sca_storyboard" json:"storyboard_id"`
    CharacterID  uint           `gorm:"not null;index:idx_sca_character" json:"character_id"`

    // 锚点配置
    ViewType       *string `gorm:"type:varchar(20)" json:"view_type"`        // 使用的视角：front, side, back
    OutfitID       *uint   `json:"outfit_id"`                                // 使用的服装ID
    Position       *string `gorm:"type:varchar(50)" json:"position"`         // 位置：left, center, right, foreground, background
    Depth          *string `gorm:"type:varchar(20)" json:"depth"`            // 景深：foreground, midground, background
    Scale          *float64 `json:"scale"`                                   // 缩放比例
    Pose           *string `gorm:"type:text" json:"pose"`                    // 姿态描述
    Action         *string `gorm:"type:text" json:"action"`                  // 动作描述
    FacialExpression *string `gorm:"type:varchar(100)" json:"facial_expression"` // 面部表情

    // 高级选项
    IsPrimary      bool    `gorm:"default:false" json:"is_primary"`          // 是否为主要角色
    Visibility     float64 `gorm:"default:1.0" json:"visibility"`            // 可见度 0-1
    IsVisible      bool    `gorm:"default:true" json:"is_visible"`           // 是否可见（用于隐藏角色）

    CreatedAt      time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
    UpdatedAt      time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
    DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

    Storyboard Storyboard `gorm:"foreignKey:StoryboardID" json:"storyboard,omitempty"`
    Character  Character  `gorm:"foreignKey:CharacterID" json:"character,omitempty"`
    Outfit     *CharacterOutfit `gorm:"foreignKey:OutfitID" json:"outfit,omitempty"`
}

func (s *StoryboardCharacterAnchor) TableName() string {
    return "storyboard_character_anchors"
}
```

**Position 枚举值：**
- `left` - 画面左侧
- `center` - 画面中央
- `right` - 画面右侧
- `foreground` - 前景
- `background` - 背景

**Depth 枚举值：**
- `foreground` - 前景
- `midground` - 中景
- `background` - 后景

#### 2.1.4 修改现有 Character 模型

```go
type Character struct {
    // ... 保持现有字段

    // 新增字段
    PrimaryViewID   *uint  `gorm:"index" json:"primary_view_id"`      // 主视图ID
    DefaultOutfitID *uint  `gorm:"index" json:"default_outfit_id"`   // 默认服装ID
    VoiceSeed       *string `gorm:"type:varchar(100)" json:"voice_seed"` // 语音种子（用于TTS）

    // 运行时字段（不存储）
    Views   []CharacterView    `gorm:"foreignKey:CharacterID" json:"views,omitempty"`
    Outfits []CharacterOutfit  `gorm:"foreignKey:CharacterID" json:"outfits,omitempty"`
}
```

### 2.2 服务层设计

#### 2.2.1 CharacterAnchorService

```go
package services

import (
    "gorm.io/gorm"
    "github.com/drama-generator/backend/domain/models"
    "github.com/drama-generator/backend/pkg/logger"
)

type CharacterAnchorService struct {
    db        *gorm.DB
    log       *logger.Logger
    aiService *AIService
}

// CharacterAnchorConfig 角色锚点配置
type CharacterAnchorConfig struct {
    CharacterID     uint
    DefaultViewType string
    DefaultOutfitID *uint
    ReferenceImages []string
    CharacterPrompt string
}

// GetCharacterDefaultAnchor 获取角色的默认锚点配置
func (s *CharacterAnchorService) GetCharacterDefaultAnchor(characterID uint) (*CharacterAnchorConfig, error) {
    var character models.Character
    if err := s.db.First(&character, characterID).Error; err != nil {
        return nil, err
    }

    config := &CharacterAnchorConfig{
        CharacterID: characterID,
    }

    // 获取主视图
    if character.PrimaryViewID != nil {
        var view models.CharacterView
        if err := s.db.First(&view, *character.PrimaryViewID).Error; err == nil {
            config.DefaultViewType = view.ViewType
            if view.LocalPath != "" {
                config.ReferenceImages = append(config.ReferenceImages, view.LocalPath)
            } else {
                config.ReferenceImages = append(config.ReferenceImages, view.ImageURL)
            }
        }
    }

    // 获取默认服装
    config.DefaultOutfitID = character.DefaultOutfitID

    // 构建角色提示词
    config.CharacterPrompt = s.buildCharacterPrompt(&character)

    return config, nil
}

// SetStoryboardCharacterAnchor 为分镜设置角色锚点
func (s *CharacterAnchorService) SetStoryboardCharacterAnchor(
    storyboardID, characterID uint,
    anchor *models.StoryboardCharacterAnchor,
) error {
    anchor.StoryboardID = storyboardID
    anchor.CharacterID = characterID

    return s.db.Transaction(func(tx *gorm.DB) error {
        // 检查是否已存在锚点
        var existing models.StoryboardCharacterAnchor
        err := tx.Where("storyboard_id = ? AND character_id = ?", storyboardID, characterID).
            First(&existing).Error

        if err == gorm.ErrRecordNotFound {
            // 新建
            return tx.Create(anchor).Error
        } else if err == nil {
            // 更新
            anchor.ID = existing.ID
            return tx.Model(&existing).Updates(anchor).Error
        }

        return err
    })
}

// GetStoryboardCharacterAnchors 获取分镜的所有角色锚点
func (s *CharacterAnchorService) GetStoryboardCharacterAnchors(storyboardID uint) ([]*models.StoryboardCharacterAnchor, error) {
    var anchors []*models.StoryboardCharacterAnchor
    err := s.db.Where("storyboard_id = ?", storyboardID).
        Preload("Character").
        Preload("Character.Views").
        Preload("Outfit").
        Find(&anchors).Error

    return anchors, err
}

// BuildCharacterConsistencyPrompt 构建角色一致性提示词
func (s *CharacterAnchorService) BuildCharacterConsistencyPrompt(anchors []*models.StoryboardCharacterAnchor) string {
    var parts []string

    for _, anchor := range anchors {
        if !anchor.IsVisible {
            continue
        }

        char := anchor.Character
        var charParts []string

        charParts = append(charParts, fmt.Sprintf("角色：%s", char.Name))

        if char.Appearance != nil && *char.Appearance != "" {
            charParts = append(charParts, fmt.Sprintf("外貌：%s", *char.Appearance))
        }

        if anchor.Outfit != nil {
            charParts = append(charParts, fmt.Sprintf("服装：%s", anchor.Outfit.Name))
        }

        if anchor.Pose != nil {
            charParts = append(charParts, fmt.Sprintf("姿态：%s", *anchor.Pose))
        }

        if anchor.Action != nil {
            charParts = append(charParts, fmt.Sprintf("动作：%s", *anchor.Action))
        }

        if anchor.FacialExpression != nil {
            charParts = append(charParts, fmt.Sprintf("表情：%s", *anchor.FacialExpression))
        }

        parts = append(parts, strings.Join(charParts, "，"))
    }

    return strings.Join(parts, "\n")
}

// GetCharacterReferenceImages 获取角色参考图列表（用于AI生成）
func (s *CharacterAnchorService) GetCharacterReferenceImages(characterID uint, viewType string) ([]string, error) {
    var views []models.CharacterView
    query := s.db.Where("character_id = ?", characterID)

    if viewType != "" {
        query = query.Where("view_type = ?", viewType)
    }

    if err := query.Order("is_primary DESC").Find(&views).Error; err != nil {
        return nil, err
    }

    var images []string
    for _, view := range views {
        if view.LocalPath != "" {
            images = append(images, view.LocalPath)
        } else {
            images = append(images, view.ImageURL)
        }
    }

    return images, nil
}

func (s *CharacterAnchorService) buildCharacterPrompt(character *models.Character) string {
    var parts []string

    parts = append(parts, fmt.Sprintf("角色：%s", character.Name))

    if character.Appearance != nil && *character.Appearance != "" {
        parts = append(parts, fmt.Sprintf("外貌：%s", *character.Appearance))
    }

    if character.Personality != nil && *character.Personality != "" {
        parts = append(parts, fmt.Sprintf("性格：%s", *character.Personality))
    }

    return strings.Join(parts, "，")
}
```

#### 2.2.2 增强 FramePromptService

```go
// CharacterAnchorContext 角色锚点上下文
type CharacterAnchorContext struct {
    PrimaryCharacters   []*models.StoryboardCharacterAnchor
    BackgroundCharacters []*models.StoryboardCharacterAnchor
    ReferenceImages     []string
    PromptSections      []string
}

// buildCharacterAnchors 构建角色锚点信息
func (s *FramePromptService) buildCharacterAnchors(storyboardID uint) (*CharacterAnchorContext, error) {
    var anchors []*models.StoryboardCharacterAnchor
    if err := s.db.Where("storyboard_id = ?", storyboardID).
        Preload("Character").
        Preload("Character.Views").
        Preload("Outfit").
        Find(&anchors).Error; err != nil {
        return nil, err
    }

    ctx := &CharacterAnchorContext{
        PrimaryCharacters:    []*models.StoryboardCharacterAnchor{},
        BackgroundCharacters: []*models.StoryboardCharacterAnchor{},
        ReferenceImages:      []string{},
        PromptSections:       []string{},
    }

    for _, anchor := range anchors {
        if !anchor.IsVisible {
            continue
        }

        // 获取角色参考图
        viewImageURL := s.getCharacterViewImage(anchor.CharacterID, anchor.ViewType)

        // 构建角色提示词片段
        charPrompt := s.buildCharacterPrompt(anchor, viewImageURL)
        ctx.PromptSections = append(ctx.PromptSections, charPrompt)

        if viewImageURL != "" {
            ctx.ReferenceImages = append(ctx.ReferenceImages, viewImageURL)
        }

        if anchor.IsPrimary {
            ctx.PrimaryCharacters = append(ctx.PrimaryCharacters, anchor)
        } else {
            ctx.BackgroundCharacters = append(ctx.BackgroundCharacters, anchor)
        }
    }

    return ctx, nil
}

// getCharacterViewImage 获取角色指定视角的参考图
func (s *FramePromptService) getCharacterViewImage(characterID uint, viewType *string) string {
    if viewType == nil {
        return ""
    }

    var view models.CharacterView
    if err := s.db.Where("character_id = ? AND view_type = ?", characterID, *viewType).
        Order("is_primary DESC").
        First(&view).Error; err != nil {
        return ""
    }

    if view.LocalPath != "" {
        return view.LocalPath
    }
    return view.ImageURL
}

// buildCharacterPrompt 构建单个角色的提示词
func (s *FramePromptService) buildCharacterPrompt(anchor *models.StoryboardCharacterAnchor, viewImageURL string) string {
    char := anchor.Character
    var parts []string

    // 基础信息
    parts = append(parts, fmt.Sprintf("角色：%s", char.Name))

    // 外貌特征
    if char.Appearance != nil && *char.Appearance != "" {
        parts = append(parts, fmt.Sprintf("外貌：%s", *char.Appearance))
    }

    // 服装信息
    if anchor.Outfit != nil {
        parts = append(parts, fmt.Sprintf("服装：%s", anchor.Outfit.Name))
        if anchor.Outfit.Description != nil {
            parts = append(parts, *anchor.Outfit.Description)
        }
    }

    // 位置信息
    if anchor.Position != nil {
        positionText := map[string]string{
            "left":        "画面左侧",
            "center":      "画面中央",
            "right":       "画面右侧",
            "foreground":  "前景",
            "background":  "背景",
        }
        if text, ok := positionText[*anchor.Position]; ok {
            parts = append(parts, fmt.Sprintf("位置：%s", text))
        }
    }

    // 姿态和动作
    if anchor.Pose != nil {
        parts = append(parts, fmt.Sprintf("姿态：%s", *anchor.Pose))
    }
    if anchor.Action != nil {
        parts = append(parts, fmt.Sprintf("动作：%s", *anchor.Action))
    }
    if anchor.FacialExpression != nil {
        parts = append(parts, fmt.Sprintf("表情：%s", *anchor.FacialExpression))
    }

    return strings.Join(parts, "，")
}
```

---

## 三、实施计划

### 3.1 Phase 1: 基础三视图系统（2-3周）

**核心功能：**
1. CharacterView 数据模型和迁移
2. 角色视图上传/管理 API
3. 前端视图管理界面
4. 图片生成时附加选定的角色视图

**任务清单：**

| 任务 | 负责人 | 工作量 | 状态 |
|------|--------|--------|------|
| CharacterView 数据模型 | 后端 | 1天 | 待开始 |
| 数据库迁移脚本 | 后端 | 1天 | 待开始 |
| 视图管理 API 开发 | 后端 | 3天 | 待开始 |
| 前端视图管理界面 | 前端 | 4天 | 待开始 |
| 图片生成集成 | 后端 | 2天 | 待开始 |
| 单元测试 | 后端 | 1天 | 待开始 |
| 集成测试 | QA | 2天 | 待开始 |

**小计：14天**

### 3.2 Phase 2: 锚点关联增强（2-3周）

**核心功能：**
1. StoryboardCharacterAnchor 模型
2. 锚点配置 API
3. 前端锚点编辑器
4. 提示词构建增强

**任务清单：**

| 任务 | 负责人 | 工作量 | 状态 |
|------|--------|--------|------|
| StoryboardCharacterAnchor 模型 | 后端 | 2天 | 待开始 |
| 锚点管理 API 开发 | 后端 | 4天 | 待开始 |
| 前端锚点编辑器 | 前端 | 5天 | 待开始 |
| 提示词构建增强 | 后端 | 3天 | 待开始 |
| 单元测试 | 后端 | 2天 | 待开始 |
| 集成测试 | QA | 2天 | 待开始 |

**小计：18天**

### 3.3 Phase 3: 衣橱系统（2周）

**核心功能：**
1. CharacterOutfit 模型
2. 服装管理 API
3. 前端衣橱界面
4. 服装选择逻辑

**任务清单：**

| 任务 | 负责人 | 工作量 | 状态 |
|------|--------|--------|------|
| CharacterOutfit 模型 | 后端 | 1天 | 待开始 |
| 服装管理 API 开发 | 后端 | 3天 | 待开始 |
| 前端衣橱界面 | 前端 | 4天 | 待开始 |
| 服装选择逻辑 | 后端 | 2天 | 待开始 |
| 单元测试 | 后端 | 1天 | 待开始 |
| 集成测试 | QA | 2天 | 待开始 |

**小计：13天**

### 3.4 Phase 4: 高级特性与优化（2-3周）

**核心功能：**
1. 角色一致性算法优化
2. 批量操作支持
3. 性能优化
4. 文档完善

**任务清单：**

| 任务 | 负责人 | 工作量 | 状态 |
|------|--------|--------|------|
| 角色一致性算法优化 | 算法 | 4天 | 待开始 |
| 批量操作支持 | 后端 | 3天 | 待开始 |
| 性能优化 | 后端 | 3天 | 待开始 |
| 文档完善 | 技术文档 | 3天 | 待开始 |
| 用户测试 | 产品 | 3天 | 待开始 |

**小计：16天**

### 3.5 总体时间表

| 阶段 | 工期 | 开始日期 | 结束日期 |
|------|------|---------|---------|
| Phase 1 | 14天 | 待定 | 待定 |
| Phase 2 | 18天 | 待定 | 待定 |
| Phase 3 | 13天 | 待定 | 待定 |
| Phase 4 | 16天 | 待定 | 待定 |
| **总计** | **61天** | | |

**约 2.5 个月（按每月20工作日计算）**

---

## 四、风险评估

### 4.1 技术风险

| 风险项 | 影响 | 概率 | 缓解措施 | 负责人 |
|--------|------|------|----------|--------|
| **AI 角色一致性效果不理想** | 🔴 高 | 🟡 中 | 1. 增加提示词工程调优<br>2. 预留人工调整接口<br>3. 支持多参考图组合 | 算法工程师 |
| **数据库迁移复杂度** | 🟡 中 | 🟢 低 | 1. 渐进式迁移<br>2. 保持向后兼容<br>3. 充分测试迁移脚本 | 后端开发 |
| **性能影响** | 🟡 中 | 🟡 中 | 1. 引入缓存机制<br>2. 异步处理<br>3. 数据库索引优化 | 后端开发 |
| **用户学习成本** | 🟡 中 | 🟡 中 | 1. 提供预设模板<br>2. 智能默认值<br>3. 详细的帮助文档 | 产品+设计 |

### 4.2 业务风险

| 风险项 | 影响 | 概率 | 缓解措施 | 负责人 |
|--------|------|------|----------|--------|
| **需求变更** | 🟡 中 | 🟡 中 | 1. 敏捷开发<br>2. 快速迭代<br>3. 用户反馈机制 | 产品经理 |
| **资源不足** | 🔴 高 | 🟢 低 | 1. 合理排期<br>2. 外部协作<br>3. 优先级管理 | 项目经理 |

### 4.3 风险应对策略

**高风险项目应对：**
1. **AI 效果不理想**
   - 准备多套提示词模板
   - A/B 测试验证
   - 保留手动调整能力

2. **资源不足**
   - 分阶段交付
   - 核心功能优先
   - 外部资源补充

---

## 五、API设计

### 5.1 角色视图管理 API

#### 5.1.1 上传角色视图

```http
POST /api/v1/characters/:id/views
Content-Type: multipart/form-data

{
    "view_type": "front",        # 必填：视图类型
    "image": <file>,             # 必填：视图图片
    "description": "正面标准站姿", # 可选：描述
    "is_primary": true            # 可选：是否为主视图
}

# 响应
{
    "success": true,
    "data": {
        "id": 1,
        "character_id": 10,
        "view_type": "front",
        "image_url": "https://...",
        "local_path": "/data/storage/...",
        "is_primary": true,
        "created_at": "2025-02-22T10:00:00Z"
    }
}
```

#### 5.1.2 获取角色视图列表

```http
GET /api/v1/characters/:id/views

# 响应
{
    "success": true,
    "data": [
        {
            "id": 1,
            "view_type": "front",
            "image_url": "https://...",
            "is_primary": true
        },
        {
            "id": 2,
            "view_type": "side",
            "image_url": "https://...",
            "is_primary": false
        }
    ]
}
```

#### 5.1.3 设置主视图

```http
PUT /api/v1/characters/:id/views/:viewId/primary

# 响应
{
    "success": true,
    "message": "主视图设置成功"
}
```

#### 5.1.4 删除视图

```http
DELETE /api/v1/characters/:id/views/:viewId

# 响应
{
    "success": true,
    "message": "视图删除成功"
}
```

### 5.2 角色服装管理 API

#### 5.2.1 添加服装

```http
POST /api/v1/characters/:id/outfits
Content-Type: multipart/form-data

{
    "name": "校服",              # 必填：服装名称
    "image": <file>,             # 必填：服装图片
    "description": "蓝白相间校服", # 可选：描述
    "tags": "school,casual",      # 可选：标签
    "is_default": false           # 可选：是否默认
}

# 响应
{
    "success": true,
    "data": {
        "id": 1,
        "character_id": 10,
        "name": "校服",
        "image_url": "https://...",
        "tags": "school,casual",
        "is_default": false
    }
}
```

#### 5.2.2 获取服装列表

```http
GET /api/v1/characters/:id/outfits

# 响应
{
    "success": true,
    "data": [
        {
            "id": 1,
            "name": "校服",
            "image_url": "https://...",
            "is_default": true
        }
    ]
}
```

#### 5.2.3 设置默认服装

```http
PUT /api/v1/characters/:id/outfits/:outfitId/default

# 响应
{
    "success": true,
    "message": "默认服装设置成功"
}
```

### 5.3 分镜锚点管理 API

#### 5.3.1 获取分镜的角色锚点

```http
GET /api/v1/storyboards/:id/character-anchors

# 响应
{
    "success": true,
    "data": [
        {
            "id": 1,
            "storyboard_id": 100,
            "character_id": 10,
            "character": {
                "id": 10,
                "name": "张三",
                "image_url": "https://..."
            },
            "view_type": "front",
            "outfit_id": 1,
            "position": "center",
            "pose": "站立",
            "action": "向前走",
            "is_primary": true,
            "is_visible": true
        }
    ]
}
```

#### 5.3.2 添加角色锚点

```http
POST /api/v1/storyboards/:id/character-anchors
Content-Type: application/json

{
    "character_id": 10,
    "view_type": "front",
    "outfit_id": 1,
    "position": "center",
    "depth": "midground",
    "scale": 1.0,
    "pose": "站立",
    "action": "向前走",
    "facial_expression": "微笑",
    "is_primary": true,
    "visibility": 1.0,
    "is_visible": true
}

# 响应
{
    "success": true,
    "data": {
        "id": 1,
        "storyboard_id": 100,
        "character_id": 10,
        "view_type": "front",
        "position": "center",
        "is_primary": true
    }
}
```

#### 5.3.3 更新锚点配置

```http
PUT /api/v1/storyboards/:id/character-anchors/:anchorId
Content-Type: application/json

{
    "view_type": "side",
    "position": "left",
    "pose": "坐下"
}

# 响应
{
    "success": true,
    "data": {
        "id": 1,
        "view_type": "side",
        "position": "left"
    }
}
```

#### 5.3.4 批量设置锚点

```http
POST /api/v1/storyboards/:id/character-anchors/batch
Content-Type: application/json

{
    "anchors": [
        {
            "character_id": 10,
            "view_type": "front",
            "position": "center"
        },
        {
            "character_id": 11,
            "view_type": "side",
            "position": "left"
        }
    ]
}

# 响应
{
    "success": true,
    "message": "成功设置 2 个角色锚点",
    "data": {
        "created": 2,
        "updated": 0
    }
}
```

#### 5.3.5 删除锚点

```http
DELETE /api/v1/storyboards/:id/character-anchors/:anchorId

# 响应
{
    "success": true,
    "message": "角色锚点删除成功"
}
```

### 5.4 AI生成增强 API

#### 5.4.1 使用角色锚点生成图片

```http
POST /api/v1/images/generate
Content-Type: application/json

{
    "storyboard_id": 123,
    "use_character_anchors": true,    # 新增：是否使用角色锚点
    "character_anchor_mode": "smart", # 新增：auto, smart, manual
    "model": "dall-e-3",
    "size": "1024x1024"
}

# 响应
{
    "success": true,
    "data": {
        "id": 456,
        "prompt": "角色：张三，外貌：瘦高...",
        "image_url": "https://...",
        "character_references": [      # 新增：角色参考信息
            {
                "character_id": 1,
                "character_name": "张三",
                "view_type": "front",
                "reference_image": "data:image/jpeg;base64,..."
            }
        ]
    }
}
```

---

## 六、UI设计

### 6.1 角色三视图管理界面

**组件：`CharacterViewsPanel.vue`**

```vue
<template>
  <div class="character-views-panel">
    <div class="panel-header">
      <h3>角色三视图</h3>
      <el-button type="primary" size="small" @click="uploadView">
        <el-icon><Plus /></el-icon>
        添加视图
      </el-button>
    </div>

    <div class="views-grid">
      <!-- 视图卡片 -->
      <div
        v-for="view in views"
        :key="view.id"
        class="view-item"
        :class="{ active: view.is_primary }"
      >
        <div class="view-image">
          <img :src="view.image_url" :alt="view.view_type" />
          <el-tag size="small" class="view-type-tag">
            {{ viewTypeLabels[view.view_type] }}
          </el-tag>
          <el-tag v-if="view.is_primary" type="success" size="small" class="primary-tag">
            主视图
          </el-tag>
        </div>

        <div class="view-info">
          <p v-if="view.description">{{ view.description }}</p>
        </div>

        <div class="view-actions">
          <el-button
            size="small"
            @click="setPrimary(view.id)"
            :disabled="view.is_primary"
          >
            {{ view.is_primary ? '主视图' : '设为主视图' }}
          </el-button>
          <el-button size="small" @click="editView(view)">编辑</el-button>
          <el-button
            size="small"
            type="danger"
            @click="removeView(view.id)"
          >
            删除
          </el-button>
        </div>
      </div>
    </div>

    <!-- 上传对话框 -->
    <el-dialog v-model="uploadDialogVisible" title="添加角色视图" width="500px">
      <el-form :model="uploadForm" label-width="80px">
        <el-form-item label="视图类型">
          <el-select v-model="uploadForm.view_type">
            <el-option label="正面视图" value="front" />
            <el-option label="侧面视图" value="side" />
            <el-option label="背面视图" value="back" />
            <el-option label="三分之四视图" value="three_quarter" />
          </el-select>
        </el-form-item>

        <el-form-item label="视图图片">
          <el-upload
            :auto-upload="false"
            :on-change="handleFileChange"
            accept="image/*"
          >
            <el-button>选择图片</el-button>
          </el-upload>
        </el-form-item>

        <el-form-item label="描述">
          <el-input
            v-model="uploadForm.description"
            type="textarea"
            :rows="3"
          />
        </el-form-item>

        <el-form-item label="设为主视图">
          <el-switch v-model="uploadForm.is_primary" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitUpload">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { characterApi } from '@/api/character'

const props = defineProps<{
  characterId: number
}>()

const views = ref([])
const uploadDialogVisible = ref(false)
const uploadForm = ref({
  view_type: 'front',
  description: '',
  is_primary: false
})

const viewTypeLabels = {
  front: '正面',
  side: '侧面',
  back: '背面',
  three_quarter: '3/4'
}

const loadViews = async () => {
  try {
    const { data } = await characterApi.getViews(props.characterId)
    views.value = data
  } catch (error) {
    ElMessage.error('加载视图失败')
  }
}

const uploadView = () => {
  uploadForm.value = {
    view_type: 'front',
    description: '',
    is_primary: false
  }
  uploadDialogVisible.value = true
}

const setPrimary = async (viewId: number) => {
  try {
    await characterApi.setViewPrimary(props.characterId, viewId)
    ElMessage.success('主视图设置成功')
    loadViews()
  } catch (error) {
    ElMessage.error('设置失败')
  }
}

const removeView = async (viewId: number) => {
  try {
    await characterApi.deleteView(props.characterId, viewId)
    ElMessage.success('删除成功')
    loadViews()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadViews()
})
</script>

<style scoped>
.character-views-panel {
  padding: 20px;
}

.views-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
  margin-top: 16px;
}

.view-item {
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.3s;
}

.view-item.active {
  border-color: #10b981;
  box-shadow: 0 0 0 2px rgba(16, 185, 129, 0.2);
}

.view-image {
  position: relative;
  aspect-ratio: 1;
}

.view-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.view-type-tag {
  position: absolute;
  top: 8px;
  left: 8px;
}

.primary-tag {
  position: absolute;
  top: 8px;
  right: 8px;
}

.view-actions {
  padding: 12px;
  display: flex;
  gap: 8px;
}
</style>
```

### 6.2 衣橱管理界面

**组件：`CharacterWardrobePanel.vue`**

```vue
<template>
  <div class="wardrobe-panel">
    <div class="panel-header">
      <h3>角色衣橱</h3>
      <el-button type="primary" size="small" @click="addOutfit">
        <el-icon><Plus /></el-icon>
        添加造型
      </el-button>
    </div>

    <div class="outfit-list">
      <!-- 服装卡片 -->
      <div
        v-for="outfit in outfits"
        :key="outfit.id"
        class="outfit-item"
        :class="{ active: outfit.is_default }"
      >
        <div class="outfit-image">
          <img :src="outfit.image_url" :alt="outfit.name" />
          <el-tag v-if="outfit.is_default" type="success" class="default-tag">
            默认
          </el-tag>
        </div>

        <div class="outfit-info">
          <h4>{{ outfit.name }}</h4>
          <p v-if="outfit.description">{{ outfit.description }}</p>
          <div class="outfit-tags" v-if="outfit.tags">
            <el-tag
              v-for="tag in outfit.tags.split(',')"
              :key="tag"
              size="small"
            >
              {{ tag }}
            </el-tag>
          </div>
        </div>

        <div class="outfit-actions">
          <el-button
            size="small"
            @click="setDefault(outfit.id)"
            :disabled="outfit.is_default"
          >
            {{ outfit.is_default ? '默认造型' : '设为默认' }}
          </el-button>
          <el-button size="small" @click="editOutfit(outfit)">编辑</el-button>
        </div>
      </div>
    </div>

    <!-- 添加/编辑对话框 -->
    <el-dialog
      v-model="outfitDialogVisible"
      :title="editingOutfit ? '编辑造型' : '添加造型'"
      width="500px"
    >
      <el-form :model="outfitForm" label-width="80px">
        <el-form-item label="造型名称">
          <el-input v-model="outfitForm.name" placeholder="如：校服" />
        </el-form-item>

        <el-form-item label="造型图片">
          <el-upload
            :auto-upload="false"
            :on-change="handleOutfitFileChange"
            accept="image/*"
          >
            <el-button>选择图片</el-button>
          </el-upload>
        </el-form-item>

        <el-form-item label="描述">
          <el-input
            v-model="outfitForm.description"
            type="textarea"
            :rows="3"
          />
        </el-form-item>

        <el-form-item label="标签">
          <el-input v-model="outfitForm.tags" placeholder="如：school,casual" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="outfitDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitOutfit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { characterApi } from '@/api/character'

const props = defineProps<{
  characterId: number
}>()

const outfits = ref([])
const outfitDialogVisible = ref(false)
const editingOutfit = ref(null)
const outfitForm = ref({
  name: '',
  description: '',
  tags: '',
  is_default: false
})

const loadOutfits = async () => {
  try {
    const { data } = await characterApi.getOutfits(props.characterId)
    outfits.value = data
  } catch (error) {
    ElMessage.error('加载衣橱失败')
  }
}

const addOutfit = () => {
  editingOutfit.value = null
  outfitForm.value = {
    name: '',
    description: '',
    tags: '',
    is_default: false
  }
  outfitDialogVisible.value = true
}

const setDefault = async (outfitId: number) => {
  try {
    await characterApi.setOutfitDefault(props.characterId, outfitId)
    ElMessage.success('默认造型设置成功')
    loadOutfits()
  } catch (error) {
    ElMessage.error('设置失败')
  }
}

onMounted(() => {
  loadOutfits()
})
</script>

<style scoped>
.wardrobe-panel {
  padding: 20px;
}

.outfit-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 16px;
  margin-top: 16px;
}

.outfit-item {
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
}

.outfit-item.active {
  border-color: #10b981;
}

.outfit-image {
  position: relative;
  aspect-ratio: 1;
}

.outfit-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.default-tag {
  position: absolute;
  top: 8px;
  right: 8px;
}

.outfit-info {
  padding: 12px;
}

.outfit-info h4 {
  margin: 0 0 8px 0;
}

.outfit-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
  margin-top: 8px;
}

.outfit-actions {
  padding: 12px;
  border-top: 1px solid #e5e7eb;
  display: flex;
  gap: 8px;
}
</style>
```

### 6.3 分镜锚点编辑器

**组件：`StoryboardCharacterAnchor.vue`**

```vue
<template>
  <div class="character-anchor-editor">
    <div class="editor-header">
      <h3>角色锚点配置</h3>
      <el-button type="primary" size="small" @click="addCharacter">
        <el-icon><Plus /></el-icon>
        添加角色
      </el-button>
    </div>

    <div class="anchor-list">
      <div
        v-for="anchor in anchors"
        :key="anchor.id"
        class="anchor-item"
      >
        <!-- 角色头部 -->
        <div class="anchor-header">
          <div class="character-info">
            <el-avatar :size="40" :src="anchor.character?.image_url" />
            <span>{{ anchor.character?.name }}</span>
          </div>
          <div class="header-actions">
            <el-switch
              v-model="anchor.is_visible"
              @change="updateVisibility(anchor)"
              active-text="显示"
              inactive-text="隐藏"
            />
            <el-button
              size="small"
              type="danger"
              @click="removeAnchor(anchor.id)"
            >
              移除
            </el-button>
          </div>
        </div>

        <!-- 锚点配置 -->
        <div class="anchor-config" v-if="anchor.is_visible">
          <!-- 视角选择 -->
          <div class="config-row">
            <label>视角：</label>
            <el-radio-group v-model="anchor.view_type" @change="updateAnchor(anchor)">
              <el-radio-button label="front">正面</el-radio-button>
              <el-radio-button label="side">侧面</el-radio-button>
              <el-radio-button label="back">背面</el-radio-button>
            </el-radio-group>
          </div>

          <!-- 服装选择 -->
          <div class="config-row">
            <label>服装：</label>
            <el-select v-model="anchor.outfit_id" @change="updateAnchor(anchor)">
              <el-option label="默认造型" :value="null" />
              <el-option
                v-for="outfit in getCharacterOutfits(anchor.character_id)"
                :key="outfit.id"
                :label="outfit.name"
                :value="outfit.id"
              />
            </el-select>
          </div>

          <!-- 位置选择 -->
          <div class="config-row">
            <label>位置：</label>
            <div class="position-grid">
              <div
                v-for="pos in positionOptions"
                :key="pos.value"
                class="position-btn"
                :class="{ active: anchor.position === pos.value }"
                @click="setPosition(anchor, pos.value)"
              >
                {{ pos.label }}
              </div>
            </div>
          </div>

          <!-- 景深选择 -->
          <div class="config-row">
            <label>景深：</label>
            <el-radio-group v-model="anchor.depth" @change="updateAnchor(anchor)">
              <el-radio-button label="foreground">前景</el-radio-button>
              <el-radio-button label="midground">中景</el-radio-button>
              <el-radio-button label="background">后景</el-radio-button>
            </el-radio-group>
          </div>

          <!-- 姿态描述 -->
          <div class="config-row">
            <label>姿态：</label>
            <el-input
              v-model="anchor.pose"
              type="textarea"
              :rows="2"
              placeholder="描述角色的姿态，如：站立、坐下、弯腰..."
              @blur="updateAnchor(anchor)"
            />
          </div>

          <!-- 动作描述 -->
          <div class="config-row">
            <label>动作：</label>
            <el-input
              v-model="anchor.action"
              type="textarea"
              :rows="2"
              placeholder="描述角色的动作，如：向前走、回头、挥手..."
              @blur="updateAnchor(anchor)"
            />
          </div>

          <!-- 表情描述 -->
          <div class="config-row">
            <label>表情：</label>
            <el-input
              v-model="anchor.facial_expression"
              placeholder="如：微笑、愤怒、悲伤..."
              @blur="updateAnchor(anchor)"
            />
          </div>

          <!-- 高级选项 -->
          <div class="config-row advanced">
            <label>高级选项：</label>
            <div class="advanced-options">
              <div class="option-item">
                <span>主要角色：</span>
                <el-switch v-model="anchor.is_primary" @change="updateAnchor(anchor)" />
              </div>
              <div class="option-item">
                <span>可见度：</span>
                <el-slider v-model="anchor.visibility" :min="0" :max="1" :step="0.1" @change="updateAnchor(anchor)" />
              </div>
              <div class="option-item">
                <span>缩放：</span>
                <el-slider v-model="anchor.scale" :min="0.5" :max="2" :step="0.1" @change="updateAnchor(anchor)" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { storyboardApi } from '@/api/storyboard'
import { characterApi } from '@/api/character'

const props = defineProps<{
  storyboardId: number
}>()

const anchors = ref([])
const characterOutfits = ref({})

const positionOptions = [
  { label: '左', value: 'left' },
  { label: '中', value: 'center' },
  { label: '右', value: 'right' },
  { label: '前', value: 'foreground' },
  { label: '后', value: 'background' }
]

const loadAnchors = async () => {
  try {
    const { data } = await storyboardApi.getCharacterAnchors(props.storyboardId)
    anchors.value = data
  } catch (error) {
    ElMessage.error('加载锚点失败')
  }
}

const getCharacterOutfits = (characterId: number) => {
  return characterOutfits.value[characterId] || []
}

const updateAnchor = async (anchor: any) => {
  try {
    await storyboardApi.updateCharacterAnchor(props.storyboardId, anchor.id, anchor)
    ElMessage.success('更新成功')
  } catch (error) {
    ElMessage.error('更新失败')
  }
}

const setPosition = async (anchor: any, position: string) => {
  anchor.position = position
  await updateAnchor(anchor)
}

onMounted(() => {
  loadAnchors()
})
</script>

<style scoped>
.character-anchor-editor {
  padding: 20px;
}

.anchor-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-top: 16px;
}

.anchor-item {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
}

.anchor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
}

.character-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.anchor-config {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.config-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.config-row label {
  min-width: 60px;
  font-weight: 500;
}

.position-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 8px;
}

.position-btn {
  padding: 8px 16px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
}

.position-btn:hover {
  background: #f3f4f6;
}

.position-btn.active {
  background: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

.advanced {
  flex-direction: column;
}

.advanced-options {
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex: 1;
}

.option-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
</style>
```

---

## 附录

### A. 数据库迁移脚本

```sql
-- ============================================
-- 角色锚点系统优化 - 数据库迁移脚本
-- 版本: v1.0
-- 日期: 2025-02-22
-- ============================================

-- 1. 创建角色视图表
CREATE TABLE IF NOT EXISTS character_views (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL,
    view_type VARCHAR(20) NOT NULL, -- front, side, back, three_quarter
    image_url VARCHAR(500) NOT NULL,
    local_path VARCHAR(500),
    description TEXT,
    is_primary BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE,
    CONSTRAINT fk_char_views_character FOREIGN KEY (character_id) REFERENCES characters(id)
);

CREATE INDEX idx_character_views_character_id ON character_views(character_id);
CREATE INDEX idx_character_views_view_type ON character_views(view_type);
CREATE INDEX idx_character_views_deleted_at ON character_views(deleted_at);

-- 2. 创建角色服装表
CREATE TABLE IF NOT EXISTS character_outfits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    image_url VARCHAR(500) NOT NULL,
    local_path VARCHAR(500),
    tags VARCHAR(500),
    is_default BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);

CREATE INDEX idx_character_outfits_character_id ON character_outfits(character_id);
CREATE INDEX idx_character_outfits_deleted_at ON character_outfits(deleted_at);

-- 3. 创建分镜角色锚点表
CREATE TABLE IF NOT EXISTS storyboard_character_anchors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    storyboard_id INTEGER NOT NULL,
    character_id INTEGER NOT NULL,
    view_type VARCHAR(20),
    outfit_id INTEGER,
    position VARCHAR(50),
    depth VARCHAR(20),
    scale REAL,
    pose TEXT,
    action TEXT,
    facial_expression VARCHAR(100),
    is_primary BOOLEAN DEFAULT 0,
    visibility REAL DEFAULT 1.0,
    is_visible BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (storyboard_id) REFERENCES storyboards(id) ON DELETE CASCADE,
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE,
    FOREIGN KEY (outfit_id) REFERENCES character_outfits(id) ON DELETE SET NULL
);

CREATE INDEX idx_sca_storyboard ON storyboard_character_anchors(storyboard_id);
CREATE INDEX idx_sca_character ON storyboard_character_anchors(character_id);
CREATE INDEX idx_sca_deleted_at ON storyboard_character_anchors(deleted_at);

-- 4. 修改角色表，添加新字段
ALTER TABLE characters ADD COLUMN primary_view_id INTEGER;
ALTER TABLE characters ADD COLUMN default_outfit_id INTEGER;
ALTER TABLE characters ADD COLUMN voice_seed VARCHAR(100);

CREATE INDEX idx_characters_primary_view ON characters(primary_view_id);
CREATE INDEX idx_characters_default_outfit ON characters(default_outfit_id);

-- 5. 迁移现有数据（如果有）
-- 将现有的 character image_url 转换为主视图
-- INSERT INTO character_views (character_id, view_type, image_url, local_path, is_primary)
-- SELECT id, 'front', image_url, local_path, 1
-- FROM characters
-- WHERE image_url IS NOT NULL;

-- 6. 创建触发器（可选，用于自动更新 updated_at）
CREATE TRIGGER IF NOT EXISTS update_character_views_timestamp
AFTER UPDATE ON character_views
BEGIN
    UPDATE character_views SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_character_outfits_timestamp
AFTER UPDATE ON character_outfits
BEGIN
    UPDATE character_outfits SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_storyboard_character_anchors_timestamp
AFTER UPDATE ON storyboard_character_anchors
BEGIN
    UPDATE storyboard_character_anchors SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
```

### B. 相关文档

- [DATA_MIGRATION.md](./DATA_MIGRATION.md) - 数据迁移指南
- [API.md](./API.md) - API 文档
- [UI_GUIDE.md](./UI_GUIDE.md) - UI 设计指南

---

**文档变更历史：**

| 版本 | 日期 | 变更内容 | 作者 |
|------|------|---------|------|
| v1.0 | 2025-02-22 | 初始版本 | Claude Code |

---

*如有疑问或建议，请联系开发团队*
