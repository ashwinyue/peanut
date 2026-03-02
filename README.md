# Peanut

基于 Go + Gin + SQLite 构建的现代化 Web 应用脚手架，集成豆包元宝 GEO（生成式引擎优化）智能体。

## 🎯 项目简介

Peanut 是一个功能完备的 Web 应用框架，提供：

- **RESTful API**：基于 Gin 框架的高性能 HTTP 服务
- **数据持久化**：GORM ORM + SQLite 数据库（轻量级，零配置）
- **GEO 智能体**：针对豆包元宝（字节跳动生成式搜索引擎）的内容优化分析
- **数据库存储**：分析结果持久化存储，支持历史查询

## 📋 目录

- [技术栈](#技术栈)
- [快速开始](#快速开始)
- [项目结构](#项目结构)
- [GEO 智能体](#geo-智能体)
- [开发指南](#开发指南)
- [API 文档](#api-文档)

## 🛠 技术栈

| 组件 | 技术 | 说明 |
|------|------|------|
| Web 框架 | [Gin](https://github.com/gin-gonic/gin) | 高性能 HTTP 框架 |
| ORM | [GORM](https://gorm.io) | Go 语言 ORM 库 |
| 数据库 | SQLite 3 | 嵌入式数据库（零配置） |
| 配置管理 | [Viper](https://github.com/spf13/viper) | 配置文件解析 |
| 日志 | [Zap](https://github.com/uber-go/zap) | 结构化日志 |
| LLM | 豆包（字节跳动） | 大语言模型 |
| AI 框架 | [Eino](https://github.com/cloudwego/eino) | 字节跳动 AI 框架 |

## 🚀 快速开始

### 环境要求

- Go 1.24.0+
- SQLite 3（内置，无需额外安装）

### 安装

```bash
# 克隆仓库
git clone https://github.com/solariswu/peanut.git
cd peanut

# 安装依赖
go mod download

# 编译
make build

# 运行（自动创建 peanut.db 数据库文件）
make run
```

### 配置

创建 `configs/config.yaml` 文件：

```yaml
server:
  port: 8080
  mode: debug

log:
  level: info
  format: console

ark:
  api_key: "your-api-key"  # 豆包 API Key
  base_url: "https://ark.cn-beijing.volces.com/api/v3"
  model: "doubao-pro-256k-240628"
```

或使用环境变量：

```bash
export ARK_API_KEY="your-api-key"
./bin/peanut
```

### Docker 部署

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

**注意**：使用 Docker 前需要先配置 `ARK_API_KEY` 环境变量。

## 📁 项目结构

## 📁 项目结构

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
│   │       ├── agents/       # Agent 实现
│   │       ├── tools/        # 工具（搜索、爬取）
│   │       ├── llm/          # LLM 客户端
│   │       ├── models/       # 数据模型
│   │       └── parser/       # 报告解析
│   └── pkg/                  # 内部包
│       ├── database/         # 数据库连接
│       ├── cache/            # 缓存连接
│       └── response/         # 统一响应
├── configs/                  # 配置文件
│   └── config.yaml
├── migrations/               # 数据库迁移
├── web/                      # Web 前端
├── docs/                     # 设计文档
│   └── GEO_AGENT_DESIGN.md   # GEO 智能体设计文档
├── Makefile                  # 构建脚本
├── docker-compose.yaml       # Docker 编排
└── README.md
```

## 🤖 GEO 智能体

### 什么是 GEO？

**GEO**（Generative Engine Optimization，生成式引擎优化）是 SEO 在 AI 时代的进化版。

| 维度 | SEO（传统） | GEO（AI 时代） |
|------|-----------|---------------|
| 目标 | 搜索结果排名 | 被 AI 引用/推荐 |
| 优化对象 | 关键词、链接 | 内容质量、权威性 |
| 搜索结果 | 10 个蓝色链接 | AI 生成摘要 + 引用 |

### 豆包元宝优化

本项目的 GEO 智能体专门针对**豆包元宝**（字节跳动的生成式搜索引擎）进行优化：

#### 核心评分要素

- **权威性（40%）**：官方来源、政府机构、权威媒体
- **时效性（25%）**：最新数据、时间标注
- **结构化（20%）**：列表、表格清晰呈现
- **中文质量（15%）**：流畅表达、符合中文习惯

#### 工作流程

1. **爬取网页**：提取标题和主要内容
2. **查询发散**：基于国内搜索引擎发现相关查询
3. **主查询提取**：识别核心搜索意图
4. **AI 摘要生成**：模拟豆包元宝 AI 摘要
5. **查询总结**：提炼关键主题
6. **优化报告**：生成可操作的优化建议

### 使用示例

```bash
# 分析 URL 的豆包元宝优化潜力
curl -X POST http://localhost:8080/api/v1/geo/analyze/stream \
  -H "Content-Type: application/json" \
  -d '{"url": "https://your-site.com"}'
```

详细设计文档请参阅 [docs/GEO_AGENT_DESIGN.md](docs/GEO_AGENT_DESIGN.md)

## 💻 开发指南

### 可用命令

```bash
# 构建
make build              # 构建到 bin/peanut

# 依赖管理
make tidy               # 整理 go.mod 依赖

# 运行
make run                # 直接运行
make dev                # 热重载模式（需安装 air）

# 测试
make test               # 运行测试（含 race 检测）
make test-coverage      # 生成覆盖率报告

# 代码质量
make lint               # golangci-lint 检查
make fmt                # 格式化代码

# 数据库
make migrate-up         # 执行迁移
make migrate-down       # 回滚迁移

# Docker
make docker-build       # 构建 Docker 镜像
make docker-run         # 运行 Docker 容器
```

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `ARK_API_KEY` | 豆包 LLM API Key | - |
| `ARK_BASE_URL` | 豆包 API 地址 | `https://ark.cn-beijing.volces.com/api/v3` |
| `ARK_MODEL` | 豆包模型名称 | `doubao-pro-256k-240628` |

## 📚 API 文档

### 健康检查

```bash
GET /health
```

### GEO 分析

```bash
# 同步分析
POST /api/v1/geo/analyze
Content-Type: application/json

{
  "url": "https://example.com"
}

# 流式分析（推荐）
POST /api/v1/geo/analyze/stream
Content-Type: application/json

{
  "url": "https://example.com"
}
```

启动服务后访问 `http://localhost:8080/swagger/` 查看完整 API 文档。

### 用户管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/users` | 用户列表 |
| POST | `/api/v1/users` | 创建用户 |
| GET | `/api/v1/users/:id` | 获取用户 |
| PUT | `/api/v1/users/:id` | 更新用户 |
| DELETE | `/api/v1/users/:id` | 删除用户 |

## 📄 License

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📞 联系方式

- 作者：solariswu
- 项目地址：[https://github.com/solariswu/peanut](https://github.com/solariswu/peanut)
