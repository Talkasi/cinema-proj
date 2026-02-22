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

// @Summary Poluchit spisok kinoseansov
// @Description Vozvraschaet paginirovannyy spisok vsekh kinoseansov s filtratsiey
// @Tags Kinoseansy
// @Produce json
// @Param page query int false "Nomer stranitsy" default(1) minimum(1)
// @Param limit query int false "Kolichestvo elementov na stranitse" default(20) minimum(1) maximum(100)
// @Param movie_id query string false "Filtr po ID filma (mozhno ukazat neskolko cherez zapyatuyu)"
// @Param hall_id query string false "Filtr po ID zala (mozhno ukazat neskolko cherez zapyatuyu)"
// @Param language query string false "Filtr po yazykam pokaza (mozhno ukazat neskolko cherez zapyatuyu)"
// @Param date query string false "Filtr po date (YYYY-MM-DD)"
// @Param start_time_from query string false "Filtr po vremeni nachala (ot)"
// @Param start_time_to query string false "Filtr po vremeni nachala (do)"
// @Success 200 {object} dto.PaginatedMovieShowResponse "Spisok kinoseansov"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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

// @Summary Poluchit kinoseans po ID
// @Description Vozvraschaet informatsiyu o kinoseanse po ego identifikatoru
// @Tags Kinoseansy
// @Produce json
// @Param id path string true "UUID kinoseansa"
// @Success 200 {object} dto.MovieShowResponse "Informatsiya o kinoseanse"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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

// @Summary Sozdat kinoseans
// @Description Sozdaet novyy kinoseans (tolko dlya administratorov)
// @Tags Kinoseansy
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie_show body dto.CreateMovieShowRequest true "Dannye kinoseansa"
// @Success 201 {object} dto.CreateResponse "Kinoseans sozdan"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
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

// @Summary Obnovit kinoseans
// @Description Obnovlyaet informatsiyu o kinoseanse (tolko dlya administratorov)
// @Tags Kinoseansy
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID kinoseansa"
// @Param movie_show body dto.UpdateMovieShowRequest true "Novye dannye kinoseansa"
// @Success 200 "Kinoseans obnovlen"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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

// @Summary Udalit kinoseans
// @Description Udalyaet kinoseans po identifikatoru (tolko dlya administratorov)
// @Tags Kinoseansy
// @Security BearerAuth
// @Param id path string true "UUID kinoseansa"
// @Success 204 "Kinoseans udalen"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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
