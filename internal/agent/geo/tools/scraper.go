package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
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

// Info 返回工具信息 (实现 tool.InvokableTool 接口)
func (s *HTTPScraper) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "scrape_webpage",
		Desc: "爬取网页内容，提取标题和主要文本",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"url": {
				Type: "string",
				Desc: "要爬取的网页 URL",
			},
		}),
	}, nil
}

// InvokableRun 执行工具 (实现 tool.InvokableTool 接口)
func (s *HTTPScraper) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return "", fmt.Errorf("解析请求失败: %w", err)
	}

	result, err := s.Scrape(ctx, req.URL)
	if err != nil {
		return "", fmt.Errorf("爬取网页失败: %w", err)
	}

	resp := struct {
		URL   string `json:"url"`
		Title string `json:"title"`
		H1    string `json:"h1"`
	}{
		URL:   result.URL,
		Title: result.Title,
		H1:    result.H1,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("序列化响应失败: %w", err)
	}

	return string(data), nil
}
