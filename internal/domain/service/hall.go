package service

import (
	"context"

	"cw/internal/dataAccess/repository"
	domain "cw/internal/domain/models"
	"cw/internal/utils"
)

type HallService struct {
	repo repository.HallRepository
}

func NewHallService(repo repository.HallRepository) *HallService {
	return &HallService{repo: repo}
}

func (s *HallService) GetAll(ctx context.Context, filters domain.HallFilters, page, limit int) (domain.PaginatedResponse[domain.Hall], *utils.Error) {
	data, total, err := s.repo.GetAll(ctx, filters, page, limit)
	if err != nil {
		return domain.PaginatedResponse[domain.Hall]{}, err
	}

	return domain.PaginatedResponse[domain.Hall]{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *HallService) GetByID(ctx context.Context, id string) (domain.Hall, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *HallService) Create(ctx context.Context, hall domain.Hall) (domain.Hall, *utils.Error) {
	return s.repo.Create(ctx, hall)
}

func (s *HallService) Update(ctx context.Context, id string, hall domain.Hall) (domain.Hall, *utils.Error) {
	return s.repo.Update(ctx, id, hall)
}

func (s *HallService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}
