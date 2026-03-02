# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Requirements

- Go 1.25.0+
- SQLite 3+ (默认数据库)
- Node.js 18+ (前端开发，可选)

## Environment Setup

```bash
cp .env.example .env    # 复制环境变量模板
# 编辑 .env 配置数据库和 LLM 连接信息
```

环境变量优先级高于 `configs/config.yaml`（Viper AutomaticEnv）

**关键环境变量：**
- `ARK_API_KEY` - 豆包 LLM API Key（GEO 分析功能必需）
- `ARK_BASE_URL` - 豆包 API 地址
- `ARK_MODEL` - 使用的模型名称

## Build & Development Commands

```bash
# Build
make build              # 构建到 bin/peanut

# Dependencies
make tidy               # 整理 go.mod 依赖

# Run
make run                # 直接运行
make dev                # 热重载模式（需安装 air）

# Test
make test               # 运行测试（含 race 检测）
make test-coverage      # 生成覆盖率报告

# Lint
make lint               # golangci-lint 检查
make fmt                # 格式化代码（gofmt + goimports）

# Swagger API 文档
make swagger            # 生成 Swagger 文档到 api/v1/docs/

# Docker
make docker-build       # 构建 Docker 镜像
make docker-run         # 运行 Docker 容器
make docker-up          # docker-compose up -d (启动所有服务)
make docker-down        # docker-compose down (停止所有服务)
make docker-logs        # 查看容器日志
```

## Architecture Overview

```
cmd/server/main.go          # 应用入口，依赖注入
    ↓
internal/handler/           # HTTP 处理器层，参数校验，调用 service
    ↓
internal/service/           # 业务逻辑层，事务编排
    ↓
internal/repository/        # 数据访问层，GORM 操作
    ↓
internal/model/             # 数据模型，含 GORM tag 和请求/响应 DTO
```

**依赖流向**：Handler → Service → Repository → Model（单向）

**API 路由结构**：
- `GET /health` - 健康检查
- `GET /swagger/*any` - Swagger API 文档
- `GET /api/v1/users` - 用户列表
- `POST /api/v1/users` - 创建用户
- `GET /api/v1/users/:id` - 获取用户
- `PUT /api/v1/users/:id` - 更新用户
- `DELETE /api/v1/users/:id` - 删除用户
- `POST /api/v1/geo/analyze` - GEO 内容分析
- `GET /api/v1/geo/analyze/:id` - 获取分析结果

**internal/pkg/ 目录**：
- `database/` - SQLite 连接封装
- `response/` - 统一响应工具

**internal/agent/ 目录**：
- `geo/` - GEO 分析 Agent，集成豆包 LLM 进行内容优化

**关键技术栈**：
- Web 框架：Gin
- ORM：GORM（SQLite / PostgreSQL 兼容）
- 配置：Viper（YAML）
- 日志：Zap
- AI/LLM：豆包（火山引擎 ARK API）
- API 文档：Swagger (swag)

## Web Frontend (web/)

React 前端项目，用于 GEO 分析结果展示：

```bash
cd web
npm install             # 安装依赖
npm run dev             # 开发服务器 (http://localhost:5173)
npm run build           # 生产构建
npm run lint            # ESLint 检查
```

**技术栈**：
- React 19 + TypeScript
- Vite 构建工具
- TailwindCSS + shadcn/ui 组件
- TanStack Query (React Query)
- Zustand 状态管理
- React Hook Form + Zod 表单验证

## Code Conventions

### 添加新功能模块

1. 在 `internal/model/` 定义模型和 DTO
2. 在 `internal/repository/` 实现数据访问
3. 在 `internal/service/` 实现业务逻辑
4. 在 `internal/handler/` 实现 HTTP 处理器
5. 在 `cmd/server/main.go` 注册依赖和路由

### 统一响应格式

```go
// 成功
response.Success(c, data)
response.SuccessPage(c, list, total, page, pageSize)

// 错误
response.BadRequest(c, "错误信息")
response.NotFound(c, "资源不存在")
response.ServerError(c, "服务器错误")
```

### GORM 模型规范

```go
type User struct {
    BaseModel  // 嵌入 ID, CreatedAt, UpdatedAt
    Username string `json:"username" gorm:"type:varchar(32);uniqueIndex;not null"`
}
```

### Service 错误定义

```go
var (
    ErrUserNotFound = errors.New("用户不存在")
)
```

## Configuration

配置文件：`configs/config.yaml`

环境变量可覆盖配置（Viper AutomaticEnv）

启动时指定配置：`./peanut -config /path/to/config.yaml`

## Database

默认使用 SQLite（`peanut.db`），无需额外配置。

如需切换到 PostgreSQL，修改 `configs/config.yaml` 中的数据库连接配置。
