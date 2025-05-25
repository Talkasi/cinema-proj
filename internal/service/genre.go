package service

import (
	"context"

	"cw/internal/models"
	"cw/internal/repository"
	"cw/internal/utils"
)

type GenreService struct {
	repo repository.GenreRepository
}

func NewGenreService(repo repository.GenreRepository) *GenreService {
	return &GenreService{repo: repo}
}

func (s *GenreService) GetAll(ctx context.Context) ([]models.Genre, *utils.Error) {
	return s.repo.GetAll(ctx)
}

func (s *GenreService) GetByID(ctx context.Context, id string) (models.Genre, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *GenreService) Create(ctx context.Context, genre models.Genre) (models.Genre, *utils.Error) {
	return s.repo.Create(ctx, genre)
}

func (s *GenreService) Update(ctx context.Context, genre models.Genre) (models.Genre, *utils.Error) {
	return s.repo.Update(ctx, genre)
}

func (s *GenreService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}

func (s *GenreService) SearchByName(ctx context.Context, name string) ([]models.Genre, *utils.Error) {
	return s.repo.SearchByName(ctx, name)
}
