package service

import (
	"context"

	"cw/internal/domain"
	"cw/internal/repository"
	"cw/internal/utils"
)

type SeatTypeService struct {
	repo repository.SeatTypeRepository
}

func NewSeatTypeService(repo repository.SeatTypeRepository) *SeatTypeService {
	return &SeatTypeService{repo: repo}
}

func (s *SeatTypeService) GetAll(ctx context.Context) ([]domain.SeatType, *utils.Error) {
	return s.repo.GetAll(ctx)
}

func (s *SeatTypeService) GetByID(ctx context.Context, id string) (domain.SeatType, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SeatTypeService) Create(ctx context.Context, SeatType domain.SeatType) (domain.SeatType, *utils.Error) {
	return s.repo.Create(ctx, SeatType)
}

func (s *SeatTypeService) Update(ctx context.Context, SeatType domain.SeatType) (domain.SeatType, *utils.Error) {
	return s.repo.Update(ctx, SeatType)
}

func (s *SeatTypeService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}

func (s *SeatTypeService) SearchByName(ctx context.Context, name string) ([]domain.SeatType, *utils.Error) {
	return s.repo.SearchByName(ctx, name)
}
