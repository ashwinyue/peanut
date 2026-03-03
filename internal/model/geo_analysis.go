package model

import (
	"time"

	"gorm.io/gorm"
)

// GEOAnalysis GEO 分析记录
type GEOAnalysis struct {
	BaseModel
	URL           string  `json:"url" gorm:"type:varchar(500);not null;index"`
	Title         string  `json:"title" gorm:"type:varchar(500)"`
	MainQuery     string  `json:"main_query" gorm:"type:varchar(200)"`
	Platform      string  `json:"platform" gorm:"type:varchar(20);default:'google'"` // 目标平台：google
	OverallScore  int     `json:"overall_score" gorm:"type:int;default:0"`
	OptimizedScore int    `json:"optimized_score" gorm:"type:int;default:0"` // 优化后评分
	Status        string  `json:"status" gorm:"type:varchar(20);index"` // pending, processing, completed, failed
	ErrorMessage  string  `json:"error_message,omitempty" gorm:"type:text"`

	// 中间结果
	QueryFanout        string `json:"query_fanout,omitempty" gorm:"type:text"`
	AIOverview         string `json:"ai_overview,omitempty" gorm:"type:text"`
	QueryFanoutSummary string `json:"query_fanout_summary,omitempty" gorm:"type:text"`
	OptimizationReport string `json:"optimization_report,omitempty" gorm:"type:text"`
	OptimizedArticle   string `json:"optimized_article,omitempty" gorm:"type:text"`

	// 统计信息
	ContentGaps           string `json:"content_gaps,omitempty" gorm:"type:text"` // JSON 数组
	OptimizationSuggestions string `json:"optimization_suggestions,omitempty" gorm:"type:text"` // JSON 数组

	// 验证结果
	ValidationResult string `json:"validation_result,omitempty" gorm:"type:text"` // JSON 格式的验证结果

	// 元数据
	UserID     *int64    `json:"user_id,omitempty" gorm:"index"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// TableName 指定表名
func (GEOAnalysis) TableName() string {
	return "geo_analyses"
}

// BeforeCreate GORM hook
func (g *GEOAnalysis) BeforeCreate(tx *gorm.DB) error {
	if g.Status == "" {
		g.Status = "pending"
	}
	return nil
}

// GEOAnalysisCreateRequest 创建请求
type GEOAnalysisCreateRequest struct {
	URL      string `json:"url" binding:"required"`
	Platform string `json:"platform"` // 目标平台：google
}

// GEOAnalysisListRequest 列表查询请求
type GEOAnalysisListRequest struct {
	Page      int     `form:"page,default=1"`
	PageSize  int     `form:"page_size,default=10"`
	Status    string  `form:"status"`
	UserID    *int64  `form:"user_id"`
	OrderBy   string  `form:"order_by,default=created_at"`
	OrderDesc bool    `form:"order_desc,default=true"`
}

// GEOAnalysisResponse 响应
type GEOAnalysisResponse struct {
	ID             int64     `json:"id"`
	URL            string    `json:"url"`
	Title          string    `json:"title"`
	MainQuery      string    `json:"main_query"`
	Platform       string    `json:"platform"`        // 目标平台
	OverallScore   int       `json:"overall_score"`
	OptimizedScore int       `json:"optimized_score"` // 优化后评分
	Status         string    `json:"status"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	QueryFanout        string `json:"query_fanout,omitempty"`
	AIOverview         string `json:"ai_overview,omitempty"`
	QueryFanoutSummary string `json:"query_fanout_summary,omitempty"`
	OptimizationReport string `json:"optimization_report,omitempty"`
	OptimizedArticle   string `json:"optimized_article,omitempty"`
	ContentGaps           string `json:"content_gaps,omitempty"`
	OptimizationSuggestions string `json:"optimization_suggestions,omitempty"`
	ValidationResult   string `json:"validation_result,omitempty"` // 验证结果
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
