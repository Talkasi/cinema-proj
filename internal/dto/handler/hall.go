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

type HallHandler struct {
	hallService *service.HallService
}

func NewHallHandler(hs *service.HallService) *HallHandler {
	return &HallHandler{hallService: hs}
}

// @Summary Poluchit spisok zalov
// @Description Vozvraschaet paginirovannyy spisok vsekh kinozalov s filtratsiey
// @Tags Kinozaly
// @Produce json
// @Param page query int false "Nomer stranitsy" default(1) minimum(1)
// @Param limit query int false "Kolichestvo elementov na stranitse" default(20) minimum(1) maximum(100)
// @Param name query string false "Poisk po nazvaniyu zala (registronezavisimyy poisk vkhozhdeniy)"
// @Param screen_type_id query string false "Filtr po ID tipa ekrana"
// @Param description query string false "Poisk po opisaniyu zala (registronezavisimyy poisk vkhozhdeniy)"
// @Success 200 {object} dto.PaginatedHallResponse "Spisok zalov"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /halls [get]
func (h *HallHandler) GetHalls(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	dtoFilters := dto.HallFilters{
		Name:         r.URL.Query().Get("name"),
		ScreenTypeID: r.URL.Query().Get("screen_type_id"),
		Description:  r.URL.Query().Get("description"),
	}

	domainFilters := domain.HallFiltersFromDTO(dtoFilters)

	result, err := h.hallService.GetAll(r.Context(), domainFilters, page, limit)
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

// @Summary Poluchit zal po ID
// @Description Vozvraschaet informatsiyu o kinozale po ego identifikatoru
// @Tags Kinozaly
// @Produce json
// @Param id path string true "UUID zala"
// @Success 200 {object} dto.HallResponse "Informatsiya o zale"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /halls/{id} [get]
func (h *HallHandler) GetHallByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	hall, err := h.hallService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(hall); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Sozdat zal
// @Description Sozdaet novyy kinozal (tolko dlya administratorov)
// @Tags Kinozaly
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hall body dto.CreateHallRequest true "Dannye zala"
// @Success 201 {object} dto.CreateResponse "Zal sozdan"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Router /halls [post]
func (h *HallHandler) CreateHall(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateHallRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainHall := domain.CreateHallFromDTO(req)

	result, err := h.hallService.Create(r.Context(), domainHall)
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

// @Summary Obnovit zal
// @Description Obnovlyaet informatsiyu o kinozale (tolko dlya administratorov)
// @Tags Kinozaly
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID zala"
// @Param hall body dto.UpdateHallRequest true "Novye dannye zala"
// @Success 200 "Zal obnovlen"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /halls/{id} [put]
func (h *HallHandler) UpdateHall(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateHallRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainHall := domain.UpdateHallFromDTO(req)

	_, err := h.hallService.Update(r.Context(), id, domainHall)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Udalit zal
// @Description Udalyaet kinozal po identifikatoru (tolko dlya administratorov)
// @Tags Kinozaly
// @Security BearerAuth
// @Param id path string true "UUID zala"
// @Success 204 "Zal udalen"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /halls/{id} [delete]
func (h *HallHandler) DeleteHall(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.hallService.Delete(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HallHandler) RegisterRoutes(r chi.Router) {
	r.Route("/halls", func(r chi.Router) {
		r.Get("/", h.GetHalls)
		r.Post("/", h.CreateHall)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetHallByID)
			r.Put("/", h.UpdateHall)
			r.Delete("/", h.DeleteHall)
		})
	})
}
