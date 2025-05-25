package service

import (
	"context"
	"time"

	"cw/internal/models"
	"cw/internal/repository"
	"cw/internal/utils"
)

type MovieShowService struct {
	repo repository.MovieShowRepository
}

func NewMovieShowService(repo repository.MovieShowRepository) *MovieShowService {
	return &MovieShowService{repo: repo}
}

func (s *MovieShowService) GetAll(ctx context.Context) ([]models.MovieShow, *utils.Error) {
	return s.repo.GetAll(ctx)
}

func (s *MovieShowService) GetByID(ctx context.Context, id string) (models.MovieShow, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MovieShowService) Create(ctx context.Context, MovieShow models.MovieShow) (models.MovieShow, *utils.Error) {
	return s.repo.Create(ctx, MovieShow)
}

func (s *MovieShowService) Update(ctx context.Context, MovieShow models.MovieShow) (models.MovieShow, *utils.Error) {
	return s.repo.Update(ctx, MovieShow)
}

func (s *MovieShowService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}

func (s *MovieShowService) GetByMovie(ctx context.Context, movieId string) ([]models.MovieShow, *utils.Error) {
	return s.repo.GetByMovie(ctx, movieId)
}

func (s *MovieShowService) GetByDate(ctx context.Context, date time.Time) ([]models.MovieShow, *utils.Error) {
	return s.repo.GetByDate(ctx, date)
}

func (s *MovieShowService) GetUpcoming(ctx context.Context, hours int) ([]models.MovieShow, *utils.Error) {
	return s.repo.GetUpcoming(ctx, hours)
}
