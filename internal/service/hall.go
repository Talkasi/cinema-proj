package service

import (
	"context"

	"cw/internal/models"
	"cw/internal/repository"
	"cw/internal/utils"
)

type HallService struct {
	repo repository.HallRepository
}

func NewHallService(repo repository.HallRepository) *HallService {
	return &HallService{repo: repo}
}

func (s *HallService) GetAll(ctx context.Context) ([]models.Hall, *utils.Error) {
	return s.repo.GetAll(ctx)
}

func (s *HallService) GetByID(ctx context.Context, id string) (models.Hall, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *HallService) Create(ctx context.Context, hall models.Hall) (models.Hall, *utils.Error) {
	return s.repo.Create(ctx, hall)
}

func (s *HallService) Update(ctx context.Context, hall models.Hall) (models.Hall, *utils.Error) {
	return s.repo.Update(ctx, hall)
}

func (s *HallService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}

func (s *HallService) SearchByName(ctx context.Context, name string) ([]models.Hall, *utils.Error) {
	return s.repo.SearchByName(ctx, name)
}

func (s *HallService) GetByScreenType(ctx context.Context, screenTypeId string) ([]models.Hall, *utils.Error) {
	return s.repo.GetByScreenType(ctx, screenTypeId)
}
