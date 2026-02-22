package service

import (
	"context"
	"cw/internal/dataAccess/repository"
	domain "cw/internal/domain/models"
	"cw/internal/utils"
)

type GenreService struct {
	repo repository.GenreRepository
}

func NewGenreService(repo repository.GenreRepository) *GenreService {
	return &GenreService{repo: repo}
}

func (s *GenreService) GetAll(ctx context.Context, filters domain.GenreFilters, page, limit int) (domain.PaginatedResponse[domain.Genre], *utils.Error) {
	data, total, err := s.repo.GetAll(ctx, filters, page, limit)
	if err != nil {
		return domain.PaginatedResponse[domain.Genre]{}, err
	}

	return domain.PaginatedResponse[domain.Genre]{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *GenreService) GetByID(ctx context.Context, id string) (domain.Genre, *utils.Error) {
	return s.repo.GetByID(ctx, id)
}

func (s *GenreService) Create(ctx context.Context, genre domain.Genre) (domain.Genre, *utils.Error) {
	return s.repo.Create(ctx, genre)
}

func (s *GenreService) Update(ctx context.Context, id string, genre domain.Genre) (domain.Genre, *utils.Error) {
	return s.repo.Update(ctx, id, genre)
}

func (s *GenreService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}
