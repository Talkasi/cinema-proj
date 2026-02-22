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

// @Summary Get a list of movies
// @Description Returns a paginated list of all movies
// @Tags Movies
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param genre query string false "Filter by genre"
// @Param title query string false "Search by name"
// @Success 200 {object} dto.PaginatedMovieResponse "Movie list"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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

// @Summary Get a movie by ID
// @Description Returns information about a movie by its identifier
// @Tags Movies
// @Produce json
// @Param id path string true "Movie UUID"
// @Success 200 {object} dto.MovieResponse "Movie information"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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

// @Summary Create a new movie
// @Description Creates a new movie (administrators only)
// @Tags Movies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie body dto.CreateMovieRequest true "Movie data"
// @Success 201 {object} dto.CreateResponse "Movie created"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
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

// @Summary Update a movie
// @Description Fully updates information about a movie (administrators only)
// @Tags Movies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Movie UUID"
// @Param movie body dto.UpdateMovieRequest true "New movie data"
// @Success 200 "Movie updated"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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

// @Summary Delete a movie
// @Description Deletes a movie by identifier (administrators only)
// @Tags Movies
// @Security BearerAuth
// @Param id path string true "Movie UUID"
// @Success 204 "Movie deleted"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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
