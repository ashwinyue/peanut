package tools

import (
	"context"
	"os"
	"testing"
)

// TestDoubaoSERPProvider_RealAPI 测试真实的豆包 API 调用
//
// 运行此测试前需要设置环境变量：
// export ARK_API_KEY="your-api-key"
// export ARK_BASE_URL="https://ark.cn-beijing.volces.com/api/v3"
// export ARK_MODEL="doubao-pro-256k-240628"
func TestDoubaoSERPProvider_RealAPI(t *testing.T) {
	// 检查是否设置了 API Key
	if os.Getenv("ARK_API_KEY") == "" {
		t.Skip("跳过测试：未设置 ARK_API_KEY 环境变量")
	}

	// 创建 provider
	provider, err := NewDoubaoSERPProvider()
	if err != nil {
		t.Fatalf("创建 DoubaoSERPProvider 失败: %v", err)
	}

	// 测试查询
	testQuery := "GEO 优化是什么"
	ctx := context.Background()

	overview, err := provider.GetAIOverview(ctx, testQuery)
	if err != nil {
		t.Fatalf("GetAIOverview 失败: %v", err)
	}

	// 验证结果
	if overview == nil {
		t.Fatal("overview 不应为 nil")
	}

	if overview.Query != testQuery {
		t.Errorf("查询不匹配: got %s, want %s", overview.Query, testQuery)
	}

	if overview.Summary == "" {
		t.Error("Summary 不应为空")
	}

	if len(overview.Sources) == 0 {
		t.Error("至少应该有一个来源")
	}

	t.Logf("✅ 测试通过")
	t.Logf("查询: %s", overview.Query)
	t.Logf("摘要: %s", overview.Summary)
	t.Logf("片段: %s", overview.Snippet)
	t.Logf("来源: %v", overview.Sources)
}

// TestMockSERPProvider 测试 Mock Provider
func TestMockSERPProvider(t *testing.T) {
	provider := NewMockSERPProvider()
	ctx := context.Background()

	overview, err := provider.GetAIOverview(ctx, "test query")
	if err != nil {
		t.Fatalf("GetAIOverview 失败: %v", err)
	}

	if overview == nil {
		t.Fatal("overview 不应为 nil")
	}

	if overview.Query != "test query" {
		t.Errorf("查询不匹配: got %s, want test query", overview.Query)
	}

	t.Logf("✅ Mock provider 测试通过")
}

// BenchmarkDoubaoSERPProvider 性能测试
func BenchmarkDoubaoSERPProvider(b *testing.B) {
	if os.Getenv("ARK_API_KEY") == "" {
		b.Skip("跳过基准测试：未设置 ARK_API_KEY 环境变量")
	}

	provider, err := NewDoubaoSERPProvider()
	if err != nil {
		b.Fatalf("创建 provider 失败: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = provider.GetAIOverview(ctx, "测试查询")
	}
}

// TestSERPProviderInterface 测试接口实现
func TestSERPProviderInterface(t *testing.T) {
	// 验证 MockSERPProvider 实现了 SERPProvider 接口
	var _ SERPProvider = (*MockSERPProvider)(nil)
	var _ SERPProvider = (*DoubaoSERPProvider)(nil)

	t.Log("✅ 接口实现验证通过")
}

// ExampleDoubaoSERPProvider 使用示例
func ExampleDoubaoSERPProvider() {
	// 设置环境变量（在实际使用中应该从配置文件或环境变量读取）
	// os.Setenv("ARK_API_KEY", "your-api-key")

	provider, err := NewDoubaoSERPProvider()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	overview, err := provider.GetAIOverview(ctx, "什么是 GEO 优化")
	if err != nil {
		panic(err)
	}

	_ = overview
	// 输出:
	// &models.AIOverview{
	//   Query: "什么是 GEO 优化",
	//   Summary: "GEO（Generative Engine Optimization）是 AI 时代的 SEO...",
	//   ...
	// }
}

// TestParseAIOverview 测试响应解析
func TestParseAIOverview(t *testing.T) {
	provider := &DoubaoSERPProvider{}

	tests := []struct {
		name    string
		content string
		query   string
		wantNil bool
	}{
		{
			name:    "空内容",
			content: "",
			query:   "test",
			wantNil: false, // 应该返回默认结构，不应该是 nil
		},
		{
			name:    "纯文本内容",
			content: "这是一段测试内容。",
			query:   "test",
			wantNil: false,
		},
		{
			name:    "JSON 格式",
			content: `{"summary":"测试摘要","key_points":["要点1","要点2"],"sources":["source1","source2"]}`,
			query:   "test",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.parseAIOverview(tt.content, tt.query)

			if tt.wantNil && result != nil {
				t.Errorf("parseAIOverview() 应该返回 nil")
			}
			if !tt.wantNil && result == nil {
				t.Errorf("parseAIOverview() 不应该返回 nil")
			}
			if result != nil && result.Query != tt.query {
				t.Errorf("parseAIOverview() Query = %v, want %v", result.Query, tt.query)
			}
		})
	}
}
