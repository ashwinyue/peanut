.PHONY: all build run test clean lint fmt swagger help

# 变量
APP_NAME := peanut
MAIN_PATH := ./cmd/server
BUILD_DIR := ./bin
GO := go
GOFLAGS := -v

# 默认目标
all: build

# 构建
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

# 运行
run:
	@echo "Running $(APP_NAME)..."
	$(GO) run $(MAIN_PATH)

# 开发模式（热重载需要安装 air）
dev:
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	air

# 测试
test:
	@echo "Running tests..."
	$(GO) test -v -race -cover ./...

# 测试覆盖率
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# 代码检查
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

# 格式化
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@which goimports > /dev/null || go install golang.org/x/tools/cmd/goimports@latest
	goimports -w .

# 整理依赖
tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy

# 清理
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Swagger 文档生成
swagger:
	@echo "Generating Swagger docs..."
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g cmd/server/main.go -o api/v1/docs --parseDependency --parseInternal

# 数据库迁移（向上）
migrate-up:
	@echo "Running migrations..."
	@which migrate > /dev/null || go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	migrate -path ./scripts/migrations -database "postgres://postgres:postgres@localhost:5432/peanut?sslmode=disable" up

# 数据库迁移（向下）
migrate-down:
	@echo "Rolling back migrations..."
	migrate -path ./scripts/migrations -database "postgres://postgres:postgres@localhost:5432/peanut?sslmode=disable" down

# Docker 构建
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

# Docker 运行
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(APP_NAME):latest

# 帮助
help:
	@echo "Available targets:"
	@echo "  build          - 构建应用程序"
	@echo "  run            - 运行应用程序"
	@echo "  dev            - 开发模式（热重载）"
	@echo "  test           - 运行测试"
	@echo "  test-coverage  - 运行测试并生成覆盖率报告"
	@echo "  lint           - 代码检查"
	@echo "  fmt            - 格式化代码"
	@echo "  tidy           - 整理依赖"
	@echo "  swagger        - 生成 Swagger API 文档"
	@echo "  clean          - 清理构建文件"
	@echo "  migrate-up     - 运行数据库迁移"
	@echo "  migrate-down   - 回滚数据库迁移"
	@echo "  docker-build   - 构建 Docker 镜像"
	@echo "  docker-run     - 运行 Docker 容器"
