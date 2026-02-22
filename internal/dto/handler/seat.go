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

type SeatHandler struct {
	seatService *service.SeatService
}

func NewSeatHandler(st *service.SeatService) *SeatHandler {
	return &SeatHandler{seatService: st}
}

// @Summary Get seats in a hall
// @Description Returns a paginated list of seats in the specified hall with filtering
// @Tags Seats
// @Produce json
// @Param hall_id path string true "Hall UUID"
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param seat_type_id query string false "Filter by seat type ID"
// @Param row_number_min query int false "Minimum row number" minimum(1)
// @Param row_number_max query int false "Maximum row number" minimum(1)
// @Param seat_number_min query int false "Minimum seat number" minimum(1)
// @Param seat_number_max query int false "Maximum seat number" minimum(1)
// @Success 200 {object} dto.PaginatedSeatResponse "Seat list"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /halls/{hall_id}/seats [get]
func (s *SeatHandler) GetSeatsByHall(w http.ResponseWriter, r *http.Request) {
	hallId := chi.URLParam(r, "hall_id")

	page := s.getPageParam(r)
	limit := s.getLimitParam(r)
	dtoFilters := s.parseSeatFilters(r)

	domainFilters := domain.SeatFiltersFromDTO(dtoFilters)

	result, err := s.seatService.GetByHall(r.Context(), hallId, domainFilters, page, limit)
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

// getPageParam extracts and validates the page parameter
func (s *SeatHandler) getPageParam(r *http.Request) int {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	return page
}

// getLimitParam extracts and validates the limit parameter
func (s *SeatHandler) getLimitParam(r *http.Request) int {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return limit
}

// parseSeatFilters extracts seat filters from the request
func (s *SeatHandler) parseSeatFilters(r *http.Request) dto.SeatFilters {
	dtoFilters := dto.SeatFilters{
		SeatTypeID: r.URL.Query().Get("seat_type_id"),
	}

	dtoFilters.RowNumberMin = s.parseIntParam(r, "row_number_min")
	dtoFilters.RowNumberMax = s.parseIntParam(r, "row_number_max")
	dtoFilters.SeatNumberMin = s.parseIntParam(r, "seat_number_min")
	dtoFilters.SeatNumberMax = s.parseIntParam(r, "seat_number_max")

	return dtoFilters
}

// parseIntParam safely parses an integer parameter from the request
func (s *SeatHandler) parseIntParam(r *http.Request, paramName string) int {
	if paramValue := r.URL.Query().Get(paramName); paramValue != "" {
		if val, err := strconv.Atoi(paramValue); err == nil {
			return val
		}
	}
	return 0
}

// @Summary Get a seat by ID
// @Description Returns information about a seat by its identifier
// @Tags Seats
// @Produce json
// @Param seat_id path string true "Seat UUID"
// @Success 200 {object} dto.SeatResponse "Seat information"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /seats/{seat_id} [get]
func (s *SeatHandler) GetSeatByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "seat_id")
	seat, err := s.seatService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(seat); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Create a seat
// @Description Creates a new seat (administrators only)
// @Tags Seats
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hall_id path string true "Hall UUID"
// @Param seat body dto.CreateSeatRequest true "Seat data"
// @Success 201 {object} dto.CreateResponse "Seat created"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Router /halls/{hall_id}/seats [post]
func (s *SeatHandler) CreateSeat(w http.ResponseWriter, r *http.Request) {
	hallID := chi.URLParam(r, "hall_id")

	var req dto.CreateSeatRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	req.HallID = hallID

	domainSeat := domain.CreateSeatFromDTO(req)

	result, err := s.seatService.Create(r.Context(), domainSeat)
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

// @Summary Update a seat
// @Description Fully updates information about a seat (administrators only)
// @Tags Seats
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param seat_id path string true "Seat UUID"
// @Param seat body dto.UpdateSeatRequest true "New seat data"
// @Success 200 "Seat updated"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /seats/{seat_id} [put]
func (s *SeatHandler) UpdateSeat(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "seat_id")

	var req dto.UpdateSeatRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainSeat := domain.UpdateSeatFromDTO(req)

	_, err := s.seatService.Update(r.Context(), id, domainSeat)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Delete a seat
// @Description Deletes a seat by identifier (administrators only)
// @Tags Seats
// @Security BearerAuth
// @Param seat_id path string true "Seat UUID"
// @Success 204 "Seat deleted"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /seats/{seat_id} [delete]
func (s *SeatHandler) DeleteSeat(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "seat_id")
	err := s.seatService.Delete(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *SeatHandler) RegisterRoutes(r chi.Router) {
	r.Route("/seats/{seat_id}", func(r chi.Router) {
		r.Get("/", s.GetSeatByID)
		r.Put("/", s.UpdateSeat)
		r.Delete("/", s.DeleteSeat)
	})

	r.Route("/halls/{hall_id}/seats", func(r chi.Router) {
		r.Get("/", s.GetSeatsByHall)
		r.Post("/", s.CreateSeat)
	})
}
