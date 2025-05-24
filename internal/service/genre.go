package service

import (
	"context"

	"cw/internal/domain"
	"cw/internal/repository"
)

type GenreService struct {
	repo repository.GenreRepository
}

func NewGenreService(repo repository.GenreRepository) *GenreService {
	return &GenreService{repo: repo}
}

func (s *GenreService) GetAll(ctx context.Context) ([]domain.Genre, error) {
	return s.repo.GetAll(ctx)
}

func (s *GenreService) GetByID(ctx context.Context, id string) (domain.Genre, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *GenreService) Create(ctx context.Context, genre domain.Genre) (domain.Genre, error) {
	return s.repo.Create(ctx, genre)
}

func (s *GenreService) Update(ctx context.Context, genre domain.Genre) (domain.Genre, error) {
	return s.repo.Update(ctx, genre)
}

func (s *GenreService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *GenreService) SearchByName(ctx context.Context, name string) ([]domain.Genre, error) {
	return s.repo.SearchByName(ctx, name)
}
