package service

import (
	"context"
	"errors"

	"DeepSight/internal/dto"
	"DeepSight/internal/repository"

	"gorm.io/gorm"
)

// ResearchRunner executes deep research analysis and emits SSE events.
type ResearchRunner interface {
	Run(ctx context.Context, emit func(dto.AnalysisSSEEvent)) (string, error)
}

// ResearchFactory creates a ResearchRunner for a given user, knowledge base, and conversation.
type ResearchFactory func(userID, kbID, convID uint) ResearchRunner

// AnalysisService orchestrates deep research analysis and analysis reports.
type AnalysisService struct {
	analysisRepo    *repository.AnalysisRepository
	researchFactory ResearchFactory
}

// NewAnalysisService creates an AnalysisService.
func NewAnalysisService(
	analysisRepo *repository.AnalysisRepository,
	factory ResearchFactory,
) *AnalysisService {
	return &AnalysisService{
		analysisRepo:    analysisRepo,
		researchFactory: factory,
	}
}

// StreamAnalysis runs deep research and returns SSE events via channel.
func (s *AnalysisService) StreamAnalysis(ctx context.Context, userID, kbID, convID uint) (<-chan dto.AnalysisSSEEvent, error) {
	runner := s.researchFactory(userID, kbID, convID)

	events := make(chan dto.AnalysisSSEEvent, 20)

	go func() {
		defer close(events)
		runner.Run(ctx, func(event dto.AnalysisSSEEvent) {
			select {
			case events <- event:
			case <-ctx.Done():
			}
		})
	}()

	return events, nil
}

// ListAllReports returns paginated analysis reports for a user across all KBs.
func (s *AnalysisService) ListAllReports(userID uint, page, pageSize int) (*dto.AnalysisReportListResponse, error) {
	page = normalizePage(page)
	pageSize = normalizePageSize(pageSize)

	reports, total, err := s.analysisRepo.GetAnalysisReportsByUserID(userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]dto.AnalysisReportResponse, len(reports))
	for i, r := range reports {
		items[i] = dto.AnalysisReportResponse{
			ID:              r.ID,
			KnowledgeBaseID: r.KnowledgeBaseID,
			ReportType:      r.ReportType,
			Content:         r.Content,
			CreatedAt:       r.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return &dto.AnalysisReportListResponse{
		Reports:  items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetReport returns a single analysis report by ID, with ownership check.
func (s *AnalysisService) GetReport(userID, reportID uint) (*dto.AnalysisReportResponse, error) {
	report, err := s.analysisRepo.GetAnalysisReportByID(reportID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dto.ErrNotFound
		}
		return nil, err
	}
	if report.UserID != userID {
		return nil, dto.ErrForbidden
	}
	return &dto.AnalysisReportResponse{
		ID:              report.ID,
		KnowledgeBaseID: report.KnowledgeBaseID,
		ReportType:      report.ReportType,
		Content:         report.Content,
		CreatedAt:       report.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// DeleteReport deletes an analysis report after ownership check.
func (s *AnalysisService) DeleteReport(userID, reportID uint) error {
	report, err := s.analysisRepo.GetAnalysisReportByID(reportID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ErrNotFound
		}
		return err
	}
	if report.UserID != userID {
		return dto.ErrForbidden
	}
	return s.analysisRepo.DeleteAnalysisReport(reportID)
}
