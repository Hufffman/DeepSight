package model

import "gorm.io/gorm"

// AnalysisReport 分析报告模型
type AnalysisReport struct {
	gorm.Model
	UserID          uint   `gorm:"not null;index" json:"user_id"`
	KnowledgeBaseID uint   `gorm:"not null;index" json:"knowledge_base_id"`
	ReportType      string `gorm:"size:20" json:"report_type"`
	Content         string `gorm:"type:text" json:"content"`
	PlanJSON        string `gorm:"type:text" json:"plan_json"`
	ExecutionLog    string `gorm:"type:text" json:"execution_log"`
}

func (AnalysisReport) TableName() string {
	return "analysis_reports"
}
