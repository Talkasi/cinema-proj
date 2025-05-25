package handler

import (
	"cw/internal/dto"
	"cw/internal/service"
	"cw/internal/utils"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type SeatTypeHandler struct {
	seatTypeService service.SeatTypeService
}

func NewSeatTypeHandler(sts service.SeatTypeService) *SeatTypeHandler {
	return &SeatTypeHandler{seatTypeService: sts}
}

// @Summary Получить все типы мест (guest | user | admin)
// @Description Возвращает список всех типов мест, содержащихся в базе данных.
// @Tags Типы мест
// @Produce json
// @Success 200 {array} SeatType "Список типов мест"
// @Failure 404 {object} ErrorResponse "Типы мест не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /seat-types [get]
func (st *SeatTypeHandler) GetSeatTypes(w http.ResponseWriter, r *http.Request) {
	seatType, err := st.seatTypeService.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(seatType)
}

// @Summary Получить тип места по ID (guest | user | admin)
// @Description Возвращает тип места по ID.
// @Tags Типы мест
// @Produce json
// @Param id path string true "ID типа места"
// @Success 200 {object} SeatType "Тип места"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 404 {object} ErrorResponse "Тип места не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /seat-types/{id} [get]
func (st *SeatTypeHandler) GetSeatTypeByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	seatType, err := st.seatTypeService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(dto.SeatTypeFromDomain(seatType))
}

// @Summary Создать тип места (admin)
// @Description Создаёт новый тип места.
// @Tags Типы мест
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param seat_type body SeatTypeAdmin true "Данные типа места"
// @Success 201 {object} CreateResponse "ID созданного типа места"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /seat-types [post]
func (st *SeatTypeHandler) CreateSeatType(w http.ResponseWriter, r *http.Request) {
	var data dto.SeatTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	seatType, err := st.seatTypeService.Create(r.Context(), data.ToDomain())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(seatType.ID)
}

// @Summary Обновить тип места (admin)
// @Description Обновляет существующий тип места.
// @Tags Типы мест
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID типа места"
// @Param seat_type body SeatTypeAdmin true "Обновлённые данные типа места"
// @Success 200 "Данные о типе места успешно обновлены"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Тип места не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /seat-types/{id} [put]
func (st *SeatTypeHandler) UpdateSeatType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var data dto.SeatTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	_, err := st.seatTypeService.Update(r.Context(), id, data.ToDomain())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w)
}

// @Summary Удалить тип места (admin)
// @Description Удаляет тип места по ID.
// @Tags Типы мест
// @Param id path string true "ID типа места"
// @Security BearerAuth
// @Success 204 "Данные о типе места успешно удалены"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Тип места не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /seat-types/{id} [delete]
func (st *SeatTypeHandler) DeleteSeatType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := st.seatTypeService.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Поиск типов места по названию (guest | user | admin)
// @Description Возвращает типы места, название которых содержит указанную строку.
// @Tags Типы мест
// @Produce json
// @Param query query string true "Поисковый запрос"
// @Success 200 {array} SeatType "Список типов мест"
// @Failure 400 {object} ErrorResponse "Строка поиска пуста"
// @Failure 404 {object} ErrorResponse "Типы мест не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /seat-types/search [get]
func (st *SeatTypeHandler) SearchSeatTypes(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	query = utils.PrepareString(query)
	if query == "" {
		http.Error(w, "Строка поиска пуста", http.StatusBadRequest)
		return
	}

	seatTypes, err := st.seatTypeService.SearchByName(r.Context(), query)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.SeatTypeFromDomainList(seatTypes))
}
