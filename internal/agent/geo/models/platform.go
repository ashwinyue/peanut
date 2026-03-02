package models

// PlatformType 平台类型
type PlatformType string

const (
	// PlatformDoubao 豆包
	PlatformDoubao PlatformType = "doubao"
	// PlatformWeChat 微信公众号/搜一搜
	PlatformWeChat PlatformType = "wechat"
	// PlatformZhihu 知乎
	PlatformZhihu PlatformType = "zhihu"
	// PlatformXiaohongshu 小红书
	PlatformXiaohongshu PlatformType = "xiaohongshu"
	// PlatformWenxin 百度文心一言
	PlatformWenxin PlatformType = "wenxin"
	// PlatformYuanbao 腾讯元宝
	PlatformYuanbao PlatformType = "yuanbao"
)

// PlatformWeight 平台权重配置
type PlatformWeight struct {
	Authority   int `json:"authority"`   // 权威性
	Timeliness  int `json:"timeliness"`  // 时效性
	Structure   int `json:"structure"`   // 结构化
	Engagement  int `json:"engagement"`  // 互动指标（平台特有）
	Originality int `json:"originality"` // 原创度
}

// PlatformConfig 平台配置
type PlatformConfig struct {
	Type        PlatformType `json:"type"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Weight      PlatformWeight `json:"weight"`
}

// PlatformWeights 各平台默认权重配置
var PlatformWeights = map[PlatformType]PlatformWeight{
	PlatformDoubao: {
		Authority:   40,
		Timeliness:  25,
		Structure:   20,
		Engagement:  10,
		Originality: 5,
	},
	PlatformWeChat: {
		Authority:   25,
		Timeliness:  20,
		Structure:   15,
		Engagement:  30,
		Originality: 10,
	},
	PlatformZhihu: {
		Authority:   35,
		Timeliness:  15,
		Structure:   25,
		Engagement:  20,
		Originality: 5,
	},
	PlatformXiaohongshu: {
		Authority:   10,
		Timeliness:  20,
		Structure:   20,
		Engagement:  40,
		Originality: 10,
	},
	PlatformWenxin: {
		Authority:   40,
		Timeliness:  20,
		Structure:   20,
		Engagement:  10,
		Originality: 10,
	},
	PlatformYuanbao: {
		Authority:   35,
		Timeliness:  25,
		Structure:   20,
		Engagement:  10,
		Originality: 10,
	},
}

// PlatformNames 平台显示名称
var PlatformNames = map[PlatformType]string{
	PlatformDoubao:      "豆包",
	PlatformWeChat:      "微信公众号",
	PlatformZhihu:       "知乎",
	PlatformXiaohongshu: "小红书",
	PlatformWenxin:      "百度文心一言",
	PlatformYuanbao:     "腾讯元宝",
}

// GetPlatformWeight 获取平台权重配置
func GetPlatformWeight(platform PlatformType) PlatformWeight {
	if weight, ok := PlatformWeights[platform]; ok {
		return weight
	}
	// 默认返回豆包元宝配置
	return PlatformWeights[PlatformDoubao]
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
		{Type: PlatformDoubao, Name: "豆包", Description: "字节跳动旗下 AI 搜索", Weight: PlatformWeights[PlatformDoubao]},
		{Type: PlatformWeChat, Name: "微信公众号", Description: "微信生态内容优化", Weight: PlatformWeights[PlatformWeChat]},
		{Type: PlatformZhihu, Name: "知乎", Description: "专业知识问答平台", Weight: PlatformWeights[PlatformZhihu]},
		{Type: PlatformXiaohongshu, Name: "小红书", Description: "生活方式种草平台", Weight: PlatformWeights[PlatformXiaohongshu]},
		{Type: PlatformWenxin, Name: "百度文心一言", Description: "百度 AI 搜索", Weight: PlatformWeights[PlatformWenxin]},
		{Type: PlatformYuanbao, Name: "腾讯元宝", Description: "腾讯 AI 助手", Weight: PlatformWeights[PlatformYuanbao]},
	}
}
