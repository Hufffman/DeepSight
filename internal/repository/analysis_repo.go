package repository

import (
	"DeepSight/internal/model"

	"gorm.io/gorm"
)

type AnalysisRepository struct {
	db *gorm.DB
}

func NewAnalysisRepository(db *gorm.DB) *AnalysisRepository {
	return &AnalysisRepository{db: db}
}

// -- AnalysisReport --

func (r *AnalysisRepository) CreateAnalysisReport(report *model.AnalysisReport) error {
	return r.db.Create(report).Error
}

func (r *AnalysisRepository) GetAnalysisReportsByUserID(userID uint, page, pageSize int) ([]model.AnalysisReport, int64, error) {
	var reports []model.AnalysisReport
	var total int64
	if err := r.db.Model(&model.AnalysisReport{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reports).Error; err != nil {
		return nil, 0, err
	}
	return reports, total, nil
}

func (r *AnalysisRepository) GetAnalysisReportByID(id uint) (*model.AnalysisReport, error) {
	var report model.AnalysisReport
	err := r.db.First(&report, id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *AnalysisRepository) DeleteAnalysisReport(id uint) error {
	return r.db.Delete(&model.AnalysisReport{}, id).Error
}
