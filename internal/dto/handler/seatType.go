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

type SeatTypeHandler struct {
	seatTypeService *service.SeatTypeService
}

func NewSeatTypeHandler(sts *service.SeatTypeService) *SeatTypeHandler {
	return &SeatTypeHandler{seatTypeService: sts}
}

// @Summary Poluchit spisok tipov mest
// @Description Vozvraschaet paginirovannyy spisok vsekh tipov mest s filtratsiey
// @Tags Tipy mest
// @Produce json
// @Param page query int false "Nomer stranitsy" default(1) minimum(1)
// @Param limit query int false "Kolichestvo elementov na stranitse" default(20) minimum(1) maximum(100)
// @Param name query string false "Poisk po nazvaniyu tipa mesta (registronezavisimyy poisk vkhozhdeniy)"
// @Param description query string false "Poisk po opisaniyu tipa mesta (registronezavisimyy poisk vkhozhdeniy)"
// @Success 200 {object} dto.PaginatedSeatTypeResponse "Spisok tipov mest"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /seat-types [get]
func (st *SeatTypeHandler) GetSeatTypes(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	dtoFilters := dto.SeatTypeFilters{
		Name:        r.URL.Query().Get("name"),
		Description: r.URL.Query().Get("description"),
	}

	domainFilters := domain.SeatTypeFiltersFromDTO(dtoFilters)

	result, err := st.seatTypeService.GetAll(r.Context(), domainFilters, page, limit)
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

// @Summary Poluchit tip mesta po ID
// @Description Vozvraschaet informatsiyu o tipe mesta po ego identifikatoru
// @Tags Tipy mest
// @Produce json
// @Param id path string true "UUID tipa mesta"
// @Success 200 {object} dto.SeatTypeResponse "Informatsiya o tipe mesta"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /seat-types/{id} [get]
func (st *SeatTypeHandler) GetSeatTypeByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	seatType, err := st.seatTypeService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(seatType); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Sozdat tip mesta
// @Description Sozdaet novyy tip mesta (tolko dlya administratorov)
// @Tags Tipy mest
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param seat_type body dto.CreateSeatTypeRequest true "Dannye tipa mesta"
// @Success 201 {object} dto.CreateResponse "Tip mesta sozdan"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Router /seat-types [post]
func (st *SeatTypeHandler) CreateSeatType(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSeatTypeRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainSeatType := domain.CreateSeatTypeFromDTO(req)

	result, err := st.seatTypeService.Create(r.Context(), domainSeatType)
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

// @Summary Obnovit tip mesta
// @Description Polnostyu obnovlyaet informatsiyu o tipe mesta (tolko dlya administratorov)
// @Tags Tipy mest
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID tipa mesta"
// @Param seat_type body dto.UpdateSeatTypeRequest true "Novye dannye tipa mesta"
// @Success 200 "Tip mesta obnovlen"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /seat-types/{id} [put]
func (st *SeatTypeHandler) UpdateSeatType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateSeatTypeRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainSeatType := domain.UpdateSeatTypeFromDTO(req)

	_, err := st.seatTypeService.Update(r.Context(), id, domainSeatType)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Udalit tip mesta
// @Description Udalyaet tip mesta po identifikatoru (tolko dlya administratorov)
// @Tags Tipy mest
// @Security BearerAuth
// @Param id path string true "UUID tipa mesta"
// @Success 204 "Tip mesta udalen"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
// @Router /seat-types/{id} [delete]
func (st *SeatTypeHandler) DeleteSeatType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := st.seatTypeService.Delete(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (st *SeatTypeHandler) RegisterRoutes(r chi.Router) {
	r.Route("/seat-types", func(r chi.Router) {
		r.Get("/", st.GetSeatTypes)
		r.Post("/", st.CreateSeatType)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", st.GetSeatTypeByID)
			r.Put("/", st.UpdateSeatType)
			r.Delete("/", st.DeleteSeatType)
		})
	})
}
