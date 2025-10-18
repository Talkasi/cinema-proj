package service

import (
	"context"

	"cw/internal/dataAccess/repository"
	domain "cw/internal/domain/models"
	"cw/internal/utils"
)

type MovieShowService struct {
	repo repository.MovieShowRepository
}

func NewMovieShowService(repo repository.MovieShowRepository) *MovieShowService {
	return &MovieShowService{repo: repo}
}

func (s *MovieShowService) GetAll(ctx context.Context, filters domain.MovieShowFilters, page, limit int) (domain.PaginatedResponse[domain.MovieShow], *utils.Error) {
	data, total, err := s.repo.GetAll(ctx, filters, page, limit)
	if err != nil {
		return domain.PaginatedResponse[domain.MovieShow]{}, err
	}

	return domain.PaginatedResponse[domain.MovieShow]{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *MovieShowService) GetByID(ctx context.Context, id string) (domain.MovieShow, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MovieShowService) Create(ctx context.Context, movieShow domain.MovieShow) (domain.MovieShow, *utils.Error) {
	return s.repo.Create(ctx, movieShow)
}

func (s *MovieShowService) Update(ctx context.Context, id string, movieShow domain.MovieShow) (domain.MovieShow, *utils.Error) {
	return s.repo.Update(ctx, id, movieShow)
}

func (s *MovieShowService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}
