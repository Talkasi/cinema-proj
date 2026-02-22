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

// @Summary Get a list of seat types
// @Description Returns a paginated list of all seat types with filtering
// @Tags Seat Types
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param name query string false "Search by seat type name (case-insensitive substring search)"
// @Param description query string false "Search by seat type description (case-insensitive substring search)"
// @Success 200 {object} dto.PaginatedSeatTypeResponse "Seat type list"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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

// @Summary Get a seat type by ID
// @Description Returns information about a seat type by its identifier
// @Tags Seat Types
// @Produce json
// @Param id path string true "Seat type UUID"
// @Success 200 {object} dto.SeatTypeResponse "Seat type information"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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

// @Summary Create a seat type
// @Description Creates a new seat type (administrators only)
// @Tags Seat Types
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param seat_type body dto.CreateSeatTypeRequest true "Seat type data"
// @Success 201 {object} dto.CreateResponse "Seat type created"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
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

// @Summary Update a seat type
// @Description Fully updates information about a seat type (administrators only)
// @Tags Seat Types
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Seat type UUID"
// @Param seat_type body dto.UpdateSeatTypeRequest true "New seat type data"
// @Success 200 "Seat type updated"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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

// @Summary Delete a seat type
// @Description Deletes a seat type by identifier (administrators only)
// @Tags Seat Types
// @Security BearerAuth
// @Param id path string true "Seat type UUID"
// @Success 204 "Seat type deleted"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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
