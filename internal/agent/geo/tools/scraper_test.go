package tools

import (
	"context"
	"testing"
)

// TestHTTPScraper_Scrape 测试 HTTP 爬取功能
func TestHTTPScraper_Scrape(t *testing.T) {
	scraper := NewHTTPScraper()

	tests := []struct {
		name           string
		url            string
		expectTitle    bool
		expectH1       bool
		expectError    bool
		skipIfNoInternet bool
	}{
		{
			name:           "爬取百度首页",
			url:            "https://www.baidu.com",
			expectTitle:    true,
			expectH1:       true,
			expectError:    false,
			skipIfNoInternet: true,
		},
		{
			name:           "爬取 GitHub Trending",
			url:            "https://github.com/trending",
			expectTitle:    true,
			expectH1:       true,
			expectError:    false,
			skipIfNoInternet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if testing.Short() && tt.skipIfNoInternet {
				t.Skip("跳过需要网络的测试")
			}

			result, err := scraper.Scrape(context.Background(), tt.url)

			if tt.expectError {
				if err == nil {
					t.Errorf("期望返回错误，但没有返回错误")
				}
				return
			}

			if err != nil {
				t.Errorf("爬取失败: %v", err)
				return
			}

			t.Logf("✅ 爬取成功")
			t.Logf("  URL: %s", result.URL)
			t.Logf("  Title: %s", result.Title)
			t.Logf("  H1: %s", result.H1)

			// 验证标题
			if tt.expectTitle && result.Title == "" {
				t.Errorf("期望提取到标题，但为空")
			} else if tt.expectTitle && result.Title == "未找到标题" {
				t.Errorf("标题提取失败: %s", result.Title)
			}

			// 验证 H1
			if tt.expectH1 && result.H1 == "" {
				t.Errorf("期望提取到 H1，但为空")
			} else if tt.expectH1 && result.H1 == "未找到 H1 标签" {
				t.Errorf("H1 提取失败: %s", result.H1)
			}
		})
	}
}

// TestCleanText 测试文本清理功能
func TestCleanText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "清理换行符",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "Line 1 Line 2 Line 3",
		},
		{
			name:     "清理多余空格",
			input:    "  Multiple    Spaces   ",
			expected: "Multiple Spaces",
		},
		{
			name:     "清理制表符",
			input:    "Tab\tSeparated\tValues",
			expected: "Tab Separated Values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanText(tt.input)
			if result != tt.expected {
				t.Errorf("cleanText() = %q, want %q", result, tt.expected)
			}
		})
	}
}
