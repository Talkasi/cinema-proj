package handler

import (
	domain "cw/internal/domain/models"
	"cw/internal/domain/service"
	dto "cw/internal/dto/models"
	"cw/internal/utils"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type MovieShowHandler struct {
	movieShowService *service.MovieShowService
}

func NewMovieShowHandler(ms *service.MovieShowService) *MovieShowHandler {
	return &MovieShowHandler{movieShowService: ms}
}

// @Summary Get a list of movie shows
// @Description Returns a paginated list of all movie shows with filtering
// @Tags Movie Shows
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param movie_id query string false "Filter by movie ID (multiple comma-separated values allowed)"
// @Param hall_id query string false "Filter by hall ID (multiple comma-separated values allowed)"
// @Param language query string false "Filter by screening languages (multiple comma-separated values allowed)"
// @Param date query string false "Filter by date (YYYY-MM-DD)"
// @Param start_time_from query string false "Filter by start time (from)"
// @Param start_time_to query string false "Filter by start time (to)"
// @Success 200 {object} dto.PaginatedMovieShowResponse "Movie show list"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /movie-shows [get]
func (ms *MovieShowHandler) GetMovieShows(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	dtoFilters := dto.MovieShowFilters{
		MovieIDs:      splitCommaSeparated(r.URL.Query().Get("movie_id")),
		HallIDs:       splitCommaSeparated(r.URL.Query().Get("hall_id")),
		Languages:     splitCommaSeparated(r.URL.Query().Get("language")),
		Date:          r.URL.Query().Get("date"),
		StartTimeFrom: r.URL.Query().Get("start_time_from"),
		StartTimeTo:   r.URL.Query().Get("start_time_to"),
	}

	domainFilters := domain.MovieShowFiltersFromDTO(dtoFilters)

	result, err := ms.movieShowService.GetAll(r.Context(), domainFilters, page, limit)
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

// @Summary Get a movie show by ID
// @Description Returns information about a movie show by its identifier
// @Tags Movie Shows
// @Produce json
// @Param id path string true "Movie show UUID"
// @Success 200 {object} dto.MovieShowResponse "Movie show information"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /movie-shows/{id} [get]
func (ms *MovieShowHandler) GetMovieShowByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	movieShow, err := ms.movieShowService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(movieShow); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Create a movie show
// @Description Creates a new movie show (administrators only)
// @Tags Movie Shows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie_show body dto.CreateMovieShowRequest true "Movie show data"
// @Success 201 {object} dto.CreateResponse "Movie show created"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Router /movie-shows [post]
func (ms *MovieShowHandler) CreateMovieShow(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateMovieShowRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainMovieShow := domain.CreateMovieShowFromDTO(req)

	result, err := ms.movieShowService.Create(r.Context(), domainMovieShow)
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

// @Summary Update a movie show
// @Description Updates information about a movie show (administrators only)
// @Tags Movie Shows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Movie show UUID"
// @Param movie_show body dto.UpdateMovieShowRequest true "New movie show data"
// @Success 200 "Movie show updated"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /movie-shows/{id} [put]
func (ms *MovieShowHandler) UpdateMovieShow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateMovieShowRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainMovieShow := domain.UpdateMovieShowFromDTO(req)

	_, err := ms.movieShowService.Update(r.Context(), id, domainMovieShow)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Delete a movie show
// @Description Deletes a movie show by identifier (administrators only)
// @Tags Movie Shows
// @Security BearerAuth
// @Param id path string true "Movie show UUID"
// @Success 204 "Movie show deleted"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /movie-shows/{id} [delete]
func (ms *MovieShowHandler) DeleteMovieShow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := ms.movieShowService.Delete(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func splitCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func (ms *MovieShowHandler) RegisterRoutes(r chi.Router) {
	r.Route("/movie-shows", func(r chi.Router) {
		r.Get("/", ms.GetMovieShows)
		r.Post("/", ms.CreateMovieShow)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", ms.GetMovieShowByID)
			r.Put("/", ms.UpdateMovieShow)
			r.Delete("/", ms.DeleteMovieShow)
		})
	})
}
