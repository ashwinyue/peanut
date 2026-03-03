package llm

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
)

// NewChatModel 创建统一的 ChatModel
// 使用火山引擎 Ark 模型（豆包），如果未配置则返回错误
func NewChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	// 尝试创建 Ark 模型
	cm, err := NewArkChatModel()
	if err == nil {
		return cm, nil
	}

	return nil, fmt.Errorf("未配置有效的 LLM，请设置 ARK_API_KEY 环境变量: %w", err)
}
