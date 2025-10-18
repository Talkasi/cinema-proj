package service

import (
	"context"

	"cw/internal/dataAccess/repository"
	domain "cw/internal/domain/models"
	"cw/internal/utils"
)

type MovieService struct {
	repo repository.MovieRepository
}

func NewMovieService(repo repository.MovieRepository) *MovieService {
	return &MovieService{repo: repo}
}

func (s *MovieService) GetAll(ctx context.Context, filters domain.MovieFilters, page, limit int) (domain.PaginatedResponse[domain.Movie], *utils.Error) {
	data, total, err := s.repo.GetAll(ctx, filters, page, limit)
	if err != nil {
		return domain.PaginatedResponse[domain.Movie]{}, err
	}

	return domain.PaginatedResponse[domain.Movie]{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *MovieService) GetByID(ctx context.Context, id string) (domain.Movie, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MovieService) Create(ctx context.Context, movie domain.Movie) (domain.Movie, *utils.Error) {
	return s.repo.Create(ctx, movie)
}

func (s *MovieService) Update(ctx context.Context, id string, movie domain.Movie) (domain.Movie, *utils.Error) {
	return s.repo.Update(ctx, id, movie)
}

func (s *MovieService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}

func (s *MovieService) CalculateMovieRating(ctx context.Context, movieID string) (float64, *utils.Error) {
	return s.repo.CalculateMovieRating(ctx, movieID)
}
