package models

// PlatformType 平台类型
type PlatformType string

const (
	// PlatformGoogle Google AI Overview
	PlatformGoogle PlatformType = "google"
)

// PlatformWeight 平台权重配置
type PlatformWeight struct {
	Authority   int `json:"authority"`   // 权威性
	Timeliness  int `json:"timeliness"`  // 时效性
	Structure   int `json:"structure"`   // 结构化
	Engagement  int `json:"engagement"`  // 互动指标
	Originality int `json:"originality"` // 原创度
}

// PlatformConfig 平台配置
type PlatformConfig struct {
	Type        PlatformType   `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Weight      PlatformWeight `json:"weight"`
}

// PlatformWeights Google AI Overview 权重配置
// 参考: Google AI Overview 的核心目标是为用户提供准确、权威、时效性强的答案
var PlatformWeights = map[PlatformType]PlatformWeight{
	PlatformGoogle: {
		Authority:   45, // 权威性最重要 - 引用官方、学术、权威媒体
		Timeliness:  30, // 时效性次之 - 最新数据和信息
		Structure:   15, // 结构化帮助 AI 理解和提取
		Engagement:  0,  // AI Overview 不关注互动指标
		Originality: 10, // 原创度有一定价值
	},
}

// PlatformNames 平台显示名称
var PlatformNames = map[PlatformType]string{
	PlatformGoogle: "Google AI Overview",
}

// GetPlatformWeight 获取平台权重配置
func GetPlatformWeight(platform PlatformType) PlatformWeight {
	if weight, ok := PlatformWeights[platform]; ok {
		return weight
	}
	// 默认返回 Google 配置
	return PlatformWeights[PlatformGoogle]
}

// GetPlatformName 获取平台显示名称
func GetPlatformName(platform PlatformType) string {
	if name, ok := PlatformNames[platform]; ok {
		return name
	}
	return string(platform)
}

// IsValidPlatform 检查平台是否有效
func IsValidPlatform(platform PlatformType) bool {
	_, ok := PlatformWeights[platform]
	return ok
}

// AllPlatforms 获取所有支持的平台
func AllPlatforms() []PlatformConfig {
	return []PlatformConfig{
		{Type: PlatformGoogle, Name: "Google AI Overview", Description: "Google 搜索 AI 摘要", Weight: PlatformWeights[PlatformGoogle]},
	}
}
