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

// @Summary Poluchit spisok zhanrov
// @Description Vozvraschaet paginirovannyy spisok vsekh zhanrov s filtratsiey
// @Tags Zhanry
// @Produce json
// @Param page query int false "Nomer stranitsy" default(1) minimum(1)
// @Param limit query int false "Kolichestvo elementov na stranitse" default(20) minimum(1) maximum(100)
// @Param name query string false "Poisk po nazvaniyu zhanra (registronezavisimyy poisk vkhozhdeniy)"
// @Param description query string false "Poisk po opisaniyu zhanra (registronezavisimyy poisk vkhozhdeniy)"
// @Success 200 {object} dto.PaginatedGenreResponse "Spisok zhanrov"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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

// @Summary Poluchit zhanr po ID
// @Description Vozvraschaet informatsiyu o zhanre po ego identifikatoru
// @Tags Zhanry
// @Produce json
// @Param id path string true "UUID zhanra"
// @Success 200 {object} dto.GenreResponse "Informatsiya o zhanre"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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

// @Summary Sozdat zhanr
// @Description Sozdaet novyy zhanr (tolko dlya administratorov)
// @Tags Zhanry
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param genre body dto.CreateGenreRequest true "Dannye zhanra"
// @Success 201 {object} dto.CreateResponse "Zhanr sozdan"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
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

// @Summary Obnovit zhanr
// @Description Obnovlyaet informatsiyu o zhanre (tolko dlya administratorov)
// @Tags Zhanry
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID zhanra"
// @Param genre body dto.UpdateGenreRequest true "Novye dannye zhanra"
// @Success 200 "Zhanr obnovlen"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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

// @Summary Udalit zhanr
// @Description Udalyaet zhanr po identifikatoru (tolko dlya administratorov)
// @Tags Zhanry
// @Security BearerAuth
// @Param id path string true "UUID zhanra"
// @Success 204 "Zhanr udalen"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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
