package service

import (
	"context"

	"cw/internal/models"
	"cw/internal/repository"
	"cw/internal/utils"
)

type MovieService struct {
	repo repository.MovieRepository
}

func NewMovieService(repo repository.MovieRepository) *MovieService {
	return &MovieService{repo: repo}
}

func (s *MovieService) GetAll(ctx context.Context) ([]models.Movie, *utils.Error) {
	return s.repo.GetAll(ctx)
}

func (s *MovieService) GetByID(ctx context.Context, id string) (models.Movie, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MovieService) Create(ctx context.Context, movie models.Movie) (models.Movie, *utils.Error) {
	return s.repo.Create(ctx, movie)
}

func (s *MovieService) Update(ctx context.Context, movie models.Movie) (models.Movie, *utils.Error) {
	return s.repo.Update(ctx, movie)
}

func (s *MovieService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}

func (s *MovieService) SearchByName(ctx context.Context, name string) ([]models.Movie, *utils.Error) {
	return s.repo.SearchByName(ctx, name)
}

func (s *MovieService) GetByGenres(ctx context.Context, genreIds []string) ([]models.Movie, *utils.Error) {
	return s.repo.GetByGenres(ctx, genreIds)
}
