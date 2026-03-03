# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Peanut is a Go-based web application with a React frontend that provides GEO (Generative Engine Optimization) analysis. It uses the Eino framework for AI agent workflows to analyze and optimize web content for AI search engines.

## Build & Development Commands

### Backend (Go)

```bash
# Build the application
make build              # Builds to bin/peanut

# Run the application
make run                # Runs ./cmd/server/main.go
make dev                # Hot reload mode (requires air)

# Testing
make test               # Run all tests with race detection and coverage
make test-coverage      # Generate coverage report (coverage.html)

# Code quality
make lint               # Run golangci-lint
make fmt                # Format code with gofmt and goimports
make tidy               # Tidy go.mod dependencies

# Generate Swagger docs
make swagger            # Generates API docs to api/v1/docs/

# Cleanup
make clean              # Remove build files, coverage reports, and database
```

### Frontend (React + Vite)

```bash
cd web

# Development
npm install             # Install dependencies
npm run dev             # Start dev server (http://localhost:5173)

# Build
npm run build           # Production build
npm run lint            # ESLint check
npm run preview         # Preview production build
```

### Running Single Tests

```bash
# Run a specific test
 go test -v ./internal/agent/geo/parser/... -run TestExtractJSON

# Run tests for a specific package
 go test -v ./internal/agent/geo/...
```

### Docker Commands

```bash
make docker-build       # Build Docker image
make docker-run         # Run Docker container
make docker-up          # docker-compose up -d
make docker-down        # docker-compose down
make docker-logs        # View container logs
```

## Architecture

### Backend Architecture

The backend follows a layered architecture:

```
cmd/server/main.go          # Application entry point, dependency injection
    ↓
internal/handler/           # HTTP handlers (Gin), parameter validation
    ↓
internal/service/           # Business logic, transaction orchestration
    ↓
internal/repository/        # Data access layer (GORM)
    ↓
internal/model/             # Data models with GORM tags
```

**Dependency flow**: Handler → Service → Repository → Model (unidirectional)

### GEO Agent Flow Architecture

The GEO analysis uses the Eino framework with a Graph-based flow:

```
internal/agent/geo/
├── service.go              # Entry point, implements AgentService interface
├── flow/
│   ├── builder.go          # Builds the Eino compose.Graph
│   ├── state.go            # Flow state management
│   └── agents/             # 7 agent implementations:
│       ├── title_scraper.go          # Step 1: Scrape webpage
│       ├── query_researcher.go       # Step 2: Search related queries
│       ├── main_query_extractor.go   # Step 3: Extract main query
│       ├── ai_overview_retriever.go  # Step 4: Get AI overview
│       ├── query_summarizer.go       # Step 5: Summarize queries
│       ├── content_optimizer.go      # Step 6: Generate optimization report
│       └── content_rewriter.go       # Step 7: Rewrite content
├── tools/
│   └── brightdata.go       # Bright Data API integration (SERP + Web Unlocker)
├── llm/
│   └── ark.go              # Doubao/Ark LLM client
└── models/
    ├── flow.go             # FlowState struct for inter-agent communication
    └── response.go         # OptimizationReport output struct
```

The flow uses `compose.Graph` with state passing between agents via `FlowState`.

### Frontend Architecture

```
web/src/
├── components/
│   ├── geo/                # GEO-specific components
│   │   ├── URLInputForm.tsx
│   │   ├── AnalysisResult.tsx
│   │   ├── AnalysisList.tsx
│   │   ├── ScoreCard.tsx
│   │   ├── SuggestionsCard.tsx
│   │   └── ValidationResultCard.tsx
│   ├── layout/             # Layout components
│   └── ui/                 # shadcn/ui components
├── pages/
│   └── GeoAnalysisPage.tsx
├── App.tsx
└── main.tsx
```

**Tech stack**: React 19 + TypeScript + Vite + TailwindCSS + shadcn/ui + Zustand + TanStack Query

## Environment Configuration

### Required Environment Variables

```bash
# Bright Data (for SERP and web scraping)
BRIGHT_DATA_API_KEY=your_api_key
BRIGHT_DATA_ZONE=serp_api
BRIGHT_DATA_WEB_UNLOCKER_ZONE=web_unlocker

# Doubao LLM
ARK_API_KEY=your_ark_api_key
ARK_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
ARK_MODEL=your-model-name
```

### Optional Configuration

Environment variables override `configs/config.yaml`:

```bash
SERVER_PORT=8080
DATABASE_URL=postgresql://...  # Defaults to SQLite (peanut.db)
GIN_MODE=debug  # or release
LOG_LEVEL=debug
```

## Key API Endpoints

- `GET /health` - Health check
- `GET /swagger/*any` - Swagger API documentation
- `POST /api/v1/geo/analysis` - Create GEO analysis task
- `GET /api/v1/geo/analysis/:id` - Get analysis result
- `GET /api/v1/geo/analysis/:id/progress` - SSE stream for progress updates
- `GET /api/v1/geo/analysis/platforms` - List supported platforms

## Database

Default: SQLite (`peanut.db`)

Models:
- `User` - User management
- `GEOAnalysis` - Analysis task storage with status tracking

Auto-migration happens on startup in `cmd/server/main.go`.

## Testing Notes

- Tests use standard Go testing with `testing` package
- Race detection enabled by default (`-race` flag)
- Parser tests in `internal/agent/geo/parser/parser_test.go` show JSON extraction patterns

## Dependencies

**Key Go modules**:
- `github.com/cloudwego/eino` - AI agent framework
- `github.com/gin-gonic/gin` - Web framework
- `gorm.io/gorm` + `gorm.io/driver/sqlite` - ORM
- `github.com/swaggo/gin-swagger` - API documentation
- `go.uber.org/zap` - Logging

**Key npm packages**:
- `@tanstack/react-query` - Data fetching
- `zustand` - State management
- `@radix-ui/*` - UI primitives
- `react-hook-form` + `zod` - Form handling
