package models

import "time"

// ValidationResult 验证结果
type ValidationResult struct {
	OriginalScore    ScoreDetail  `json:"original_score"`     // 原始评分
	OptimizedScore   ScoreDetail  `json:"optimized_score"`    // 优化后评分
	Improvement      Improvement  `json:"improvement"`        // 提升情况
	ComparisonTable  []Comparison `json:"comparison_table"`   // 对比表格
	Suggestions      []string     `json:"suggestions"`        // 进一步改进建议
	Timestamp        time.Time    `json:"timestamp"`
}

// ScoreDetail 详细评分
type ScoreDetail struct {
	Authority   int `json:"authority"`   // 权威性得分
	Timeliness  int `json:"timeliness"`  // 时效性得分
	Structure   int `json:"structure"`   // 结构化得分
	Engagement  int `json:"engagement"`  // 互动指标得分
	Originality int `json:"originality"` // 原创度得分
	Total       int `json:"total"`       // 总分
}

// Improvement 提升情况
type Improvement struct {
	TotalDiff      int     `json:"total_diff"`       // 总分差异
	Percentage     float64 `json:"percentage"`       // 提升百分比
	AuthorityDiff  int     `json:"authority_diff"`   // 权威性提升
	TimelinessDiff int     `json:"timeliness_diff"`  // 时效性提升
	StructureDiff  int     `json:"structure_diff"`   // 结构化提升
	EngagementDiff int     `json:"engagement_diff"`  // 互动指标提升
	OriginalityDiff int    `json:"originality_diff"` // 原创度提升
}

// Comparison 单维度对比
type Comparison struct {
	Dimension   string `json:"dimension"`    // 维度名称
	Weight      int    `json:"weight"`       // 权重
	Original    int    `json:"original"`     // 原始得分
	Optimized   int    `json:"optimized"`    // 优化后得分
	Diff        int    `json:"diff"`         // 差异
	Status      string `json:"status"`       // 状态：提升/持平/下降
	Comment     string `json:"comment"`      // 评语
}

// CalculateTotal 计算总分
func (s *ScoreDetail) CalculateTotal(weight PlatformWeight) int {
	total := s.Authority*weight.Authority/100 +
		s.Timeliness*weight.Timeliness/100 +
		s.Structure*weight.Structure/100 +
		s.Engagement*weight.Engagement/100 +
		s.Originality*weight.Originality/100
	return total
}

// CalculateImprovement 计算提升情况
func CalculateImprovement(original, optimized ScoreDetail, weight PlatformWeight) Improvement {
	origTotal := original.CalculateTotal(weight)
	optTotal := optimized.CalculateTotal(weight)

	var percentage float64
	if origTotal > 0 {
		percentage = float64(optTotal-origTotal) * 100 / float64(origTotal)
	}

	return Improvement{
		TotalDiff:       optTotal - origTotal,
		Percentage:      percentage,
		AuthorityDiff:   optimized.Authority - original.Authority,
		TimelinessDiff:  optimized.Timeliness - original.Timeliness,
		StructureDiff:   optimized.Structure - original.Structure,
		EngagementDiff:  optimized.Engagement - original.Engagement,
		OriginalityDiff: optimized.Originality - original.Originality,
	}
}

// NewValidationResult 创建验证结果
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Timestamp: time.Now(),
	}
}
