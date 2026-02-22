package handler

import (
	domain "cw/internal/domain/models"
	"cw/internal/domain/service"
	dto "cw/internal/dto/models"
	"cw/internal/utils"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type MovieHandler struct {
	movieService *service.MovieService
}

func NewMovieHandler(ms *service.MovieService) *MovieHandler {
	return &MovieHandler{movieService: ms}
}

// @Summary Poluchit spisok filmov
// @Description Vozvraschaet paginirovannyy spisok vsekh filmov
// @Tags Filmy
// @Produce json
// @Param page query int false "Nomer stranitsy" default(1) minimum(1)
// @Param limit query int false "Kolichestvo elementov na stranitse" default(20) minimum(1) maximum(100)
// @Param genre query string false "Filtr po zhanru"
// @Param title query string false "Poisk po nazvaniyu"
// @Success 200 {object} dto.PaginatedMovieResponse "Spisok filmov"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /movies [get]
func (m *MovieHandler) GetMovies(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	dtoFilters := dto.MovieFilters{
		Title: r.URL.Query().Get("title"),
		Genre: r.URL.Query().Get("genre"),
	}

	domainFilters := domain.MovieFiltersFromDTO(dtoFilters)

	result, err := m.movieService.GetAll(r.Context(), domainFilters, page, limit)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Poluchit film po ID
// @Description Vozvraschaet informatsiyu o filme po ego identifikatoru
// @Tags Filmy
// @Produce json
// @Param id path string true "UUID filma"
// @Success 200 {object} dto.MovieResponse "Informatsiya o filme"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /movies/{id} [get]
func (m *MovieHandler) GetMovieByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	movie, err := m.movieService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(movie); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Sozdat novyy film
// @Description Sozdaet novyy film (tolko dlya administratorov)
// @Tags Filmy
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie body dto.CreateMovieRequest true "Dannye filma"
// @Success 201 {object} dto.CreateResponse "Film sozdan"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Router /movies [post]
func (m *MovieHandler) CreateMovie(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateMovieRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainMovie := domain.CreateMovieFromDTO(req)

	result, err := m.movieService.Create(r.Context(), domainMovie)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(dto.CreateResponse{ID: result.ID}); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Obnovit film
// @Description Polnostyu obnovlyaet informatsiyu o filme (tolko dlya administratorov)
// @Tags Filmy
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID filma"
// @Param movie body dto.UpdateMovieRequest true "Novye dannye filma"
// @Success 200 "Film obnovlen"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /movies/{id} [put]
func (m *MovieHandler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateMovieRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainMovie := domain.UpdateMovieFromDTO(req)

	_, err := m.movieService.Update(r.Context(), id, domainMovie)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Udalit film
// @Description Udalyaet film po identifikatoru (tolko dlya administratorov)
// @Tags Filmy
// @Security BearerAuth
// @Param id path string true "UUID filma"
// @Success 204 "Film udalen"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /movies/{id} [delete]
func (m *MovieHandler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := m.movieService.Delete(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (m *MovieHandler) RegisterRoutes(r chi.Router) {
	r.Route("/movies", func(r chi.Router) {
		r.Get("/", m.GetMovies)
		r.Post("/", m.CreateMovie)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", m.GetMovieByID)
			r.Put("/", m.UpdateMovie)
			r.Delete("/", m.DeleteMovie)
		})
	})
}
