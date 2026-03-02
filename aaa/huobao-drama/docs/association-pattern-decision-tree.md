# 关联模式选择决策树

## 快速决策

```
┌─────────────────────────────────────────────────────────────┐
│                    关联表是否有业务字段？                     │
└─────────────────────────────────────────────────────────────┘
                            │
            ┌───────────────┴───────────────┐
            │                               │
           YES                             NO
            │                               │
            ▼                               ▼
    ┌───────────────┐             ┌──────────────────┐
    │ 业务关联表 ⭐  │             │ 是否需要复杂查询？ │
    └───────────────┘             └──────────────────┘
                                        │
                        ┌───────────────┴───────────────┐
                        │                               │
                       YES                             NO
                        │                               │
                        ▼                               ▼
                ┌───────────────┐             ┌──────────────┐
                │ 业务关联表     │             │ GORM many2many│
                └───────────────┘             └──────────────┘
```

## 业务字段示例

| 字段类型 | 示例 | 是否需要业务关联表 |
|---------|------|-------------------|
| **定位/分类** | 角色类型（主角/配角）、权限级别 | ✅ 是 |
| **排序** | 显示顺序、优先级、权重 | ✅ 是 |
| **统计** | 出场次数、使用量、点击量 | ✅ 是 |
| **状态** | 审核状态、激活状态、完成度 | ✅ 是 |
| **时间** | 开始时间、结束时间、时长 | ✅ 是 |
| **扩展** | 备注、描述、配置选项 | ✅ 是 |

## 代码对比

### GORM many2many（适合纯关联）

```go
type User struct {
    gorm.Model
    Roles []Role `gorm:"many2many:user_roles;"`
}

type Role struct {
    gorm.Model
    Users []User `gorm:"many2many:user_roles;"`
}

// 适用场景：用户-角色，标签-文章，分类-商品
```

### 业务关联表（适合有业务属性）

```go
type Order struct {
    gorm.Model
    Items []OrderItem `gorm:"foreignKey:OrderID"`
}

type Product struct {
    gorm.Model
}

type OrderItem struct {
    OrderID   uint `gorm:"not null;index"`
    ProductID uint `gorm:"not null;index"`

    // 业务字段
    Quantity    int     `gorm:"not null;default:1"`
    UnitPrice   float64 `gorm:"not null"`
    Discount    float64 `gorm:"default:0"`
    Subtotal    float64 `gorm:"not null"`

    Order   Order   `gorm:"foreignKey:OrderID"`
    Product Product `gorm:"foreignKey:ProductID"`
}

// 适用场景：订单-商品，学生-课程，角色-剧集
```

## 实施建议

### 阶段 1：评估
1. 列出所有关联关系
2. 识别每个关联的业务需求
3. 标记需要扩展的关联

### 阶段 2：设计
1. 为复杂关联创建独立模型
2. 定义业务字段和约束
3. 设计索引和查询模式

### 阶段 3：迁移
1. 数据备份
2. 创建新表结构
3. 迁移现有数据
4. 更新代码逻辑

## 常见模式

### 1. 时间段关联
```go
type Enrollment struct {
    StudentID uint
    CourseID  uint
    EnrolledAt time.Time
    CompletedAt *time.Time
    Grade       *float64
}
```

### 2. 数量关联
```go
type CartItem struct {
    CartID    uint
    ProductID uint
    Quantity  int
    UnitPrice float64
}
```

### 3. 状态关联
```go
type TaskAssignment struct {
    TaskID       uint
    AssigneeID   uint
    Status       string // pending/in_progress/completed
    CompletedAt  *time.Time
}
```

### 4. 排序关联
```go
type PlaylistTrack struct {
    PlaylistID uint
    TrackID    uint
    Position   int // 播放顺序
}
```

## 参考资料

- [EPISODE_CHARACTER_ASSOCIATION.md](./EPISODE_CHARACTER_ASSOCIATION.md) - 完整重构案例
- [GORM Associations](https://gorm.io/docs/associations.html)
- [数据库关联表设计最佳实践](https://stackoverflow.com/questions/4650900/)

---

**文档版本**: v1.0
**最后更新**: 2026-02-01
