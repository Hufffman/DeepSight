package service

import (
	"errors"

	"DeepSight/internal/dto"
	"DeepSight/internal/model"
	"DeepSight/internal/repository"

	"gorm.io/gorm"
)

type KnowledgeBaseService struct {
	repo     *repository.KnowledgeBaseRepository
	fileRepo *repository.FileRepository
}

func NewKnowledgeBaseService(repo *repository.KnowledgeBaseRepository, fileRepo *repository.FileRepository) *KnowledgeBaseService {
	return &KnowledgeBaseService{repo: repo, fileRepo: fileRepo}
}

func (s *KnowledgeBaseService) CreateKnowledgeBase(userID uint, name, description string) (*model.KnowledgeBase, error) {
	kb := &model.KnowledgeBase{
		Name:        name,
		Description: description,
		UserID:      userID,
	}

	if err := s.repo.Create(kb); err != nil {
		return nil, err
	}

	return kb, nil
}

func (s *KnowledgeBaseService) GetKnowledgeBaseByID(id uint) (*model.KnowledgeBase, error) {
	return s.repo.GetByID(id)
}

func (s *KnowledgeBaseService) GetKnowledgeBaseByIDWithFiles(id uint) (*model.KnowledgeBase, error) {
	return s.repo.GetByIDWithFiles(id)
}

func (s *KnowledgeBaseService) ListKnowledgeBases(userID uint, page, pageSize int) ([]model.KnowledgeBase, int64, error) {
	page = normalizePage(page)
	pageSize = normalizePageSize(pageSize)
	return s.repo.GetByUserID(userID, page, pageSize)
}

func (s *KnowledgeBaseService) UpdateKnowledgeBase(id uint, name, description string) (*model.KnowledgeBase, error) {
	kb, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		kb.Name = name
	}
	if description != "" {
		kb.Description = description
	}

	if err := s.repo.Update(kb); err != nil {
		return nil, err
	}

	return kb, nil
}

func (s *KnowledgeBaseService) DeleteKnowledgeBase(id uint) error {
	return s.repo.Delete(id)
}

func (s *KnowledgeBaseService) ListKnowledgeBaseResponses(userID uint, page, pageSize int) (*dto.KBListResponse, error) {
	kbs, total, err := s.ListKnowledgeBases(userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	// Collect KB IDs for batch file count query
	kbIDs := make([]uint, len(kbs))
	for i := range kbs {
		kbIDs[i] = kbs[i].ID
	}
	fileCounts, _ := s.fileRepo.CountByKBIDs(kbIDs)

	responses := make([]dto.KBResponse, len(kbs))
	for i := range kbs {
		responses[i] = ToKBResponse(&kbs[i], fileCounts[kbs[i].ID])
	}

	return &dto.KBListResponse{
		KnowledgeBases: responses,
		Total:          total,
		Page:           normalizePage(page),
		PageSize:       normalizePageSize(pageSize),
	}, nil
}

func (s *KnowledgeBaseService) GetKnowledgeBaseDetailForUser(userID, id uint) (*dto.KBDetailResponse, error) {
	kb, err := s.GetKnowledgeBaseByIDWithFiles(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dto.ErrNotFound
		}
		return nil, err
	}
	if kb.UserID != userID {
		return nil, dto.ErrForbidden
	}

	response := ToKBDetailResponse(kb)
	return &response, nil
}

func (s *KnowledgeBaseService) UpdateKnowledgeBaseForUser(userID, id uint, name, description string) (*dto.KBResponse, error) {
	kb, err := s.GetKnowledgeBaseByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dto.ErrNotFound
		}
		return nil, err
	}
	if kb.UserID != userID {
		return nil, dto.ErrForbidden
	}

	kb, err = s.UpdateKnowledgeBase(id, name, description)
	if err != nil {
		return nil, err
	}

	fileCounts, _ := s.fileRepo.CountByKBIDs([]uint{kb.ID})
	response := ToKBResponse(kb, fileCounts[kb.ID])
	return &response, nil
}

func (s *KnowledgeBaseService) DeleteKnowledgeBaseForUser(userID, id uint) error {
	kb, err := s.GetKnowledgeBaseByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ErrNotFound
		}
		return err
	}
	if kb.UserID != userID {
		return dto.ErrForbidden
	}

	return s.DeleteKnowledgeBase(id)
}

func ToKBResponse(kb *model.KnowledgeBase, fileCount int64) dto.KBResponse {
	return dto.KBResponse{
		ID:          kb.ID,
		Name:        kb.Name,
		Description: kb.Description,
		UserID:      kb.UserID,
		FileCount:   fileCount,
		CreatedAt:   kb.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   kb.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func ToKBDetailResponse(kb *model.KnowledgeBase) dto.KBDetailResponse {
	files := make([]dto.FileSummary, len(kb.Files))
	for i, file := range kb.Files {
		files[i] = dto.FileSummary{
			ID:        file.ID,
			FileName:  file.FileName,
			FileType:  file.FileType,
			Status:    file.Status,
			CreatedAt: file.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return dto.KBDetailResponse{
		ID:          kb.ID,
		Name:        kb.Name,
		Description: kb.Description,
		UserID:      kb.UserID,
		Files:       files,
		CreatedAt:   kb.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   kb.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
