package repository

import (
	"time"

	"gorm.io/gorm"
	"github.com/solariswu/peanut/internal/model"
)

// GEOAnalysisRepository GEO 分析仓储
type GEOAnalysisRepository struct {
	db *gorm.DB
}

// NewGEOAnalysisRepository 创建仓储
func NewGEOAnalysisRepository(db *gorm.DB) *GEOAnalysisRepository {
	return &GEOAnalysisRepository{db: db}
}

// Create 创建分析记录
func (r *GEOAnalysisRepository) Create(analysis *model.GEOAnalysis) error {
	return r.db.Create(analysis).Error
}

// GetByID 根据 ID 获取
func (r *GEOAnalysisRepository) GetByID(id int64) (*model.GEOAnalysis, error) {
	var analysis model.GEOAnalysis
	err := r.db.First(&analysis, id).Error
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}

// Update 更新分析记录
func (r *GEOAnalysisRepository) Update(analysis *model.GEOAnalysis) error {
	return r.db.Save(analysis).Error
}

// UpdateFields 更新指定字段
func (r *GEOAnalysisRepository) UpdateFields(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.GEOAnalysis{}).Where("id = ?", id).Updates(fields).Error
}

// List 查询列表
func (r *GEOAnalysisRepository) List(req *model.GEOAnalysisListRequest) ([]model.GEOAnalysis, int64, error) {
	var analyses []model.GEOAnalysis
	var total int64

	query := r.db.Model(&model.GEOAnalysis{})

	// 状态筛选
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 用户筛选
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}

	// 计数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	order := req.OrderBy
	if req.OrderDesc {
		order += " DESC"
	} else {
		order += " ASC"
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order(order).Offset(offset).Limit(req.PageSize).Find(&analyses).Error; err != nil {
		return nil, 0, err
	}

	return analyses, total, nil
}

// GetByURL 根据 URL 获取最新的分析记录
func (r *GEOAnalysisRepository) GetByURL(url string) (*model.GEOAnalysis, error) {
	var analysis model.GEOAnalysis
	err := r.db.Where("url = ?", url).Order("created_at DESC").First(&analysis).Error
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}

// Delete 删除分析记录
func (r *GEOAnalysisRepository) Delete(id int64) error {
	return r.db.Delete(&model.GEOAnalysis{}, id).Error
}

// MarkCompleted 标记为完成
func (r *GEOAnalysisRepository) MarkCompleted(id int64, score int) error {
	now := time.Now()
	return r.db.Model(&model.GEOAnalysis{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       "completed",
			"overall_score": score,
			"completed_at": &now,
		}).Error
}

// MarkFailed 标记为失败
func (r *GEOAnalysisRepository) MarkFailed(id int64, errMsg string) error {
	return r.db.Model(&model.GEOAnalysis{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": errMsg,
		}).Error
}
