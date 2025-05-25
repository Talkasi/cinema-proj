package service

import (
	"context"

	"cw/internal/models"
	"cw/internal/repository"
	"cw/internal/utils"
)

type ScreenTypeService struct {
	repo repository.ScreenTypeRepository
}

func NewScreenTypeService(repo repository.ScreenTypeRepository) *ScreenTypeService {
	return &ScreenTypeService{repo: repo}
}

func (s *ScreenTypeService) GetAll(ctx context.Context) ([]models.ScreenType, *utils.Error) {
	return s.repo.GetAll(ctx)
}

func (s *ScreenTypeService) GetByID(ctx context.Context, id string) (models.ScreenType, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ScreenTypeService) Create(ctx context.Context, screenType models.ScreenType) (models.ScreenType, *utils.Error) {
	return s.repo.Create(ctx, screenType)
}

func (s *ScreenTypeService) Update(ctx context.Context, screenType models.ScreenType) (models.ScreenType, *utils.Error) {
	return s.repo.Update(ctx, screenType)
}

func (s *ScreenTypeService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}

func (s *ScreenTypeService) SearchByName(ctx context.Context, name string) ([]models.ScreenType, *utils.Error) {
	return s.repo.SearchByName(ctx, name)
}
