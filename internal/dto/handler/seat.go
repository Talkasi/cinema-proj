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

// @Summary Получить места в зале
// @Description Возвращает пагинированный список мест в указанном зале с фильтрацией
// @Tags Места
// @Produce json
// @Param hall_id path string true "UUID зала"
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Количество элементов на странице" default(20) minimum(1) maximum(100)
// @Param seat_type_id query string false "Фильтр по ID типа места"
// @Param row_number_min query int false "Минимальный номер ряда" minimum(1)
// @Param row_number_max query int false "Максимальный номер ряда" minimum(1)
// @Param seat_number_min query int false "Минимальный номер места" minimum(1)
// @Param seat_number_max query int false "Максимальный номер места" minimum(1)
// @Success 200 {object} dto.PaginatedSeatResponse "Список мест"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /halls/{hall_id}/seats [get]
func (s *SeatHandler) GetSeatsByHall(w http.ResponseWriter, r *http.Request) {
	hallId := chi.URLParam(r, "hall_id")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	dtoFilters := dto.SeatFilters{
		SeatTypeID: r.URL.Query().Get("seat_type_id"),
	}

	// Парсинг числовых параметров
	if rowMin := r.URL.Query().Get("row_number_min"); rowMin != "" {
		if val, err := strconv.Atoi(rowMin); err == nil {
			dtoFilters.RowNumberMin = val
		}
	}

	if rowMax := r.URL.Query().Get("row_number_max"); rowMax != "" {
		if val, err := strconv.Atoi(rowMax); err == nil {
			dtoFilters.RowNumberMax = val
		}
	}

	if seatMin := r.URL.Query().Get("seat_number_min"); seatMin != "" {
		if val, err := strconv.Atoi(seatMin); err == nil {
			dtoFilters.SeatNumberMin = val
		}
	}

	if seatMax := r.URL.Query().Get("seat_number_max"); seatMax != "" {
		if val, err := strconv.Atoi(seatMax); err == nil {
			dtoFilters.SeatNumberMax = val
		}
	}

	domainFilters := domain.SeatFiltersFromDTO(dtoFilters)

	result, err := s.seatService.GetByHall(r.Context(), hallId, domainFilters, page, limit)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// @Summary Получить место по ID
// @Description Возвращает информацию о месте по его идентификатору
// @Tags Места
// @Produce json
// @Param seat_id path string true "UUID места"
// @Success 200 {object} dto.SeatResponse "Информация о месте"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /seats/{seat_id} [get]
func (s *SeatHandler) GetSeatByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "seat_id")
	seat, err := s.seatService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(seat)
}

// @Summary Создать место
// @Description Создает новое место (только для администраторов)
// @Tags Места
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hall_id path string true "UUID зала"
// @Param seat body dto.CreateSeatRequest true "Данные места"
// @Success 201 {object} dto.CreateResponse "Место создано"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Router /halls/{hall_id}/seats [post]
func (s *SeatHandler) CreateSeat(w http.ResponseWriter, r *http.Request) {
	hallID := chi.URLParam(r, "hall_id")

	var req dto.CreateSeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
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
	json.NewEncoder(w).Encode(dto.CreateResponse{ID: result.ID})
}

// @Summary Обновить место
// @Description Полностью обновляет информацию о месте (только для администраторов)
// @Tags Места
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param seat_id path string true "UUID места"
// @Param seat body dto.UpdateSeatRequest true "Новые данные места"
// @Success 200 "Место обновлено"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /seats/{seat_id} [put]
func (s *SeatHandler) UpdateSeat(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "seat_id")

	var req dto.UpdateSeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainSeat := domain.UpdateSeatFromDTO(req)

	_, err := s.seatService.Update(r.Context(), id, domainSeat)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Удалить место
// @Description Удаляет место по идентификатору (только для администраторов)
// @Tags Места
// @Security BearerAuth
// @Param seat_id path string true "UUID места"
// @Success 204 "Место удалено"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
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
