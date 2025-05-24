package service

import (
	"context"

	"cw/internal/domain"
	"cw/internal/repository"
)

type HallService struct {
	repo repository.HallRepository
}

func NewHallService(repo repository.HallRepository) *HallService {
	return &HallService{repo: repo}
}

func (s *HallService) GetAll(ctx context.Context) ([]domain.Hall, error) {
	return s.repo.GetAll(ctx)
}

func (s *HallService) GetByID(ctx context.Context, id string) (domain.Hall, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *HallService) Create(ctx context.Context, hall domain.Hall) (domain.Hall, error) {
	return s.repo.Create(ctx, hall)
}

func (s *HallService) Update(ctx context.Context, hall domain.Hall) (domain.Hall, error) {
	return s.repo.Update(ctx, hall)
}

func (s *HallService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *HallService) SearchByName(ctx context.Context, name string) ([]domain.Hall, error) {
	return s.repo.SearchByName(ctx, name)
}

func (s *HallService) GetByScreenType(ctx context.Context, screenTypeId string) ([]domain.Hall, error) {
	return s.repo.GetByScreenType(ctx, screenTypeId)
}
