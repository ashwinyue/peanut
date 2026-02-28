# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
# Build
make build              # 构建到 bin/peanut

# Run
make run                # 直接运行
make dev                # 热重载模式（需安装 air）

# Test
make test               # 运行测试（含 race 检测）
make test-coverage      # 生成覆盖率报告

# Lint
make lint               # golangci-lint 检查
make fmt                # 格式化代码（gofmt + goimports）

# Database
make migrate-up         # 执行迁移
make migrate-down       # 回滚迁移

# Docker
docker-compose up -d    # 启动 PostgreSQL + Redis + App
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

**关键技术栈**：
- Web 框架：Gin
- ORM：GORM（PostgreSQL）
- 缓存：go-redis
- 配置：Viper（YAML）
- 日志：Zap
- 密码：bcrypt

## Code Conventions

### 添加新功能模块

1. 在 `internal/model/` 定义模型和 DTO
2. 在 `internal/repository/` 实现数据访问
3. 在 `internal/service/` 实现业务逻辑
4. 在 `internal/handler/` 实现 HTTP 处理器
5. 在 `cmd/server/main.go` 注册依赖

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
