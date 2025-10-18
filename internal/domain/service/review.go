package service

import (
	"context"

	"cw/internal/dataAccess/repository"
	domain "cw/internal/domain/models"
	"cw/internal/utils"
)

type ReviewService struct {
	repo repository.ReviewRepository
}

func NewReviewService(repo repository.ReviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}

func (s *ReviewService) GetAll(ctx context.Context, filters domain.ReviewFilters, page, limit int) (domain.PaginatedResponse[domain.Review], *utils.Error) {
	data, total, err := s.repo.GetAll(ctx, filters, page, limit)
	if err != nil {
		return domain.PaginatedResponse[domain.Review]{}, err
	}

	return domain.PaginatedResponse[domain.Review]{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *ReviewService) CreateForMovie(ctx context.Context, movieId string, review domain.Review) (domain.Review, *utils.Error) {
	return s.repo.CreateForMovie(ctx, movieId, review)
}

func (s *ReviewService) Update(ctx context.Context, id string, review domain.Review) (domain.Review, *utils.Error) {
	return s.repo.Update(ctx, id, review)
}

func (s *ReviewService) Delete(ctx context.Context, id string) *utils.Error {
	return s.repo.Delete(ctx, id)
}
