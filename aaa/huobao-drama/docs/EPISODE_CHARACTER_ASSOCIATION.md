# Episode-Character 关联关系重新设计

## 设计决策记录

**日期**: 2026-02-01
**状态**: 提议中
**作者**: Claude Code
**优先级**: 中

---

## 1. 问题分析

### 1.1 当前实现

当前使用 GORM 的 `many2many` 自动关联：

```go
// domain/models/drama.go
type Character struct {
    Episodes []Episode `gorm:"many2many:episode_characters;"`
}

type Episode struct {
    Characters []Character `gorm:"many2many:episode_characters;"`
}
```

**自动生成的关联表结构**:
```sql
CREATE TABLE episode_characters (
    episode_id BIGINT UNSIGNED NOT NULL,
    character_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (episode_id, character_id)
);
```

### 1.2 现有实现的局限性

| 局限性 | 影响 | 业务场景 |
|--------|------|----------|
| 无法记录角色定位 | 无法区分主角/配角/客串 | 片酬计算、演员署名 |
| 无出场顺序信息 | 角色列表排序混乱 | 片头字幕顺序 |
| 无戏份统计 | 无法追踪角色重要性 | 剧情分析、运营数据 |
| 无时间戳 | 无法追溯关联历史 | 审计需求 |
| 无软删除支持 | 关联删除不可逆 | 数据恢复需求 |
| 难以扩展关联属性 | 每次变更需改表结构 | 业务迭代缓慢 |

### 1.3 业务需求清单

根据 `huobao-drama` 项目特性，以下需求已确认：

- [x] 角色定位（主角/配角/客串）
- [x] 出场顺序（片头字幕排序）
- [x] 戏份统计（出场时长/台词量）
- [x] 审计追溯（创建/更新时间）
- [ ] 片酬信息（敏感字段，待确认）
- [ ] 配音演员（多语言支持）
- [ ] 角色状态（是否已出镜/待拍摄）

---

## 2. 推荐设计方案

### 2.1 架构选择

**✅ 选择方案**: 业务关联表（显式关联）

**决策依据**:
1. 关联表需要承载业务字段（角色定位、出场顺序）
2. 需要支持复杂查询（按角色定位筛选、统计）
3. 需要审计能力（时间戳、软删除）
4. 便于后续扩展（片酬、配音等）

### 2.2 数据模型设计

#### 2.2.1 EpisodeCharacter 模型

```go
// domain/models/episode_character.go
package models

import (
    "time"
    "gorm.io/gorm"
)

// EpisodeCharacter 剧集与角色的关联关系
// 承载业务属性：角色定位、出场顺序、戏份统计等
type EpisodeCharacter struct {
    // 主键
    ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

    // 外键
    EpisodeID   uint `gorm:"not null;index:idx_episode_character,priority:1" json:"episode_id"`
    CharacterID uint `gorm:"not null;index:idx_episode_character,priority:2" json:"character_id"`

    // 业务字段
    RoleType    string `gorm:"type:varchar(20);not null;default:'supporting';check:role_type IN ('main', 'supporting', 'guest', 'extra')" json:"role_type"`
    SortOrder   int    `gorm:"not null;default:0;index" json:"sort_order"` // 出场顺序，片头字幕排序
    ScreenTime  int    `gorm:"default:0" json:"screen_time"`               // 出场时长（秒）
    LineCount   int    `gorm:"default:0" json:"line_count"`                // 台词数量

    // 可选扩展字段（根据业务需求启用）
    Description    *string `gorm:"type:text" json:"description"`             // 角色在本集中的特别说明
    CostumeNote    *string `gorm:"type:varchar(500)" json:"costume_note"`    // 服装备注
    VoiceActorID   *uint   `gorm:"index" json:"voice_actor_id,omitempty"`    // 配音演员ID
    SalaryAmount   *int64  `gorm:"-" json:"salary_amount,omitempty"`          // 片酬（不存储，仅运行时使用）

    // 状态字段
    HasAppeared    bool   `gorm:"default:false" json:"has_appeared"`    // 是否已出镜
    FilmingStatus  string `gorm:"type:varchar(20);default:'pending';check:filming_status IN ('pending', 'filming', 'completed')" json:"filming_status"`

    // 元数据
    CreatedAt time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
    UpdatedAt time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

    // 关联
    Episode   Episode   `gorm:"foreignKey:EpisodeID" json:"episode,omitempty"`
    Character Character `gorm:"foreignKey:CharacterID" json:"character,omitempty"`
}

// TableName 指定表名
func (EpisodeCharacter) TableName() string {
    return "episode_characters"
}

// RoleType 枚举值
const (
    RoleTypeMain        = "main"        // 主角
    RoleTypeSupporting  = "supporting"  // 配角
    RoleTypeGuest       = "guest"       // 客串
    RoleTypeExtra       = "extra"       // 群演
)

// FilmingStatus 枚举值
const (
    FilmingStatusPending   = "pending"   // 待拍摄
    FilmingStatusFilming   = "filming"   // 拍摄中
    FilmingStatusCompleted = "completed" // 已完成
)
```

#### 2.2.2 更新 Character 模型

```go
// domain/models/drama.go
type Character struct {
    // ... 现有字段 ...

    // 替换 many2many 为显式关联
    EpisodeCharacters []EpisodeCharacter `gorm:"foreignKey:CharacterID" json:"episode_characters,omitempty"`
}

// 新增辅助方法：获取角色在指定剧集中的定位
func (c *Character) GetRoleInEpisode(episodeID uint) *EpisodeCharacter {
    for _, ec := range c.EpisodeCharacters {
        if ec.EpisodeID == episodeID {
            return &ec
        }
    }
    return nil
}
```

#### 2.2.3 更新 Episode 模型

```go
// domain/models/drama.go
type Episode struct {
    // ... 现有字段 ...

    // 替换 many2many 为显式关联
    EpisodeCharacters []EpisodeCharacter `gorm:"foreignKey:EpisodeID" json:"episode_characters,omitempty"`
    // 保留 Characters 用于简单查询（通过 EpisodeCharacters 关联）
    Characters []Character `gorm:"-" json:"characters,omitempty"` // 运行时字段
}

// 新增辅助方法：获取主角列表
func (e *Episode) GetMainCharacters() []Character {
    var mains []Character
    for _, ec := range e.EpisodeCharacters {
        if ec.RoleType == RoleTypeMain {
            mains = append(mains, ec.Character)
        }
    }
    return mains
}

// 新增辅助方法：按出场顺序获取角色
func (e *Episode) GetSortedCharacters() []Character {
    chars := make([]Character, len(e.EpisodeCharacters))
    for _, ec := range e.EpisodeCharacters {
        chars[ec.SortOrder] = ec.Character
    }
    return chars
}
```

### 2.3 数据库迁移 SQL

```sql
-- migrations/episode_characters_refactor.sql

-- 1. 备份现有数据
CREATE TABLE episode_characters_backup AS
SELECT * FROM episode_characters;

-- 2. 删除旧表（GORM 自动创建的）
DROP TABLE episode_characters;

-- 3. 创建新表
CREATE TABLE episode_characters (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,

    -- 外键
    episode_id BIGINT UNSIGNED NOT NULL,
    character_id BIGINT UNSIGNED NOT NULL,

    -- 业务字段
    role_type VARCHAR(20) NOT NULL DEFAULT 'supporting',
    sort_order INT NOT NULL DEFAULT 0,
    screen_time INT NOT NULL DEFAULT 0,
    line_count INT NOT NULL DEFAULT 0,

    -- 可选字段
    description TEXT,
    costume_note VARCHAR(500),
    voice_actor_id BIGINT UNSIGNED,

    -- 状态字段
    has_appeared BOOLEAN NOT NULL DEFAULT FALSE,
    filming_status VARCHAR(20) NOT NULL DEFAULT 'pending',

    -- 元数据
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,

    -- 联合唯一索引
    UNIQUE KEY idx_episode_character (episode_id, character_id),

    -- 索引
    INDEX idx_sort_order (sort_order),
    INDEX idx_role_type (role_type),
    INDEX idx_deleted_at (deleted_at),

    -- 外键约束
    FOREIGN KEY (episode_id) REFERENCES episodes(id) ON DELETE CASCADE,
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE,
    FOREIGN KEY (voice_actor_id) REFERENCES voice_actors(id) ON DELETE SET NULL,

    -- 检查约束
    CONSTRAINT chk_role_type CHECK (role_type IN ('main', 'supporting', 'guest', 'extra')),
    CONSTRAINT chk_filming_status CHECK (filming_status IN ('pending', 'filming', 'completed'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. 恢复备份数据
INSERT INTO episode_characters (episode_id, character_id, created_at)
SELECT episode_id, character_id, created_at
FROM episode_characters_backup;

-- 5. 更新默认角色定位（可根据业务规则调整）
UPDATE episode_characters SET role_type = 'main' WHERE sort_order <= 3;
UPDATE episode_characters SET role_type = 'supporting' WHERE sort_order > 3 AND sort_order <= 10;
UPDATE episode_characters SET role_type = 'guest' WHERE sort_order > 10;

-- 6. 清理
DROP TABLE episode_characters_backup;
```

---

## 3. 服务层改造

### 3.1 EpisodeCharacterService (新增)

```go
// application/services/episode_character_service.go
package services

import (
    "errors"
    "github.com/huobao-drama/domain/models"
    "gorm.io/gorm"
)

type EpisodeCharacterService struct {
    db *gorm.DB
}

func NewEpisodeCharacterService(db *gorm.DB) *EpisodeCharacterService {
    return &EpisodeCharacterService{db: db}
}

// CreateEpisodeCharacter 创建剧集角色关联
func (s *EpisodeCharacterService) CreateEpisodeCharacter(req *CreateEpisodeCharacterRequest) (*models.EpisodeCharacter, error) {
    // 验证关联是否已存在
    var existing models.EpisodeCharacter
    err := s.db.Unscoped().Where("episode_id = ? AND character_id = ?", req.EpisodeID, req.CharacterID).First(&existing).Error
    if err == nil && existing.DeletedAt.Valid {
        // 如果已软删除，则恢复
        existing.DeletedAt = gorm.DeletedAt{}
        existing.RoleType = req.RoleType
        existing.SortOrder = req.SortOrder
        if err := s.db.Save(&existing).Error; err != nil {
            return nil, err
        }
        return &existing, nil
    } else if err == nil {
        return nil, errors.New("关联已存在")
    }

    // 创建新关联
    ec := &models.EpisodeCharacter{
        EpisodeID:   req.EpisodeID,
        CharacterID: req.CharacterID,
        RoleType:    req.RoleType,
        SortOrder:   req.SortOrder,
        ScreenTime:  0,
        LineCount:   0,
        HasAppeared: false,
        FilmingStatus: models.FilmingStatusPending,
    }

    if err := s.db.Create(ec).Error; err != nil {
        return nil, err
    }

    // 预加载关联数据
    s.db.Preload("Character").Preload("Episode").First(ec, ec.ID)
    return ec, nil
}

// UpdateEpisodeCharacter 更新剧集角色关联
func (s *EpisodeCharacterService) UpdateEpisodeCharacter(id uint, req *UpdateEpisodeCharacterRequest) (*models.EpisodeCharacter, error) {
    var ec models.EpisodeCharacter
    if err := s.db.First(&ec, id).Error; err != nil {
        return nil, err
    }

    updates := make(map[string]interface{})
    if req.RoleType != nil {
        updates["role_type"] = *req.RoleType
    }
    if req.SortOrder != nil {
        updates["sort_order"] = *req.SortOrder
    }
    if req.ScreenTime != nil {
        updates["screen_time"] = *req.ScreenTime
    }
    if req.LineCount != nil {
        updates["line_count"] = *req.LineCount
    }
    if req.Description != nil {
        updates["description"] = *req.Description
    }

    if err := s.db.Model(&ec).Updates(updates).Error; err != nil {
        return nil, err
    }

    s.db.Preload("Character").Preload("Episode").First(&ec, ec.ID)
    return &ec, nil
}

// GetEpisodeCharacters 获取剧集的角色列表
func (s *EpisodeCharacterService) GetEpisodeCharacters(episodeID uint, roleType *string) ([]models.EpisodeCharacter, error) {
    query := s.db.Where("episode_id = ?", episodeID)
    if roleType != nil {
        query = query.Where("role_type = ?", *roleType)
    }
    query = query.Order("sort_order ASC")

    var ecs []models.EpisodeCharacter
    if err := query.Preload("Character").Find(&ecs).Error; err != nil {
        return nil, err
    }
    return ecs, nil
}

// BatchUpdateSortOrder 批量更新出场顺序
func (s *EpisodeCharacterService) BatchUpdateSortOrder(episodeID uint, characterIDs []uint) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        for i, charID := range characterIDs {
            if err := tx.Model(&models.EpisodeCharacter{}).
                Where("episode_id = ? AND character_id = ?", episodeID, charID).
                Update("sort_order", i).Error; err != nil {
                return err
            }
        }
        return nil
    })
}

// DeleteEpisodeCharacter 删除关联（软删除）
func (s *EpisodeCharacterService) DeleteEpisodeCharacter(episodeID, characterID uint) error {
    return s.db.Where("episode_id = ? AND character_id = ?", episodeID, characterID).Delete(&models.EpisodeCharacter{}).Error
}

// Request DTOs
type CreateEpisodeCharacterRequest struct {
    EpisodeID   uint   `json:"episode_id" binding:"required"`
    CharacterID uint   `json:"character_id" binding:"required"`
    RoleType    string `json:"role_type" binding:"required,oneof=main supporting guest extra"`
    SortOrder   int    `json:"sort_order" binding:"min=0"`
}

type UpdateEpisodeCharacterRequest struct {
    RoleType    *string `json:"role_type" binding:"omitempty,oneof=main supporting guest extra"`
    SortOrder   *int    `json:"sort_order" binding:"omitempty,min=0"`
    ScreenTime  *int    `json:"screen_time" binding:"omitempty,min=0"`
    LineCount   *int    `json:"line_count" binding:"omitempty,min=0"`
    Description *string `json:"description"`
}
```

### 3.2 更新 DramaService

```go
// application/services/drama_service.go

// SaveCharacters 更新：使用 EpisodeCharacterService
func (s *DramaService) SaveCharacters(req *SaveCharacterRequest) error {
    // ... 现有角色创建逻辑 ...

    // 建立关联时改为使用 EpisodeCharacterService
    if req.EpisodeID != nil && len(characterIDs) > 0 {
        ecService := NewEpisodeCharacterService(s.db)
        for i, charID := range characterIDs {
            _, err := ecService.CreateEpisodeCharacter(&CreateEpisodeCharacterRequest{
                EpisodeID:   *req.EpisodeID,
                CharacterID: charID,
                RoleType:    models.RoleTypeSupporting, // 默认配角
                SortOrder:   i,
            })
            if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
                return err
            }
        }
    }
    return nil
}

// GetCharacters 更新：支持按角色定位筛选
func (s *DramaService) GetCharacters(dramaID uint, episodeID *uint, roleType *string) ([]models.Character, error) {
    var characters []models.Character

    if episodeID != nil {
        // 通过 EpisodeCharacter 查询
        var ecs []models.EpisodeCharacter
        query := s.db.Where("episode_id = ?", *episodeID)
        if roleType != nil {
            query = query.Where("role_type = ?", *roleType)
        }
        if err := query.Preload("Character").Find(&ecs).Error; err != nil {
            return nil, err
        }

        for _, ec := range ecs {
            characters = append(characters, ec.Character)
        }
    } else {
        // 获取项目所有角色
        if err := s.db.Where("drama_id = ?", dramaID).Find(&characters).Error; err != nil {
            return nil, err
        }
    }

    return characters, nil
}
```

---

## 4. API Handler 改造

### 4.1 EpisodeCharacterHandler (新增)

```go
// api/handlers/episode_character.go
package handlers

import (
    "github.com/gin-gonic/gin"
    "github.com/huobao-drama/application/services"
    "net/http"
)

type EpisodeCharacterHandler struct {
    ecService *services.EpisodeCharacterService
}

func NewEpisodeCharacterHandler(ecService *services.EpisodeCharacterService) *EpisodeCharacterHandler {
    return &EpisodeCharacterHandler{ecService: ecService}
}

// GetEpisodeCharacters 获取剧集角色列表
// GET /api/episodes/:episode_id/characters
func (h *EpisodeCharacterHandler) GetEpisodeCharacters(c *gin.Context) {
    episodeID := c.Param("episode_id")
    roleType := c.Query("role_type")

    ecs, err := h.ecService.GetEpisodeCharacters(episodeID, roleType)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    ecs,
    })
}

// UpdateEpisodeCharacter 更新角色关联
// PUT /api/episodes/:episode_id/characters/:character_id
func (h *EpisodeCharacterHandler) UpdateEpisodeCharacter(c *gin.Context) {
    var req services.UpdateEpisodeCharacterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ec, err := h.ecService.UpdateEpisodeCharacter(id, &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    ec,
    })
}

// BatchUpdateSortOrder 批量更新出场顺序
// PUT /api/episodes/:episode_id/characters/sort
func (h *EpisodeCharacterHandler) BatchUpdateSortOrder(c *gin.Context) {
    var req struct {
        CharacterIDs []uint `json:"character_ids" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := h.ecService.BatchUpdateSortOrder(episodeID, req.CharacterIDs); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteEpisodeCharacter 删除角色关联
// DELETE /api/episodes/:episode_id/characters/:character_id
func (h *EpisodeCharacterHandler) DeleteEpisodeCharacter(c *gin.Context) {
    episodeID := c.Param("episode_id")
    characterID := c.Param("character_id")

    if err := h.ecService.DeleteEpisodeCharacter(episodeID, characterID); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"success": true})
}
```

---

## 5. 路由配置

```go
// api/routes/routes.go

func RegisterEpisodeCharacterRoutes(r *gin.Engine, ecHandler *handlers.EpisodeCharacterHandler) {
    episodes := r.Group("/api/episodes")
    {
        episodes.GET("/:episode_id/characters", ecHandler.GetEpisodeCharacters)
        episodes.PUT("/:episode_id/characters/:character_id", ecHandler.UpdateEpisodeCharacter)
        episodes.PUT("/:episode_id/characters/sort", ecHandler.BatchUpdateSortOrder)
        episodes.DELETE("/:episode_id/characters/:character_id", ecHandler.DeleteEpisodeCharacter)
    }
}
```

---

## 6. 迁移计划

### 6.1 阶段划分

| 阶段 | 任务 | 风险等级 |
|------|------|----------|
| **Phase 1** | 创建 EpisodeCharacter 模型 | 🟢 低 |
| **Phase 2** | 数据库迁移（保留旧数据） | 🟡 中 |
| **Phase 3** | 服务层改造 | 🟡 中 |
| **Phase 4** | API 更新 | 🟡 中 |
| **Phase 5** | 删除旧的 many2many 代码 | 🟢 低 |
| **Phase 6** | 测试验证 | 🔴 高 |

### 6.2 回滚方案

```sql
-- migrations/rollback_episode_characters.sql

-- 1. 备份新表数据
CREATE TABLE episode_characters_new_backup AS SELECT * FROM episode_characters;

-- 2. 恢复旧结构
DROP TABLE episode_characters;

CREATE TABLE episode_characters (
    episode_id BIGINT UNSIGNED NOT NULL,
    character_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (episode_id, character_id),
    FOREIGN KEY (episode_id) REFERENCES episodes(id) ON DELETE CASCADE,
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 3. 恢复数据
INSERT INTO episode_characters (episode_id, character_id)
SELECT episode_id, character_id
FROM episode_characters_new_backup
WHERE deleted_at IS NULL;
```

---

## 7. 测试用例

### 7.1 单元测试

```go
// services/episode_character_service_test.go
package services_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCreateEpisodeCharacter(t *testing.T) {
    // 测试创建关联
    // 测试重复关联
    // 测试软删除恢复
}

func TestBatchUpdateSortOrder(t *testing.T) {
    // 测试批量更新
    // 测试事务回滚
}

func TestGetEpisodeCharactersByRoleType(t *testing.T) {
    // 测试按角色类型筛选
}
```

### 7.2 集成测试

```go
// tests/integration/episode_character_test.go
func TestEpisodeCharacterFlow(t *testing.T) {
    // 1. 创建剧集和角色
    // 2. 建立关联
    // 3. 更新角色定位
    // 4. 软删除
    // 5. 恢复
}
```

---

## 8. 性能影响评估

| 操作 | 旧实现 | 新实现 | 影响 |
|------|--------|--------|------|
| 创建关联 | O(1) | O(1) | 无 |
| 查询剧集角色 | O(n) | O(n log n) 排序 | 轻微 |
| 按角色类型筛选 | 不支持 | O(n) 索引 | 新增能力 |
| 批量更新排序 | 不支持 | O(m) | 新增能力 |
| 存储空间 | 16 bytes/row | ~100 bytes/row | +6倍 |

**建议索引**:
```sql
-- 复合索引优化查询
CREATE INDEX idx_episode_role_sort ON episode_characters(episode_id, role_type, sort_order);
```

---

## 9. 待确认事项

- [ ] 片酬字段是否需要存储（敏感信息）
- [ ] 是否需要多语言配音支持
- [ ] 戏份统计（screen_time）的计算规则
- [ ] 历史关联记录的审计需求
- [ ] 与其他模块的集成影响（如 Storyboard、Scene）

---

## 10. 参考资料

- [GORM Associations](https://gorm.io/docs/associations.html)
- [数据库范式 vs 业务关联表](https://stackoverflow.com/questions/4650900/)
- [Go 项目最佳实践](https://github.com/golang-standards/project-layout)

---

**文档版本**: v1.0
**最后更新**: 2026-02-01
