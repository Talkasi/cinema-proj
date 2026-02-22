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

type ScreenTypeHandler struct {
	screenTypeService *service.ScreenTypeService
}

func NewScreenTypeHandler(sts *service.ScreenTypeService) *ScreenTypeHandler {
	return &ScreenTypeHandler{screenTypeService: sts}
}

// @Summary Poluchit spisok tipov ekranov
// @Description Vozvraschaet paginirovannyy spisok vsekh tipov ekranov s filtratsiey
// @Tags Tipy ekranov
// @Produce json
// @Param page query int false "Nomer stranitsy" default(1) minimum(1)
// @Param limit query int false "Kolichestvo elementov na stranitse" default(20) minimum(1) maximum(100)
// @Param name query string false "Poisk po nazvaniyu tipa ekrana (registronezavisimyy poisk vkhozhdeniy)"
// @Param description query string false "Poisk po opisaniyu tipa ekrana (registronezavisimyy poisk vkhozhdeniy)"
// @Success 200 {object} dto.PaginatedScreenTypeResponse "Spisok tipov ekranov"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /screen-types [get]
func (st *ScreenTypeHandler) GetScreenTypes(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	dtoFilters := dto.ScreenTypeFilters{
		Name:        r.URL.Query().Get("name"),
		Description: r.URL.Query().Get("description"),
	}

	domainFilters := domain.ScreenTypeFiltersFromDTO(dtoFilters)

	result, err := st.screenTypeService.GetAll(r.Context(), domainFilters, page, limit)
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

// @Summary Poluchit tip ekrana po ID
// @Description Vozvraschaet informatsiyu o tipe ekrana po ego identifikatoru
// @Tags Tipy ekranov
// @Produce json
// @Param id path string true "UUID tipa ekrana"
// @Success 200 {object} dto.ScreenTypeResponse "Informatsiya o tipe ekrana"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /screen-types/{id} [get]
func (st *ScreenTypeHandler) GetScreenTypeByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	screenType, err := st.screenTypeService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(screenType); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Sozdat tip ekrana
// @Description Sozdaet novyy tip ekrana (tolko dlya administratorov)
// @Tags Tipy ekranov
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param screen_type body dto.CreateScreenTypeRequest true "Dannye tipa ekrana"
// @Success 201 {object} dto.CreateResponse "Tip ekrana sozdan"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Router /screen-types [post]
func (st *ScreenTypeHandler) CreateScreenType(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateScreenTypeRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainScreenType := domain.CreateScreenTypeFromDTO(req)

	result, err := st.screenTypeService.Create(r.Context(), domainScreenType)
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

// @Summary Obnovit tip ekrana
// @Description Polnostyu obnovlyaet informatsiyu o tipe ekrana (tolko dlya administratorov)
// @Tags Tipy ekranov
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID tipa ekrana"
// @Param screen_type body dto.UpdateScreenTypeRequest true "Novye dannye tipa ekrana"
// @Success 200 "Tip ekrana obnovlen"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /screen-types/{id} [put]
func (st *ScreenTypeHandler) UpdateScreenType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateScreenTypeRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainScreenType := domain.UpdateScreenTypeFromDTO(req)

	_, err := st.screenTypeService.Update(r.Context(), id, domainScreenType)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Udalit tip ekrana
// @Description Udalyaet tip ekrana po identifikatoru (tolko dlya administratorov)
// @Tags Tipy ekranov
// @Security BearerAuth
// @Param id path string true "UUID tipa ekrana"
// @Success 204 "Tip ekrana udalen"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /screen-types/{id} [delete]
func (st *ScreenTypeHandler) DeleteScreenType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := st.screenTypeService.Delete(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (st *ScreenTypeHandler) RegisterRoutes(r chi.Router) {
	r.Route("/screen-types", func(r chi.Router) {
		r.Get("/", st.GetScreenTypes)
		r.Post("/", st.CreateScreenType)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", st.GetScreenTypeByID)
			r.Put("/", st.UpdateScreenType)
			r.Delete("/", st.DeleteScreenType)
		})
	})
}
