package handler

import (
	domain "cw/internal/domain/models"
	"cw/internal/domain/service"
	dto "cw/internal/dto/models"
	"cw/internal/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type GenreHandler struct {
	genreService *service.GenreService
}

func NewGenreHandler(gs *service.GenreService) *GenreHandler {
	return &GenreHandler{genreService: gs}
}

// @Summary Get a list of genres
// @Description Returns a paginated list of all genres with filtering
// @Tags Genres
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param name query string false "Search by genre name (case-insensitive substring search)"
// @Param description query string false "Search by genre description (case-insensitive substring search)"
// @Success 200 {object} dto.PaginatedGenreResponse "Genre list"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Failure 500 {object} dto.ErrorResponse "Server error"
// @Router /genres [get]
func (g *GenreHandler) GetGenres(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePaginationParams(r)

	dtoFilters := dto.GenreFilters{
		Name:        r.URL.Query().Get("name"),
		Description: r.URL.Query().Get("description"),
	}

	domainFilters := domain.GenreFiltersFromDTO(dtoFilters)

	result, err := g.genreService.GetAll(r.Context(), domainFilters, page, limit)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, result); err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Get a genre by ID
// @Description Returns information about a genre by its identifier
// @Tags Genres
// @Produce json
// @Param id path string true "Genre UUID"
// @Success 200 {object} dto.GenreResponse "Genre information"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /genres/{id} [get]
func (g *GenreHandler) GetGenreByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	genre, err := g.genreService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, genre); err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Create a genre
// @Description Creates a new genre (administrators only)
// @Tags Genres
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param genre body dto.CreateGenreRequest true "Genre data"
// @Success 201 {object} dto.CreateResponse "Genre created"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Router /genres [post]
func (g *GenreHandler) CreateGenre(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateGenreRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainGenre := domain.CreateGenreFromDTO(req)

	result, err := g.genreService.Create(r.Context(), domainGenre)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	if err := writeJSON(w, http.StatusCreated, dto.CreateResponse{ID: result.ID}); err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Update a genre
// @Description Updates information about a genre (administrators only)
// @Tags Genres
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Genre UUID"
// @Param genre body dto.UpdateGenreRequest true "New genre data"
// @Success 200 "Genre updated"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /genres/{id} [put]
func (g *GenreHandler) UpdateGenre(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateGenreRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainGenre := domain.UpdateGenreFromDTO(req)

	_, err := g.genreService.Update(r.Context(), id, domainGenre)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Delete a genre
// @Description Deletes a genre by identifier (administrators only)
// @Tags Genres
// @Security BearerAuth
// @Param id path string true "Genre UUID"
// @Success 204 "Genre deleted"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /genres/{id} [delete]
func (g *GenreHandler) DeleteGenre(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := g.genreService.Delete(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (g *GenreHandler) RegisterRoutes(r chi.Router) {
	r.Route("/genres", func(r chi.Router) {
		r.Get("/", g.GetGenres)
		r.Post("/", g.CreateGenre)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", g.GetGenreByID)
			r.Put("/", g.UpdateGenre)
			r.Delete("/", g.DeleteGenre)
		})
	})
}
