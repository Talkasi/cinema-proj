package service

import (
	"context"

	"cw/internal/dataAccess/repository"
	domain "cw/internal/domain/models"
	"cw/internal/utils"
)

type ScreenTypeService struct {
	repo repository.ScreenTypeRepository
}

func NewScreenTypeService(repo repository.ScreenTypeRepository) *ScreenTypeService {
	return &ScreenTypeService{repo: repo}
}

func (s *ScreenTypeService) GetAll(ctx context.Context, filters domain.ScreenTypeFilters, page, limit int) (domain.PaginatedResponse[domain.ScreenType], *utils.Error) {
	data, total, err := s.repo.GetAll(ctx, filters, page, limit)
	if err != nil {
		return domain.PaginatedResponse[domain.ScreenType]{}, err
	}

	return domain.PaginatedResponse[domain.ScreenType]{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *ScreenTypeService) GetByID(ctx context.Context, id string) (domain.ScreenType, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ScreenTypeService) Create(ctx context.Context, screenType domain.ScreenType) (domain.ScreenType, *utils.Error) {
	return s.repo.Create(ctx, screenType)
}

func (s *ScreenTypeService) Update(ctx context.Context, id string, screenType domain.ScreenType) (domain.ScreenType, *utils.Error) {
	return s.repo.Update(ctx, id, screenType)
}

func (s *ScreenTypeService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}
