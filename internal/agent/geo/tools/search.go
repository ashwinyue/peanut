package tools

import (
	"context"
	"time"

	"github.com/solariswu/peanut/internal/agent/geo/models"
)

// Searcher 搜索工具接口
type Searcher interface {
	Search(ctx context.Context, query string) (*models.QueryFanout, error)
}

// MockSearcher 模拟搜索实现（用于开发测试）
type MockSearcher struct{}

// NewMockSearcher 创建模拟搜索工具
func NewMockSearcher() *MockSearcher {
	return &MockSearcher{}
}

// Search 执行搜索，获取查询发散
func (s *MockSearcher) Search(ctx context.Context, query string) (*models.QueryFanout, error) {
	// 模拟相关查询
	relatedQueries := []string{
		query + " 是什么",
		query + " 怎么做",
		query + " 方法",
		query + " 技巧",
		query + " 注意事项",
		query + " 最佳实践",
	}

	return &models.QueryFanout{
		OriginalQuery:  query,
		RelatedQueries: relatedQueries,
		Timestamp:      time.Now(),
	}, nil
}
