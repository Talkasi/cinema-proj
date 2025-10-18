package service

import (
	"context"

	"cw/internal/dataAccess/repository"
	domain "cw/internal/domain/models"
	"cw/internal/utils"
)

type SeatService struct {
	repo repository.SeatRepository
}

func NewSeatService(repo repository.SeatRepository) *SeatService {
	return &SeatService{repo: repo}
}

func (s *SeatService) GetByHall(ctx context.Context, hallId string, filters domain.SeatFilters, page, limit int) (domain.PaginatedResponse[domain.Seat], *utils.Error) {
	data, total, err := s.repo.GetByHall(ctx, hallId, filters, page, limit)
	if err != nil {
		return domain.PaginatedResponse[domain.Seat]{}, err
	}

	return domain.PaginatedResponse[domain.Seat]{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *SeatService) GetByID(ctx context.Context, id string) (domain.Seat, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SeatService) Create(ctx context.Context, seat domain.Seat) (domain.Seat, *utils.Error) {
	return s.repo.Create(ctx, seat)
}

func (s *SeatService) Update(ctx context.Context, id string, seat domain.Seat) (domain.Seat, *utils.Error) {
	return s.repo.Update(ctx, id, seat)
}

func (s *SeatService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}
