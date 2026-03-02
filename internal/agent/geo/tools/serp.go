package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/solariswu/peanut/internal/agent/geo/models"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/responses"
)

// SERPProvider 搜索引擎结果页面提供者接口
type SERPProvider interface {
	GetAIOverview(ctx context.Context, query string) (*models.AIOverview, error)
}

// MockSERPProvider 模拟 SERP 提供者（用于开发测试）
type MockSERPProvider struct{}

// NewMockSERPProvider 创建模拟 SERP 提供者
func NewMockSERPProvider() *MockSERPProvider {
	return &MockSERPProvider{}
}

// GetAIOverview 获取 AI Overview（模拟实现）
func (p *MockSERPProvider) GetAIOverview(ctx context.Context, query string) (*models.AIOverview, error) {
	return &models.AIOverview{
		Query:   query,
		Summary: "这是一个 AI 生成的概述，用于演示目的。实际应用中应该调用真实的 SERP API。",
		Sources: []string{
			"https://example.com/source1",
			"https://example.com/source2",
			"https://example.com/source3",
		},
		Snippet: "AI 摘要片段...",
	}, nil
}

// DoubaoSERPProvider 使用豆包模型生成 AI Overview 的实现
type DoubaoSERPProvider struct {
	client      *arkruntime.Client
	modelName   string
	fallbackMsg string
}

// NewDoubaoSERPProvider 创建豆包 SERP 提供者
func NewDoubaoSERPProvider() (*DoubaoSERPProvider, error) {
	apiKey := os.Getenv("ARK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("未设置 ARK_API_KEY 环境变量")
	}

	baseURL := os.Getenv("ARK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}

	modelName := os.Getenv("ARK_MODEL")
	if modelName == "" {
		// 使用豆包专业版模型
		modelName = "doubao-pro-256k-240628"
	}

	client := arkruntime.NewClientWithApiKey(
		apiKey,
		arkruntime.WithBaseUrl(baseURL),
	)

	return &DoubaoSERPProvider{
		client:    client,
		modelName: modelName,
		fallbackMsg: "【豆包 AI 生成】\n\n基于您的问题，我生成了一份优化建议报告。",
	}, nil
}

// GetAIOverview 获取 AI Overview（使用豆包模型生成）
func (p *DoubaoSERPProvider) GetAIOverview(ctx context.Context, query string) (*models.AIOverview, error) {
	// 构建提示词，让豆包模拟豆包元宝的 AI 摘要风格
	prompt := fmt.Sprintf(`你是豆包元宝（字节跳动的生成式搜索引擎）的 AI 摘要生成器。请针对以下查询生成一份符合豆包元宝风格的 AI 摘要。

查询：%s

豆包元宝 AI 摘要特点：
1. 权威性优先：优先引用官方来源、政府机构、知名学术期刊、权威媒体
2. 时效性标注：明确标注信息发布时间
3. 结构化呈现：使用列表、表格等形式清晰组织信息
4. 中文友好：语言表达符合中文阅读习惯

请按以下格式返回（使用 JSON）：
{
  "summary": "用3-5句话总结核心信息，直接回答用户问题",
  "key_points": [
    "关键点1（可标注来源，如：根据XX权威机构）",
    "关键点2",
    "关键点3"
  ],
  "sources": [
    "官方来源1（如：XX官方网站）",
    "权威媒体1（如：人民日报、新华网）",
    "学术机构1（如：中国科学院XX研究所）"
  ]
}

要求：
1. 摘要要简洁、准确、客观，直接回答用户问题
2. 关键点要涵盖最重要的信息，按重要性排序
3. 来源优先选择国内权威网站（官方网站、gov.cn、权威媒体、学术机构）
4. 如果涉及时间敏感信息，必须标注时间
5. 使用清晰的中文表达，避免翻译腔`, query)

	// 创建输入消息
	inputMessage := &responses.ItemInputMessage{
		Role: responses.MessageRole_user,
		Content: []*responses.ContentItem{
			{
				Union: &responses.ContentItem_Text{
					Text: &responses.ContentItemText{
						Type: responses.ContentItemType_input_text,
						Text: prompt,
					},
				},
			},
		},
	}

	// 调用豆包 API
	resp, err := p.client.CreateResponses(ctx, &responses.ResponsesRequest{
		Model: p.modelName,
		Input: &responses.ResponsesInput{
			Union: &responses.ResponsesInput_ListValue{
				ListValue: &responses.InputItemList{ListValue: []*responses.InputItem{{
					Union: &responses.InputItem_InputMessage{
						InputMessage: inputMessage,
					},
				}}},
			},
		},
	})

	if err != nil {
		// 如果 API 调用失败，返回基础信息
		return &models.AIOverview{
			Query:   query,
			Summary: p.fallbackMsg,
			Sources: []string{
				"https://www.baidu.com/s?wd=" + query,
				"https://www.bing.com/search?q=" + query,
			},
			Snippet: "API 调用失败，使用备用响应",
		}, nil
	}

	// 解析响应
	if resp == nil {
		return &models.AIOverview{
			Query:   query,
			Summary: p.fallbackMsg,
			Sources: []string{},
			Snippet: "响应为空",
		}, nil
	}

	// 提取生成的内容
	// ResponseObject 的结构可能是 Output 或其他字段
	content := ""
	if resp.Output != nil {
		// 尝试从 Output 中提取内容
		if contentBytes, err := json.Marshal(resp.Output); err == nil {
			content = string(contentBytes)
		}
	}

	// 如果 Output 为空，尝试使用原始响应
	if content == "" {
		if contentBytes, err := json.Marshal(resp); err == nil {
			content = string(contentBytes)
		}
	}

	// 尝试从响应中提取 JSON
	aiOverview := p.parseAIOverview(content, query)

	return aiOverview, nil
}

// parseAIOverview 解析豆包返回的内容为 AIOverview 结构
func (p *DoubaoSERPProvider) parseAIOverview(content, query string) *models.AIOverview {
	// 尝试解析 JSON
	var result struct {
		Summary     string   `json:"summary"`
		KeyPoints   []string `json:"key_points"`
		Sources     []string `json:"sources"`
	}

	err := json.Unmarshal([]byte(content), &result)
	if err == nil && result.Summary != "" {
		// JSON 解析成功
		return &models.AIOverview{
			Query:   query,
			Summary: result.Summary,
			Sources: result.Sources,
			Snippet: strings.Join(result.KeyPoints, "\n• "),
		}
	}

	// JSON 解析失败，直接使用原始内容
	return &models.AIOverview{
		Query:   query,
		Summary: content,
		Sources: []string{
			"https://www.baidu.com/s?wd=" + query,
			"https://www.bing.com/search?q=" + query,
			"https://www.google.com/search?q=" + query,
		},
		Snippet: extractFirstSentence(content),
	}
}

// extractFirstSentence 提取第一句话作为片段
func extractFirstSentence(content string) string {
	// 查找第一个句号、问号或感叹号
	for i, c := range content {
		if c == '。' || c == '？' || c == '！' || c == '.' || c == '?' || c == '!' {
			return strings.TrimSpace(content[:i+1])
		}
	}
	// 如果没有找到标点，返回前100个字符
	if len(content) > 100 {
		return strings.TrimSpace(content[:100]) + "..."
	}
	return strings.TrimSpace(content)
}
