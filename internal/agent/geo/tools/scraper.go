package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/solariswu/peanut/internal/agent/geo/models"
)

// WebScraper 网页爬取工具接口
type WebScraper interface {
	Scrape(ctx context.Context, url string) (*models.ScrapedTitle, error)
}

// HTTPScraper HTTP 网页爬取实现（使用 goquery 解析 HTML）
type HTTPScraper struct {
	client *http.Client
}

// NewHTTPScraper 创建新的 HTTP 爬取工具
func NewHTTPScraper() *HTTPScraper {
	return &HTTPScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Scrape 爬取网页标题和 H1
func (s *HTTPScraper) Scrape(ctx context.Context, url string) (*models.ScrapedTitle, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 使用 goquery 解析 HTML
	title, h1 := extractTitleAndH1(string(body))

	return &models.ScrapedTitle{
		Title: title,
		H1:    h1,
		URL:   url,
	}, nil
}

// extractTitleAndH1 从 HTML 中提取标题和 H1
func extractTitleAndH1(htmlContent string) (title, h1 string) {
	// 使用 goquery 解析 HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		// 解析失败，返回默认值
		return "解析失败", "解析失败"
	}

	// 提取 <title> 标签
	if t := doc.Find("title").First(); t.Length() > 0 {
		title = strings.TrimSpace(t.Text())
	} else {
		title = "未找到标题"
	}

	// 提取 <h1> 标签
	if h := doc.Find("h1").First(); h.Length() > 0 {
		h1 = strings.TrimSpace(h.Text())
	} else {
		h1 = "未找到 H1 标签"
	}

	// 清理标题中的换行符和多余空格
	title = cleanText(title)
	h1 = cleanText(h1)

	return title, h1
}

// cleanText 清理文本中的多余空白字符
func cleanText(text string) string {
	// 替换连续的空白字符为单个空格
	lines := strings.Fields(text)
	return strings.Join(lines, " ")
}
