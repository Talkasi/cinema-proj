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

// @Summary Poluchit mesta v zale
// @Description Vozvraschaet paginirovannyy spisok mest v ukazannom zale s filtratsiey
// @Tags Mesta
// @Produce json
// @Param hall_id path string true "UUID zala"
// @Param page query int false "Nomer stranitsy" default(1) minimum(1)
// @Param limit query int false "Kolichestvo elementov na stranitse" default(20) minimum(1) maximum(100)
// @Param seat_type_id query string false "Filtr po ID tipa mesta"
// @Param row_number_min query int false "Minimalnyy nomer ryada" minimum(1)
// @Param row_number_max query int false "Maksimalnyy nomer ryada" minimum(1)
// @Param seat_number_min query int false "Minimalnyy nomer mesta" minimum(1)
// @Param seat_number_max query int false "Maksimalnyy nomer mesta" minimum(1)
// @Success 200 {object} dto.PaginatedSeatResponse "Spisok mest"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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

// @Summary Poluchit mesto po ID
// @Description Vozvraschaet informatsiyu o meste po ego identifikatoru
// @Tags Mesta
// @Produce json
// @Param seat_id path string true "UUID mesta"
// @Success 200 {object} dto.SeatResponse "Informatsiya o meste"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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

// @Summary Sozdat mesto
// @Description Sozdaet novoe mesto (tolko dlya administratorov)
// @Tags Mesta
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hall_id path string true "UUID zala"
// @Param seat body dto.CreateSeatRequest true "Dannye mesta"
// @Success 201 {object} dto.CreateResponse "Mesto sozdano"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
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

// @Summary Obnovit mesto
// @Description Polnostyu obnovlyaet informatsiyu o meste (tolko dlya administratorov)
// @Tags Mesta
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param seat_id path string true "UUID mesta"
// @Param seat body dto.UpdateSeatRequest true "Novye dannye mesta"
// @Success 200 "Mesto obnovleno"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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

// @Summary Udalit mesto
// @Description Udalyaet mesto po identifikatoru (tolko dlya administratorov)
// @Tags Mesta
// @Security BearerAuth
// @Param seat_id path string true "UUID mesta"
// @Success 204 "Mesto udaleno"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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
