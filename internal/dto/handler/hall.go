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

// @Summary Get a list of halls
// @Description Returns a paginated list of all halls with filtering
// @Tags Halls
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param name query string false "Search by hall name (case-insensitive substring search)"
// @Param screen_type_id query string false "Filter by screen type ID"
// @Param description query string false "Search by hall description (case-insensitive substring search)"
// @Success 200 {object} dto.PaginatedHallResponse "Hall list"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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

// @Summary Get a hall by ID
// @Description Returns information about a hall by its identifier
// @Tags Halls
// @Produce json
// @Param id path string true "Hall UUID"
// @Success 200 {object} dto.HallResponse "Hall information"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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

// @Summary Create a hall
// @Description Creates a new hall (administrators only)
// @Tags Halls
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hall body dto.CreateHallRequest true "Hall data"
// @Success 201 {object} dto.CreateResponse "Hall created"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
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

// @Summary Update a hall
// @Description Updates information about a hall (administrators only)
// @Tags Halls
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Hall UUID"
// @Param hall body dto.UpdateHallRequest true "New hall data"
// @Success 200 "Hall updated"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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

// @Summary Delete a hall
// @Description Deletes a hall by identifier (administrators only)
// @Tags Halls
// @Security BearerAuth
// @Param id path string true "Hall UUID"
// @Success 204 "Hall deleted"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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
