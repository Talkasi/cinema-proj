package service

import (
	"context"
	"time"

	"cw/internal/domain"
	"cw/internal/repository"
	"cw/internal/utils"
)

type MovieShowService struct {
	repo repository.MovieShowRepository
}

func NewMovieShowService(repo repository.MovieShowRepository) *MovieShowService {
	return &MovieShowService{repo: repo}
}

func (s *MovieShowService) GetAll(ctx context.Context) ([]domain.MovieShow, *utils.Error) {
	return s.repo.GetAll(ctx)
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

func (s *MovieShowService) GetByMovie(ctx context.Context, movieId string) ([]domain.MovieShow, *utils.Error) {
	return s.repo.GetByMovie(ctx, movieId)
}

func (s *MovieShowService) GetByDate(ctx context.Context, date time.Time) ([]domain.MovieShow, *utils.Error) {
	return s.repo.GetByDate(ctx, date)
}

func (s *MovieShowService) GetUpcoming(ctx context.Context, hours int) ([]domain.MovieShow, *utils.Error) {
	return s.repo.GetUpcoming(ctx, hours)
}
