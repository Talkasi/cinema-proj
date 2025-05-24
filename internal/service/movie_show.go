package service

import (
	"context"
	"time"

	"cw/internal/domain"
	"cw/internal/repository"
)

type MovieShowService struct {
	repo repository.MovieShowRepository
}

func NewMovieShowService(repo repository.MovieShowRepository) *MovieShowService {
	return &MovieShowService{repo: repo}
}

func (s *MovieShowService) GetAll(ctx context.Context) ([]domain.MovieShow, error) {
	return s.repo.GetAll(ctx)
}

func (s *MovieShowService) GetByID(ctx context.Context, id string) (domain.MovieShow, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MovieShowService) Create(ctx context.Context, MovieShow domain.MovieShow) (domain.MovieShow, error) {
	return s.repo.Create(ctx, MovieShow)
}

func (s *MovieShowService) Update(ctx context.Context, MovieShow domain.MovieShow) (domain.MovieShow, error) {
	return s.repo.Update(ctx, MovieShow)
}

func (s *MovieShowService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *MovieShowService) GetByMovie(ctx context.Context, movieId string) ([]domain.MovieShow, error) {
	return s.repo.GetByMovie(ctx, movieId)
}

func (s *MovieShowService) GetByDate(ctx context.Context, date time.Time) ([]domain.MovieShow, error) {
	return s.repo.GetByDate(ctx, date)
}

func (s *MovieShowService) GetUpcoming(ctx context.Context, hours int) ([]domain.MovieShow, error) {
	return s.repo.GetUpcoming(ctx, hours)
}
