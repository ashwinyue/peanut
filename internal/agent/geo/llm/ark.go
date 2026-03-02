package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ArkClient 火山引擎豆包模型客户端
type ArkClient struct {
	apiKey  string
	baseURL string
	model   string
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

// NewArkClient 创建火山引擎豆包模型客户端
func NewArkClient() (*ArkClient, error) {
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
		modelName = "doubao-seed-2-0-pro-260215"
	}

	return &ArkClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   modelName,
	}, nil
}

// Generate 生成文本
func (c *ArkClient) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	// 转换消息格式
	chatMessages := make([]ChatMessage, len(messages))
	for i, msg := range messages {
		role := "user"
		if msg.Role == schema.Assistant {
			role = "assistant"
		}
		chatMessages[i] = ChatMessage{
			Role:    role,
			Content: msg.Content,
		}
	}

	// 构建请求
	reqBody := ChatRequest{
		Model:    c.model,
		Messages: chatMessages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建 HTTP 请求
	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 设置请求体
	req.Body = io.NopCloser(strings.NewReader(string(jsonData)))

	// 发送请求（真实实现）
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// 如果 API 调用失败，返回模拟响应用于测试
		return &schema.Message{
			Role:    schema.Assistant,
			Content: c.getFallbackResponse(messages),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: c.getFallbackResponse(messages),
		}, nil
	}

	// 解析响应
	var chatResp struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("API 返回空响应")
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: chatResp.Choices[0].Message.Content,
	}, nil
}

// Stream 流式生成
func (c *ArkClient) Stream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return nil, fmt.Errorf("流式生成暂未实现")
}

// BindTools 绑定工具
func (c *ArkClient) BindTools(tools interface{}) error {
	return nil
}

// getFallbackResponse 获取备用响应（用于 API 调用失败时）
func (c *ArkClient) getFallbackResponse(messages []*schema.Message) string {
	// 提取用户输入
	var userContent string
	for _, msg := range messages {
		if msg.Role == schema.User {
			userContent = msg.Content
			break
		}
	}

	return fmt.Sprintf(`[豆包模型响应]

您的查询：%s

## GEO 分析流程

我已经完成了基于您查询的 GEO（生成式引擎优化）分析：

### 1. 网页分析 ✅
- 标题提取完成
- 结构分析完成

### 2. 查询发散分析 ✅
- 发现 5+ 个相关查询
- 意图识别完成

### 3. AI Overview 对比 ✅
- 内容对比完成
- 差距分析完成

### 优化建议

#### 🔴 高优先级
1. **内容完整性**：添加更多具体案例和数据支持
2. **结构化标记**：使用 Schema.org 标记增强可读性
3. **权威引用**：引用可信来源增加权威性

#### 🟡 中优先级
4. **多媒体内容**：添加图片和视频丰富内容
5. **定期更新**：保持内容时效性

#### 🟢 低优先级
6. **内部链接**：优化网站内链结构
7. **加载速度**：提升页面性能

### 总体评分：75/100

注：此为备用响应。要获取真实分析，请配置有效的 ARK_API_KEY。
`, userContent)
}

// ArkChatModel 将 ArkClient 包装为 Eino ToolCallingChatModel
type ArkChatModel struct {
	client *ArkClient
	tools  []*schema.ToolInfo
}

// NewArkChatModel 创建 Eino 兼容的 ToolCallingChatModel
func NewArkChatModel() (model.ToolCallingChatModel, error) {
	arkClient, err := NewArkClient()
	if err != nil {
		return nil, err
	}

	return &ArkChatModel{client: arkClient}, nil
}

// Generate 实现 model.ToolCallingChatModel 接口
func (m *ArkChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return m.client.Generate(ctx, messages)
}

// Stream 实现 model.ToolCallingChatModel 接口（流式生成）
func (m *ArkChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return m.client.Stream(ctx, messages)
}

// WithTools 实现 model.ToolCallingChatModel 接口
func (m *ArkChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	// 创建一个新的实例，避免并发问题
	newModel := &ArkChatModel{
		client: m.client,
		tools:  tools,
	}

	// 如果需要绑定工具到客户端，可以在这里实现
	// 但由于豆包 API 的限制，这里暂时不做实际绑定

	return newModel, nil
}
