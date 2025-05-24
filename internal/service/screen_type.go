package service

import (
	"context"

	"cw/internal/domain"
	"cw/internal/repository"
)

type ScreenTypeService struct {
	repo repository.ScreenTypeRepository
}

func NewScreenTypeService(repo repository.ScreenTypeRepository) *ScreenTypeService {
	return &ScreenTypeService{repo: repo}
}

func (s *ScreenTypeService) GetAll(ctx context.Context) ([]domain.ScreenType, error) {
	return s.repo.GetAll(ctx)
}

func (s *ScreenTypeService) GetByID(ctx context.Context, id string) (domain.ScreenType, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ScreenTypeService) Create(ctx context.Context, screenType domain.ScreenType) (domain.ScreenType, error) {
	return s.repo.Create(ctx, screenType)
}

func (s *ScreenTypeService) Update(ctx context.Context, screenType domain.ScreenType) (domain.ScreenType, error) {
	return s.repo.Update(ctx, screenType)
}

func (s *ScreenTypeService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *ScreenTypeService) SearchByName(ctx context.Context, name string) ([]domain.ScreenType, error) {
	return s.repo.SearchByName(ctx, name)
}
