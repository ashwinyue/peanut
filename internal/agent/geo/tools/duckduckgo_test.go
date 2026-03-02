package tools

import (
	"context"
	"testing"
)

// TestNewDuckDuckGoSearcher 测试 DuckDuckGo 搜索器创建
func TestNewDuckDuckGoSearcher(t *testing.T) {
	searcher, err := NewDuckDuckGoSearcher()
	if err != nil {
		t.Fatalf("创建 DuckDuckGo 搜索器失败: %v", err)
	}

	if searcher == nil {
		t.Fatal("搜索器不应为 nil")
	}

	t.Log("✅ DuckDuckGo 搜索器创建成功")
}

// TestDuckDuckGoSearcher_Search 测试搜索功能
func TestDuckDuckGoSearcher_Search(t *testing.T) {
	searcher, err := NewDuckDuckGoSearcher()
	if err != nil {
		t.Skipf("跳过测试: 无法创建搜索器 (%v)", err)
	}

	ctx := context.Background()
	query := "Golang 编程教程"

	result, err := searcher.Search(ctx, query)
	if err != nil {
		t.Fatalf("搜索失败: %v", err)
	}

	if result == nil {
		t.Fatal("搜索结果不应为 nil")
	}

	if result.OriginalQuery != query {
		t.Errorf("原始查询不匹配: got %s, want %s", result.OriginalQuery, query)
	}

	if len(result.RelatedQueries) == 0 {
		t.Error("相关查询列表不应为空")
	}

	t.Logf("✅ 搜索测试通过")
	t.Logf("  原始查询: %s", result.OriginalQuery)
	t.Logf("  相关查询数量: %d", len(result.RelatedQueries))
	for i, q := range result.RelatedQueries {
		if i >= 3 {
			break
		}
		t.Logf("    %d. %s", i+1, q)
	}
}

// TestNewDuckDuckGoSERPProvider 测试 DuckDuckGo SERP Provider 创建
func TestNewDuckDuckGoSERPProvider(t *testing.T) {
	provider, err := NewDuckDuckGoSERPProvider()
	if err != nil {
		t.Fatalf("创建 DuckDuckGo SERP Provider 失败: %v", err)
	}

	if provider == nil {
		t.Fatal("Provider 不应为 nil")
	}

	t.Log("✅ DuckDuckGo SERP Provider 创建成功")
}

// TestDuckDuckGoSERPProvider_GetAIOverview 测试 AI Overview 获取
func TestDuckDuckGoSERPProvider_GetAIOverview(t *testing.T) {
	provider, err := NewDuckDuckGoSERPProvider()
	if err != nil {
		t.Skipf("跳过测试: 无法创建 Provider (%v)", err)
	}

	ctx := context.Background()
	query := "什么是 GEO 优化"

	overview, err := provider.GetAIOverview(ctx, query)
	if err != nil {
		t.Fatalf("获取 AI Overview 失败: %v", err)
	}

	if overview == nil {
		t.Fatal("Overview 不应为 nil")
	}

	if overview.Query != query {
		t.Errorf("查询不匹配: got %s, want %s", overview.Query, query)
	}

	if overview.Summary == "" {
		t.Error("摘要不应为空")
	}

	t.Logf("✅ AI Overview 测试通过")
	t.Logf("  查询: %s", overview.Query)
	t.Logf("  摘要长度: %d 字符", len(overview.Summary))
	t.Logf("  来源数量: %d", len(overview.Sources))
	if overview.Snippet != "" {
		t.Logf("  片段: %s", overview.Snippet[:min(100, len(overview.Snippet))])
	}
}

// BenchmarkDuckDuckGoSearcher 性能测试
func BenchmarkDuckDuckGoSearcher(b *testing.B) {
	searcher, err := NewDuckDuckGoSearcher()
	if err != nil {
		b.Skipf("跳过基准测试: 无法创建搜索器 (%v)", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = searcher.Search(ctx, "测试查询")
	}
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
