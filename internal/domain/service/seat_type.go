package service

import (
	"context"

	"cw/internal/dataAccess/repository"
	domain "cw/internal/domain/models"
	"cw/internal/utils"
)

type SeatTypeService struct {
	repo repository.SeatTypeRepository
}

func NewSeatTypeService(repo repository.SeatTypeRepository) *SeatTypeService {
	return &SeatTypeService{repo: repo}
}

func (s *SeatTypeService) GetAll(ctx context.Context, filters domain.SeatTypeFilters, page, limit int) (domain.PaginatedResponse[domain.SeatType], *utils.Error) {
	data, total, err := s.repo.GetAll(ctx, filters, page, limit)
	if err != nil {
		return domain.PaginatedResponse[domain.SeatType]{}, err
	}

	return domain.PaginatedResponse[domain.SeatType]{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *SeatTypeService) GetByID(ctx context.Context, id string) (domain.SeatType, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SeatTypeService) Create(ctx context.Context, seatType domain.SeatType) (domain.SeatType, *utils.Error) {
	return s.repo.Create(ctx, seatType)
}

func (s *SeatTypeService) Update(ctx context.Context, id string, seatType domain.SeatType) (domain.SeatType, *utils.Error) {
	return s.repo.Update(ctx, id, seatType)
}

func (s *SeatTypeService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}
