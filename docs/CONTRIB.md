# 贡献指南

本文档介绍如何为 Peanut 项目贡献代码。

## 开发环境设置

### 前置要求

- Go 1.24.0+
- SQLite 3（内置，无需安装）
- Docker（可选，用于容器化部署）

### 快速开始

```bash
# 1. 克隆仓库
git clone https://github.com/solariswu/peanut.git
cd peanut

# 2. 安装依赖
go mod download

# 3. 配置环境变量
cp .env.example .env
# 编辑 .env 文件，设置 ARK_API_KEY

# 4. 运行应用
make run
```

## 可用命令

### 构建和运行

| 命令 | 说明 |
|------|------|
| `make build` | 构建应用程序到 `bin/peanut` |
| `make run` | 直接运行应用程序 |
| `make dev` | 开发模式（热重载，需要安装 air） |
| `make clean` | 清理构建文件和数据库 |

### 测试和代码质量

| 命令 | 说明 |
|------|------|
| `make test` | 运行测试（含 race 检测） |
| `make test-coverage` | 生成测试覆盖率报告 |
| `make lint` | 代码检查（golangci-lint） |
| `make fmt` | 格式化代码（gofmt + goimports） |
| `make tidy` | 整理 go.mod 依赖 |

### Docker

| 命令 | 说明 |
|------|------|
| `make docker-build` | 构建 Docker 镜像 |
| `make docker-run` | 运行 Docker 容器 |
| `make docker-up` | 启动所有服务（Docker Compose） |
| `make docker-down` | 停止所有服务 |
| `make docker-logs` | 查看日志 |
| `make docker-ps` | 查看容器状态 |

### 文档

| 命令 | 说明 |
|------|------|
| `make swagger` | 生成 Swagger API 文档 |

## 环境变量

### 服务器配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SERVER_PORT` | HTTP 服务端口 | `8080` |
| `SERVER_MODE` | Gin 模式（debug/release） | `debug` |

### 日志配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `LOG_LEVEL` | 日志级别（debug/info/warn/error） | `debug` |
| `LOG_FORMAT` | 日志格式（console/json） | `console` |

### LLM 配置

| 变量 | 说明 | 必需 |
|------|------|------|
| `ARK_API_KEY` | 豆包 API Key | ✅ |
| `ARK_BASE_URL` | 豆包 API 地址 | ❌ |
| `ARK_MODEL` | 豆包模型名称 | ❌ |

## 项目结构

```
peanut/
├── cmd/
│   └── server/
│       └── main.go           # 应用入口
├── internal/
│   ├── handler/              # HTTP 处理器层
│   ├── service/              # 业务逻辑层
│   ├── repository/           # 数据访问层
│   ├── model/                # 数据模型
│   ├── agent/                # AI 智能体
│   │   └── geo/              # GEO 智能体
│   └── pkg/                  # 内部工具包
├── configs/                  # 配置文件
├── docs/                     # 文档
├── web/                      # Web 前端
├── Makefile                  # 构建脚本
└── docker-compose.yaml       # Docker 编排
```

## 开发流程

### 1. 创建功能分支

```bash
git checkout -b feature/your-feature-name
```

### 2. 编写代码

- 遵循 Go 代码规范
- 添加单元测试
- 更新 Swagger 注释

### 3. 运行测试

```bash
# 运行所有测试
make test

# 生成覆盖率报告
make test-coverage
```

### 4. 代码检查

```bash
# 格式化代码
make fmt

# 运行 linter
make lint
```

### 5. 提交代码

```bash
git add .
git commit -m "feat: 添加新功能描述"
git push origin feature/your-feature-name
```

## 代码规范

### Go 代码规范

- 使用 `gofmt` 格式化代码
- 遵循 [Effective Go](https://golang.org/doc/effective_go)
- 添加适当的注释
- 使用有意义的变量名

### 提交信息规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式：

```
<type>: <subject>

<body>
```

类型：
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建/工具相关

## API 文档

启动服务后访问 Swagger UI：

```
http://localhost:8080/swagger/index.html
```

## 获取帮助

- 提交 Issue：[GitHub Issues](https://github.com/solariswu/peanut/issues)
- 查看文档：[README.md](../README.md)
