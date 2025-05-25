package handler

import (
	"cw/internal/dto"
	"cw/internal/service"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type SeatHandler struct {
	seatService service.SeatService
}

func NewSeatHandler(st service.SeatService) *SeatHandler {
	return &SeatHandler{seatService: st}
}

// @Summary Получить все места (admin)
// @Description Возвращает список всех мест, содержащихся в базе данных.
// @Tags Места
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Seat "Список мест"
// @Failure 403 {string} string "Доступ запрещён"
// @Failure 404 {string} string "Места не найдены"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seats [get]
func (s *SeatHandler) GetSeats(w http.ResponseWriter, r *http.Request) {
	seat, err := s.seatService.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(seat)
}

// @Summary Получить место по ID (guest | user | admin)
// @Description Возвращает место по ID.
// @Tags Места
// @Produce json
// @Param id path string true "ID места"
// @Success 200 {object} Seat "Место"
// @Failure 400 {string} string "Неверный формат ID"
// @Failure 404 {string} string "Место не найдено"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seats/{id} [get]
func (s *SeatHandler) GetSeatByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	seat, err := s.seatService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(dto.SeatFromDomain(seat))
}

// @Summary Создать место (admin)
// @Description Создаёт новое место.
// @Tags Места
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param seat body SeatData true "Данные места"
// @Success 201 {object} CreateResponse "ID созданного места"
// @Failure 400 {string} string "Некорректные данные"
// @Failure 403 {string} string "Доступ запрещён"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seats [post]
func (s *SeatHandler) CreateSeat(w http.ResponseWriter, r *http.Request) {
	var data dto.SeatRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	seat, err := s.seatService.Create(r.Context(), data.ToDomain())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(seat.ID)
}

// @Summary Обновить место (admin)
// @Description Обновляет существующее место.
// @Tags Места
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID места"
// @Param seat body SeatData true "Обновлённые данные места"
// @Success 200 "Данные о месте успешно обновлены"
// @Failure 400 {string} string "Некорректные данные"
// @Failure 403 {string} string "Доступ запрещён"
// @Failure 404 {string} string "Место не найдено"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seats/{id} [put]
func (s *SeatHandler) UpdateSeat(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var data dto.SeatRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	_, err := s.seatService.Update(r.Context(), id, data.ToDomain())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w)
}

// @Summary Удалить место (admin)
// @Description Удаляет место по ID.
// @Tags Места
// @Param id path string true "ID места"
// @Security BearerAuth
// @Success 204 "Данные о месте успешно удалены"
// @Failure 400 {string} string "Неверный формат ID"
// @Failure 403 {string} string "Доступ запрещён"
// @Failure 404 {string} string "Место не найдено"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seats/{id} [delete]
func (s *SeatHandler) DeleteSeat(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := s.seatService.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Получить места по ID зала (guest | user | admin)
// @Description Возвращает список мест в указанном зале.
// @Tags Места
// @Produce json
// @Param hall_id path string true "ID зала"
// @Success 200 {array} Seat "Список мест"
// @Failure 400 {string} string "Неверный формат ID зала"
// @Failure 404 {string} string "Места не найдены"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /halls/{hall_id}/seats [get]
func (s *SeatHandler) GetSeatsByHallID(w http.ResponseWriter, r *http.Request) {
	hallId := chi.URLParam(r, "hall_id")
	seat, err := s.seatService.GetByHall(r.Context(), hallId)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.SeatFromDomainList(seat))
}
