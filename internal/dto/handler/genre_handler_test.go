package handler

import (
	"context"
	"cw/internal/dataAccess/repository"
	domain "cw/internal/domain/models"
	"cw/internal/domain/service"
	dto "cw/internal/dto/models"
	"cw/internal/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var _ repository.GenreRepository = (*genreRepoStub)(nil)

type genreRepoStub struct {
	createFn    func(ctx context.Context, genre domain.Genre) (domain.Genre, *utils.Error)
	createCalls int
}

func (s *genreRepoStub) GetAll(ctx context.Context, filters domain.GenreFilters, page, limit int) ([]domain.Genre, int, *utils.Error) {
	return nil, 0, nil
}

func (s *genreRepoStub) GetByID(ctx context.Context, id string) (domain.Genre, *utils.Error) {
	return domain.Genre{}, nil
}

func (s *genreRepoStub) Create(ctx context.Context, genre domain.Genre) (domain.Genre, *utils.Error) {
	s.createCalls++
	if s.createFn != nil {
		return s.createFn(ctx, genre)
	}

	return domain.Genre{ID: "genre-1", Name: genre.Name, Description: genre.Description}, nil
}

func (s *genreRepoStub) Update(ctx context.Context, id string, genre domain.Genre) (domain.Genre, *utils.Error) {
	return domain.Genre{}, nil
}

func (s *genreRepoStub) Delete(ctx context.Context, id string) *utils.Error { return nil }

func newTestGenreHandler(repo *genreRepoStub) *GenreHandler {
	return NewGenreHandler(service.NewGenreService(repo))
}

func TestGenreHandlerCreateGenreHappyPath(t *testing.T) {
	t.Parallel()

	repo := &genreRepoStub{}
	h := newTestGenreHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/genres", strings.NewReader(`{"name":"Drama","description":"Desc"}`))
	rr := httptest.NewRecorder()

	h.CreateGenre(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
	if repo.createCalls != 1 {
		t.Fatalf("createCalls = %d, want 1", repo.createCalls)
	}

	var resp dto.CreateResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID == "" {
		t.Fatal("expected non-empty id")
	}
}

func TestGenreHandlerCreateGenreInvalidJSON(t *testing.T) {
	t.Parallel()

	repo := &genreRepoStub{}
	h := newTestGenreHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/genres", strings.NewReader(`{"name":`))
	rr := httptest.NewRecorder()

	h.CreateGenre(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if repo.createCalls != 0 {
		t.Fatalf("createCalls = %d, want 0", repo.createCalls)
	}

	assertErrorMessagePrefix(t, rr, "Invalid input")
}

func TestGenreHandlerCreateGenreValidationErrorMasked(t *testing.T) {
	t.Parallel()

	repo := &genreRepoStub{}
	h := newTestGenreHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/genres", strings.NewReader(`{"name":"","description":"ok"}`))
	rr := httptest.NewRecorder()

	h.CreateGenre(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if repo.createCalls != 0 {
		t.Fatalf("createCalls = %d, want 0", repo.createCalls)
	}

	assertErrorMessageOnly(t, rr, "Invalid input")
	if strings.Contains(strings.ToLower(rr.Body.String()), "required") ||
		strings.Contains(strings.ToLower(rr.Body.String()), "name") {
		t.Fatalf("unexpected validation details leaked: %s", rr.Body.String())
	}
}

func assertErrorMessageOnly(t *testing.T, rr *httptest.ResponseRecorder, want string) {
	t.Helper()

	var resp dto.ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Message != want {
		t.Fatalf("message = %q, want %q", resp.Message, want)
	}
}

func assertErrorMessagePrefix(t *testing.T, rr *httptest.ResponseRecorder, wantPrefix string) {
	t.Helper()

	var resp dto.ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !strings.HasPrefix(resp.Message, wantPrefix) {
		t.Fatalf("message = %q, want prefix %q", resp.Message, wantPrefix)
	}
}
