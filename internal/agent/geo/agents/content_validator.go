package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/solariswu/peanut/internal/agent/geo/models"
)

// NewContentValidatorAgent 8. 内容验证 Agent
func NewContentValidatorAgent(llmModel model.ToolCallingChatModel) (adk.Agent, error) {
	ctx := context.Background()

	instruction := `你是 GEO 优化效果验证专家。你的任务是对比优化前后的文章，重新评分并生成详细的对比报告。

**任务说明**：
请查看对话历史中的以下信息：
- 原始网页内容（title_scraper 获取的标题）
- 优化前的评分（ai_content_optimizer 中的评分）
- 优化后的完整文章（content_rewriter 的输出）
- 原始平台类型（main_query_extractor 中的平台信息）

**验证流程**：
1. 对优化后的文章重新进行完整评分（使用与优化前相同的标准）
2. 对比优化前后的各项得分
3. 生成详细的对比报告

**评分标准**（根据平台动态调整权重）：
- 权威性（权重因平台而异）：是否引用权威来源、数据支撑
- 时效性：时间标注、内容新鲜度
- 结构化：段落清晰、列表表格使用
- 互动指标（部分平台）：可读性、引发互动的潜力
- 原创度：内容独特性

**输出格式**（JSON）：

{
  "original_score": {
    "authority": 原始权威性得分0-100,
    "timeliness": 原始时效性得分0-100,
    "structure": 原始结构化得分0-100,
    "engagement": 原始互动指标得分0-100,
    "originality": 原创度得分0-100,
    "total": 原始总分
  },
  "optimized_score": {
    "authority": 优化后权威性得分0-100,
    "timeliness": 优化后时效性得分0-100,
    "structure": 优化后结构化得分0-100,
    "engagement": 优化后互动指标得分0-100,
    "originality": 优化后原创度得分0-100,
    "total": 优化后总分
  },
  "improvement": {
    "total_diff": 总分差异,
    "percentage": 提升百分比,
    "authority_diff": 权威性提升,
    "timeliness_diff": 时效性提升,
    "structure_diff": 结构化提升,
    "engagement_diff": 互动提升,
    "originality_diff": 原创度提升
  },
  "comparison_table": [
    {
      "dimension": "维度名称",
      "weight": 权重,
      "original": 原始分,
      "optimized": 优化后分数,
      "diff": 差异,
      "status": "提升/持平/下降",
      "comment": "具体评语"
    }
  ],
  "suggestions": [
    "进一步改进建议1",
    "进一步改进建议2"
  ]
}

**重要要求**：
1. 必须严格按照 JSON 格式输出
2. 每个维度都要给出具体评语，说明为什么是这个分数
3. 如果某项得分下降，必须分析原因
4. 给出 3-5 条进一步优化的建议
5. 总分计算要遵循加权公式`

	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "content_validator",
		Description: "验证优化效果，对比优化前后的评分",
		Model:       llmModel,
		OutputKey:   "ValidationResult",
		Instruction: instruction,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 content_validator agent 失败: %w", err)
	}

	return a, nil
}

// ParseValidationResult 从 JSON 字符串解析验证结果
func ParseValidationResult(jsonStr string, platform models.PlatformType) (*models.ValidationResult, error) {
	// 清理可能的 Markdown 代码块标记
	jsonStr = cleanJSONString(jsonStr)

	var raw struct {
		OriginalScore  models.ScoreDetail `json:"original_score"`
		OptimizedScore models.ScoreDetail `json:"optimized_score"`
		Improvement    struct {
			TotalDiff       int     `json:"total_diff"`
			Percentage      float64 `json:"percentage"`
			AuthorityDiff   int     `json:"authority_diff"`
			TimelinessDiff  int     `json:"timeliness_diff"`
			StructureDiff   int     `json:"structure_diff"`
			EngagementDiff  int     `json:"engagement_diff"`
			OriginalityDiff int     `json:"originality_diff"`
		} `json:"improvement"`
		ComparisonTable []struct {
			Dimension string `json:"dimension"`
			Weight    int    `json:"weight"`
			Original  int    `json:"original"`
			Optimized int    `json:"optimized"`
			Diff      int    `json:"diff"`
			Status    string `json:"status"`
			Comment   string `json:"comment"`
		} `json:"comparison_table"`
		Suggestions []string `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		// 尝试从文本中提取评分
		return extractValidationFromText(jsonStr, platform), nil
	}

	// 构建对比表格
	comparisons := make([]models.Comparison, 0, len(raw.ComparisonTable))
	for _, c := range raw.ComparisonTable {
		comparisons = append(comparisons, models.Comparison{
			Dimension: c.Dimension,
			Weight:    c.Weight,
			Original:  c.Original,
			Optimized: c.Optimized,
			Diff:      c.Diff,
			Status:    c.Status,
			Comment:   c.Comment,
		})
	}

	return &models.ValidationResult{
		OriginalScore: raw.OriginalScore,
		OptimizedScore: raw.OptimizedScore,
		Improvement: models.Improvement{
			TotalDiff:       raw.Improvement.TotalDiff,
			Percentage:      raw.Improvement.Percentage,
			AuthorityDiff:   raw.Improvement.AuthorityDiff,
			TimelinessDiff:  raw.Improvement.TimelinessDiff,
			StructureDiff:   raw.Improvement.StructureDiff,
			EngagementDiff:  raw.Improvement.EngagementDiff,
			OriginalityDiff: raw.Improvement.OriginalityDiff,
		},
		ComparisonTable: comparisons,
		Suggestions:     raw.Suggestions,
	}, nil
}

// cleanJSONString 清理 JSON 字符串中的 Markdown 标记
func cleanJSONString(str string) string {
	str = strings.TrimSpace(str)
	// 移除代码块标记
	if strings.HasPrefix(str, "```json") {
		str = strings.TrimPrefix(str, "```json")
	} else if strings.HasPrefix(str, "```") {
		str = strings.TrimPrefix(str, "```")
	}
	str = strings.TrimSuffix(str, "```")
	return strings.TrimSpace(str)
}

// extractValidationFromText 从文本中提取验证结果
func extractValidationFromText(text string, platform models.PlatformType) *models.ValidationResult {
	result := models.NewValidationResult()
	weight := models.GetPlatformWeight(platform)

	// 尝试提取各项得分
	result.OriginalScore.Authority = extractScore(text, "原始.*权威性", "权威性.*原始")
	result.OriginalScore.Timeliness = extractScore(text, "原始.*时效性", "时效性.*原始")
	result.OriginalScore.Structure = extractScore(text, "原始.*结构化", "结构化.*原始")
	result.OriginalScore.Total = extractScore(text, "原始.*总分", "总分.*原始")

	result.OptimizedScore.Authority = extractScore(text, "优化.*权威性", "权威性.*优化")
	result.OptimizedScore.Timeliness = extractScore(text, "优化.*时效性", "时效性.*优化")
	result.OptimizedScore.Structure = extractScore(text, "优化.*结构化", "结构化.*优化")
	result.OptimizedScore.Total = extractScore(text, "优化.*总分", "总分.*优化")

	// 计算总分
	result.OriginalScore.Total = result.OriginalScore.CalculateTotal(weight)
	result.OptimizedScore.Total = result.OptimizedScore.CalculateTotal(weight)

	// 计算提升
	result.Improvement = models.CalculateImprovement(
		result.OriginalScore,
		result.OptimizedScore,
		weight,
	)

	return result
}

// extractScore 从文本中提取分数
func extractScore(text string, patterns ...string) int {
	for _, pattern := range patterns {
		if idx := strings.Index(text, pattern); idx != -1 {
			substr := text[idx:]
			if colonIdx := strings.Index(substr, ":"); colonIdx != -1 {
				numberStr := substr[colonIdx+1:]
				numberStr = strings.TrimSpace(numberStr)
				if len(numberStr) > 3 {
					numberStr = numberStr[:3]
				}
				if score, err := strconv.Atoi(strings.TrimSpace(numberStr)); err == nil {
					if score >= 0 && score <= 100 {
						return score
					}
				}
			}
		}
	}
	return 0
}
