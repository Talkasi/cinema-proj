package service

import (
	"context"

	"cw/internal/models"
	"cw/internal/repository"
	"cw/internal/utils"
)

type SeatService struct {
	repo repository.SeatRepository
}

func NewSeatService(repo repository.SeatRepository) *SeatService {
	return &SeatService{repo: repo}
}

func (s *SeatService) GetAll(ctx context.Context) ([]models.Seat, *utils.Error) {
	return s.repo.GetAll(ctx)
}

func (s *SeatService) GetByID(ctx context.Context, id string) (models.Seat, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SeatService) Create(ctx context.Context, seat models.Seat) (models.Seat, *utils.Error) {
	return s.repo.Create(ctx, seat)
}

func (s *SeatService) Update(ctx context.Context, seat models.Seat) (models.Seat, *utils.Error) {
	return s.repo.Update(ctx, seat)
}

func (s *SeatService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}

func (s *SeatService) GetByHall(ctx context.Context, hallId string) ([]models.Seat, *utils.Error) {
	return s.repo.GetByHall(ctx, hallId)
}
