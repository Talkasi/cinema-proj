package service

import (
	"context"
	"errors"
	"testing"

	domain "cw/internal/domain/models"
	"cw/internal/utils"
)

type movieRepoMock struct {
	getAllFn     func(context.Context, domain.MovieFilters, int, int) ([]domain.Movie, int, *utils.Error)
	getByIDFn    func(context.Context, string) (domain.Movie, *utils.Error)
	createFn     func(context.Context, domain.Movie) (domain.Movie, *utils.Error)
	updateFn     func(context.Context, string, domain.Movie) (domain.Movie, *utils.Error)
	deleteFn     func(context.Context, string) *utils.Error
	calcRatingFn func(context.Context, string) (float64, *utils.Error)
}

func (m movieRepoMock) GetAll(ctx context.Context, f domain.MovieFilters, page, limit int) ([]domain.Movie, int, *utils.Error) {
	return m.getAllFn(ctx, f, page, limit)
}
func (m movieRepoMock) GetByID(ctx context.Context, id string) (domain.Movie, *utils.Error) {
	return m.getByIDFn(ctx, id)
}
func (m movieRepoMock) Create(ctx context.Context, mv domain.Movie) (domain.Movie, *utils.Error) {
	return m.createFn(ctx, mv)
}
func (m movieRepoMock) Update(ctx context.Context, id string, mv domain.Movie) (domain.Movie, *utils.Error) {
	return m.updateFn(ctx, id, mv)
}
func (m movieRepoMock) Delete(ctx context.Context, id string) *utils.Error {
	return m.deleteFn(ctx, id)
}
func (m movieRepoMock) CalculateMovieRating(ctx context.Context, movieID string) (float64, *utils.Error) {
	return m.calcRatingFn(ctx, movieID)
}

func TestMovieServiceCalculateMovieRatingDelegatesToRepo(t *testing.T) {
	t.Parallel()

	repo := movieRepoMock{
		getAllFn: func(context.Context, domain.MovieFilters, int, int) ([]domain.Movie, int, *utils.Error) {
			return nil, 0, nil
		},
		getByIDFn: func(context.Context, string) (domain.Movie, *utils.Error) { return domain.Movie{}, nil },
		createFn:  func(context.Context, domain.Movie) (domain.Movie, *utils.Error) { return domain.Movie{}, nil },
		updateFn:  func(context.Context, string, domain.Movie) (domain.Movie, *utils.Error) { return domain.Movie{}, nil },
		deleteFn:  func(context.Context, string) *utils.Error { return nil },
		calcRatingFn: func(_ context.Context, movieID string) (float64, *utils.Error) {
			if movieID != "m1" {
				t.Fatalf("movieID=%q", movieID)
			}
			return 8.4, nil
		},
	}

	svc := NewMovieService(repo)
	rating, err := svc.CalculateMovieRating(context.Background(), "m1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rating != 8.4 {
		t.Fatalf("rating = %v, want 8.4", rating)
	}
}

func TestMovieServiceCreatePropagatesConflict(t *testing.T) {
	t.Parallel()

	expected := utils.NewConflict("Database conflict", errors.New("duplicate"))
	repo := movieRepoMock{
		getAllFn: func(context.Context, domain.MovieFilters, int, int) ([]domain.Movie, int, *utils.Error) {
			return nil, 0, nil
		},
		getByIDFn:    func(context.Context, string) (domain.Movie, *utils.Error) { return domain.Movie{}, nil },
		createFn:     func(context.Context, domain.Movie) (domain.Movie, *utils.Error) { return domain.Movie{}, expected },
		updateFn:     func(context.Context, string, domain.Movie) (domain.Movie, *utils.Error) { return domain.Movie{}, nil },
		deleteFn:     func(context.Context, string) *utils.Error { return nil },
		calcRatingFn: func(context.Context, string) (float64, *utils.Error) { return 0, nil },
	}

	svc := NewMovieService(repo)
	_, err := svc.Create(context.Background(), domain.Movie{Title: "Duplicate"})
	if err != expected {
		t.Fatalf("error pointer was not propagated: got %p want %p", err, expected)
	}
}
