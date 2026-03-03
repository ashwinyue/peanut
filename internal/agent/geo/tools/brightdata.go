package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/solariswu/peanut/internal/agent/geo/models"
)

// BrightDataConfig Bright Data 配置
type BrightDataConfig struct {
	APIKey   string
	Zone     string
	Endpoint string
}

// LoadBrightDataConfig 从环境变量加载 Bright Data 配置
func LoadBrightDataConfig() (*BrightDataConfig, error) {
	apiKey := os.Getenv("BRIGHT_DATA_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("未设置 BRIGHT_DATA_API_KEY 环境变量")
	}

	zone := os.Getenv("BRIGHT_DATA_ZONE")
	if zone == "" {
		zone = "serp_api" // 默认 zone 名称
	}

	endpoint := os.Getenv("BRIGHT_DATA_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://api.brightdata.com/request"
	}

	return &BrightDataConfig{
		APIKey:   apiKey,
		Zone:     zone,
		Endpoint: endpoint,
	}, nil
}

// LoadWebUnlockerConfig 加载 Web Unlocker 配置
func LoadWebUnlockerConfig() (*BrightDataConfig, error) {
	apiKey := os.Getenv("BRIGHT_DATA_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("未设置 BRIGHT_DATA_API_KEY 环境变量")
	}

	zone := os.Getenv("BRIGHT_DATA_WEB_UNLOCKER_ZONE")
	if zone == "" {
		zone = "web_unlocker" // 默认 web unlocker zone
	}

	endpoint := os.Getenv("BRIGHT_DATA_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://api.brightdata.com/request"
	}

	return &BrightDataConfig{
		APIKey:   apiKey,
		Zone:     zone,
		Endpoint: endpoint,
	}, nil
}

// BrightDataSearcher 使用 Bright Data SERP API 的搜索实现
type BrightDataSearcher struct {
	config     *BrightDataConfig
	httpClient *http.Client
}

// NewBrightDataSearcher 创建 Bright Data 搜索器
func NewBrightDataSearcher() (*BrightDataSearcher, error) {
	config, err := LoadBrightDataConfig()
	if err != nil {
		return nil, err
	}

	return &BrightDataSearcher{
		config: config,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// Search 执行搜索，获取查询发散
func (s *BrightDataSearcher) Search(ctx context.Context, query string) (*models.QueryFanout, error) {
	if query == "" {
		return nil, fmt.Errorf("搜索查询不能为空")
	}

	// 构建请求体（匹配官方 API 格式）
	requestBody := map[string]any{
		"zone":   s.config.Zone,
		"url":    fmt.Sprintf("https://www.google.com/search?q=%s", url.QueryEscape(query)),
		"format": "raw",
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("构建请求体失败: %w", err)
	}

	// 创建 POST 请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.config.Endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Bright Data API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bright Data API 返回错误: %s - %s", resp.Status, string(body))
	}

	// 解析响应
	var serpResp BrightDataSERPResponse
	if err := json.Unmarshal(body, &serpResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 转换为 QueryFanout
	relatedQueries := make([]string, 0)

	// 从相关搜索中提取查询
	for _, related := range serpResp.RelatedSearches {
		if related.Query != "" {
			relatedQueries = append(relatedQueries, related.Query)
		}
	}

	// 如果相关搜索不够，从搜索结果标题中生成
	if len(relatedQueries) < 5 {
		for _, result := range serpResp.OrganicResults {
			if result.Title != "" {
				relatedQueries = append(relatedQueries, result.Title)
			}
			if len(relatedQueries) >= 10 {
				break
			}
		}
	}

	return &models.QueryFanout{
		OriginalQuery:  query,
		RelatedQueries: relatedQueries,
		Timestamp:      time.Now(),
	}, nil
}

// BrightDataSERPResponse Bright Data SERP API 响应结构
type BrightDataSERPResponse struct {
	SearchMetadata struct {
		ID          string `json:"id"`
		Status      string `json:"status"`
		CreatedAt   string `json:"created_at"`
		ProcessedAt string `json:"processed_at"`
	} `json:"search_metadata"`

	SearchParameters struct {
		Q  string `json:"q"`
		Gl string `json:"gl"`
		Hl string `json:"hl"`
	} `json:"search_parameters"`

	OrganicResults []struct {
		Title        string `json:"title"`
		URL          string `json:"link"`
		Snippet      string `json:"snippet"`
		DisplayURL   string `json:"displayed_link"`
	} `json:"organic_results"`

	RelatedSearches []struct {
		Query string `json:"query"`
		Link  string `json:"link"`
	} `json:"related_searches"`

	AIOverview *struct {
		Title    string   `json:"title"`
		Snippet  string   `json:"snippet"`
		Sources  []string `json:"sources"`
	} `json:"ai_overview,omitempty"`
}

// BrightDataSERPProvider 使用 Bright Data 作为 SERP 提供者
type BrightDataSERPProvider struct {
	config     *BrightDataConfig
	httpClient *http.Client
}

// NewBrightDataSERPProvider 创建 Bright Data SERP 提供者
func NewBrightDataSERPProvider() (*BrightDataSERPProvider, error) {
	config, err := LoadBrightDataConfig()
	if err != nil {
		return nil, err
	}

	return &BrightDataSERPProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// GetAIOverview 获取 AI Overview
func (p *BrightDataSERPProvider) GetAIOverview(ctx context.Context, query string) (*models.AIOverview, error) {
	if query == "" {
		return nil, fmt.Errorf("搜索查询不能为空")
	}

	// 构建请求体（匹配官方 API 格式）
	requestBody := map[string]any{
		"zone":             p.config.Zone,
		"url":              fmt.Sprintf("https://www.google.com/search?q=%s", url.QueryEscape(query)),
		"format":           "raw",
		"brd_ai_overview":  "2", // 启用 AI Overview
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("构建请求体失败: %w", err)
	}

	// 创建 POST 请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.Endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Bright Data API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bright Data API 返回错误: %s - %s", resp.Status, string(body))
	}

	// 解析响应
	var serpResp BrightDataSERPResponse
	if err := json.Unmarshal(body, &serpResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 提取 AI Overview 或生成摘要
	aiOverview := &models.AIOverview{
		Query:   query,
		Sources: make([]string, 0),
	}

	// 如果响应中包含 AI Overview，使用它
	if serpResp.AIOverview != nil {
		aiOverview.Summary = serpResp.AIOverview.Title + "\n\n" + serpResp.AIOverview.Snippet
		aiOverview.Sources = serpResp.AIOverview.Sources
		if len(aiOverview.Sources) == 0 {
			// 从有机结果中提取来源
			for _, result := range serpResp.OrganicResults {
				if result.URL != "" {
					aiOverview.Sources = append(aiOverview.Sources, result.URL)
				}
			}
		}
	} else {
		// 生成基于搜索结果的摘要
		aiOverview.Summary = generateSummaryFromResults(query, serpResp.OrganicResults)

		// 提取来源
		for _, result := range serpResp.OrganicResults {
			if result.URL != "" {
				aiOverview.Sources = append(aiOverview.Sources, result.URL)
			}
		}
	}

	// 提取片段（使用第一个结果的摘要）
	if len(serpResp.OrganicResults) > 0 && serpResp.OrganicResults[0].Snippet != "" {
		aiOverview.Snippet = serpResp.OrganicResults[0].Snippet
	} else {
		aiOverview.Snippet = fmt.Sprintf("关于「%s」的搜索结果", query)
	}

	return aiOverview, nil
}

// generateSummaryFromResults 从搜索结果生成摘要
func generateSummaryFromResults(query string, results []struct {
	Title      string `json:"title"`
	URL        string `json:"link"`
	Snippet    string `json:"snippet"`
	DisplayURL string `json:"displayed_link"`
}) string {
	if len(results) == 0 {
		return fmt.Sprintf("关于「%s」未找到相关搜索结果。", query)
	}

	summary := fmt.Sprintf("关于「%s」的搜索结果摘要：\n\n", query)

	for i, result := range results {
		if i >= 5 { // 最多取前5个结果
			break
		}
		summary += fmt.Sprintf("%d. %s\n   %s\n\n", i+1, result.Title, result.Snippet)
	}

	summary += fmt.Sprintf("以上是基于「%s」的搜索结果。", query)

	return summary
}

// Info 返回搜索工具信息 (实现 tool.InvokableTool 接口)
func (s *BrightDataSearcher) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "search_queries",
		Desc: "基于查询词搜索相关内容",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type: "string",
				Desc: "搜索查询词",
			},
		}),
	}, nil
}

// InvokableRun 执行搜索工具 (实现 tool.InvokableTool 接口)
func (s *BrightDataSearcher) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var req struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return "", fmt.Errorf("解析请求失败: %w", err)
	}

	result, err := s.Search(ctx, req.Query)
	if err != nil {
		return "", fmt.Errorf("搜索失败: %w", err)
	}

	resp := struct {
		OriginalQuery  string   `json:"original_query"`
		RelatedQueries []string `json:"related_queries"`
	}{
		OriginalQuery:  result.OriginalQuery,
		RelatedQueries: result.RelatedQueries,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("序列化响应失败: %w", err)
	}

	return string(data), nil
}

// Info 返回 AI Overview 工具信息 (实现 tool.InvokableTool 接口)
func (p *BrightDataSERPProvider) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_ai_overview",
		Desc: "获取生成式搜索引擎的 AI 摘要内容",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type: "string",
				Desc: "要查询的内容",
			},
		}),
	}, nil
}

// InvokableRun 执行工具 (实现 tool.InvokableTool 接口)
func (p *BrightDataSERPProvider) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var req struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return "", fmt.Errorf("解析请求失败: %w", err)
	}

	result, err := p.GetAIOverview(ctx, req.Query)
	if err != nil {
		return "", fmt.Errorf("获取 AI Overview 失败: %w", err)
	}

	resp := struct {
		Query   string   `json:"query"`
		Summary string   `json:"summary"`
		Sources []string `json:"sources"`
	}{
		Query:   result.Query,
		Summary: result.Summary,
		Sources: result.Sources,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("序列化响应失败: %w", err)
	}

	return string(data), nil
}

// BrightDataWebScraper 使用 Bright Data Web Unlocker 的网页爬取实现
type BrightDataWebScraper struct {
	config     *BrightDataConfig
	httpClient *http.Client
}

// NewBrightDataWebScraper 创建 Bright Data Web Scraper
func NewBrightDataWebScraper() (*BrightDataWebScraper, error) {
	config, err := LoadWebUnlockerConfig()
	if err != nil {
		return nil, err
	}

	return &BrightDataWebScraper{
		config: config,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// Scrape 爬取网页内容
func (s *BrightDataWebScraper) Scrape(ctx context.Context, targetURL string) (*models.ScrapedTitle, error) {
	if targetURL == "" {
		return nil, fmt.Errorf("URL 不能为空")
	}

	// 构建请求体
	requestBody := map[string]any{
		"zone":   s.config.Zone,
		"url":    targetURL,
		"format": "markdown",
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("构建请求体失败: %w", err)
	}

	// 创建 POST 请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.config.Endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Bright Data API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bright Data API 返回错误: %s - %s", resp.Status, string(body))
	}

	// 解析响应（Web Unlocker 返回的是网页内容）
	content := string(body)

	// 从内容中提取标题和 H1
	title, h1 := extractTitleAndH1FromMarkdown(content)

	return &models.ScrapedTitle{
		Title:   title,
		H1:      h1,
		URL:     targetURL,
		Content: content,
	}, nil
}

// extractTitleAndH1FromMarkdown 从 Markdown 内容中提取标题和 H1
func extractTitleAndH1FromMarkdown(content string) (title, h1 string) {
	// 尝试从 markdown 中提取 H1
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 查找 # 开头的 H1
		if strings.HasPrefix(line, "# ") && h1 == "" {
			h1 = strings.TrimPrefix(line, "# ")
			h1 = strings.TrimSpace(h1)
		}
	}

	// 如果找到 H1，也用它作为 title
	if h1 != "" {
		title = h1
	} else {
		title = "未找到标题"
	}

	return title, h1
}

// Info 返回工具信息 (实现 tool.InvokableTool 接口)
func (s *BrightDataWebScraper) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "scrape_webpage",
		Desc: "爬取网页内容，提取标题和主要文本（使用 Bright Data Web Unlocker）",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"url": {
				Type: "string",
				Desc: "要爬取的网页 URL",
			},
		}),
	}, nil
}

// InvokableRun 执行工具 (实现 tool.InvokableTool 接口)
func (s *BrightDataWebScraper) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
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
		URL     string `json:"url"`
		Title   string `json:"title"`
		H1      string `json:"h1"`
		Content string `json:"content"`
	}{
		URL:     result.URL,
		Title:   result.Title,
		H1:      result.H1,
		Content: result.Content,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("序列化响应失败: %w", err)
	}

	return string(data), nil
}

