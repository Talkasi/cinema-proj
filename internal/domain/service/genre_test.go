package service

import (
	"context"
	"errors"
	"testing"

	domain "cw/internal/domain/models"
	"cw/internal/utils"
)

type genreRepoMock struct {
	getAllFn  func(context.Context, domain.GenreFilters, int, int) ([]domain.Genre, int, *utils.Error)
	getByIDFn func(context.Context, string) (domain.Genre, *utils.Error)
	createFn  func(context.Context, domain.Genre) (domain.Genre, *utils.Error)
	updateFn  func(context.Context, string, domain.Genre) (domain.Genre, *utils.Error)
	deleteFn  func(context.Context, string) *utils.Error
}

func (m genreRepoMock) GetAll(ctx context.Context, f domain.GenreFilters, page, limit int) ([]domain.Genre, int, *utils.Error) {
	return m.getAllFn(ctx, f, page, limit)
}
func (m genreRepoMock) GetByID(ctx context.Context, id string) (domain.Genre, *utils.Error) {
	return m.getByIDFn(ctx, id)
}
func (m genreRepoMock) Create(ctx context.Context, g domain.Genre) (domain.Genre, *utils.Error) {
	return m.createFn(ctx, g)
}
func (m genreRepoMock) Update(ctx context.Context, id string, g domain.Genre) (domain.Genre, *utils.Error) {
	return m.updateFn(ctx, id, g)
}
func (m genreRepoMock) Delete(ctx context.Context, id string) *utils.Error {
	return m.deleteFn(ctx, id)
}

func TestGenreServiceGetAllBuildsPaginatedResponse(t *testing.T) {
	t.Parallel()

	repo := genreRepoMock{
		getAllFn: func(_ context.Context, f domain.GenreFilters, page, limit int) ([]domain.Genre, int, *utils.Error) {
			if f.Name != "Drama" {
				t.Fatalf("filters.Name = %q, want Drama", f.Name)
			}
			if page != 2 || limit != 15 {
				t.Fatalf("page/limit = %d/%d, want 2/15", page, limit)
			}
			return []domain.Genre{{ID: "g1", Name: "Drama"}}, 41, nil
		},
		getByIDFn: func(context.Context, string) (domain.Genre, *utils.Error) { return domain.Genre{}, nil },
		createFn:  func(context.Context, domain.Genre) (domain.Genre, *utils.Error) { return domain.Genre{}, nil },
		updateFn:  func(context.Context, string, domain.Genre) (domain.Genre, *utils.Error) { return domain.Genre{}, nil },
		deleteFn:  func(context.Context, string) *utils.Error { return nil },
	}

	svc := NewGenreService(repo)
	resp, err := svc.GetAll(context.Background(), domain.GenreFilters{Name: "Drama"}, 2, 15)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 41 || resp.Page != 2 || resp.Limit != 15 {
		t.Fatalf("unexpected pagination response: %+v", resp)
	}
	if len(resp.Data) != 1 || resp.Data[0].ID != "g1" {
		t.Fatalf("unexpected data: %+v", resp.Data)
	}
}

func TestGenreServiceGetByIDPropagatesNotFound(t *testing.T) {
	t.Parallel()

	expected := utils.NewNotFound("Genre not found", errors.New("missing"))
	repo := genreRepoMock{
		getAllFn: func(context.Context, domain.GenreFilters, int, int) ([]domain.Genre, int, *utils.Error) {
			return nil, 0, nil
		},
		getByIDFn: func(context.Context, string) (domain.Genre, *utils.Error) { return domain.Genre{}, expected },
		createFn:  func(context.Context, domain.Genre) (domain.Genre, *utils.Error) { return domain.Genre{}, nil },
		updateFn:  func(context.Context, string, domain.Genre) (domain.Genre, *utils.Error) { return domain.Genre{}, nil },
		deleteFn:  func(context.Context, string) *utils.Error { return nil },
	}

	svc := NewGenreService(repo)
	_, err := svc.GetByID(context.Background(), "missing-id")
	if err != expected {
		t.Fatalf("error pointer was not propagated: got %p want %p", err, expected)
	}
}
