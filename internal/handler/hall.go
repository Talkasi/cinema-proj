package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"cw/internal/models"
	"cw/internal/service"
)

type HallHandler struct {
	hallService service.HallService
}

func NewHallHandler(hs service.HallService) *HallHandler {
	return &HallHandler{hallService: hs}
}

// @Summary Получить все кинозалы (guest | user | admin)
// @Description Возвращает список всех кинозалов, содержащихся в базе данных.
// @Tags Кинозалы
// @Produce json
// @Success 200 {array} Hall "Список кинозалов"
// @Failure 404 {object} ErrorResponse "Данные не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /halls [get]
func (h *HallHandler) GetHalls(w http.ResponseWriter, r *http.Request) {
	halls, err := h.hallService.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Ошибка при получении залов", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(halls)
}

// @Summary Получить кинозал по ID (guest | user | admin)
// @Description Возвращает кинозал по ID.
// @Tags Кинозалы
// @Produce json
// @Param id path string true "ID зала"
// @Success 200 {object} Hall "Данные кинозала"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 404 {object} ErrorResponse "Данные не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /halls/{id} [get]
func (h *HallHandler) GetHallByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	hall, err := h.hallService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Зал не найден", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(hall)
}

// @Summary Создать кинозал (admin)
// @Description Создаёт новый кинозал.
// @Tags Кинозалы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hall body HallData true "Данные кинозала"
// @Success 201 {object} CreateResponse "ID созданного кинозала"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 409 {object} ErrorResponse "Конфликт при создании кинозала"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /halls [post]
func (h *HallHandler) CreateHall(w http.ResponseWriter, r *http.Request) {
	var data models.HallData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	hall, err := h.hallService.Create(r.Context(), data)
	if err != nil {
		http.Error(w, "Ошибка при создании зала", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(hall)
}

// @Summary Обновить кинозал (admin)
// @Description Обновляет существующий кинозал.
// @Tags Кинозалы
// @Accept json
// @Produce json
// @Param id path string true "ID зала"
// @Param hall body HallData true "Обновлённые данные зала"
// @Security BearerAuth
// @Success 200 "Данные о кинозале успешно обновлены"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Зал не найден"
// @Failure 409 {object} ErrorResponse "Конфликт при обновлении кинозала"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /halls/{id} [put]
func (h *HallHandler) UpdateHall(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var data models.HallData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	hall, err := h.hallService.Update(r.Context(), id, data)
	if err != nil {
		http.Error(w, "Зал не найден", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(hall)
}

// @Summary Удалить кинозал (admin)
// @Description Удаляет кинозал по ID.
// @Tags Кинозалы
// @Param id path string true "ID кинозала"
// @Security BearerAuth
// @Success 204 "Данные о кинозале успешно удалены"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Данные не найдены"
// @Failure 409 {object} ErrorResponse "Конфликт при удалении кинозала"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /halls/{id} [delete]
func (h *HallHandler) DeleteHall(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.hallService.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, "Зал не найден", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Получить залы по типу экрана (guest | user | admin)
// @Description Возвращает список залов с указанным типом экрана.
// @Tags Кинозалы
// @Produce json
// @Param screen_type_id query string true "ID типа экрана"
// @Success 200 {array} Hall "Список найденных кинозалов"
// @Failure 400 {object} ErrorResponse "Неверный формат ID типа экрана"
// @Failure 404 {object} ErrorResponse "Залы не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /halls/by-screen-type [get]
func (h *HallHandler) GetHallsByScreenType(w http.ResponseWriter, r *http.Request) {
	screenTypeID := r.URL.Query().Get("screen_type_id")
	if _, err := uuid.Parse(screenTypeID); err != nil {
		http.Error(w, "Неверный формат ID типа экрана", http.StatusBadRequest)
		return
	}

	halls, err := h.hallService.GetByScreenType(r.Context(), screenTypeID)
	if err != nil {
		http.Error(w, "Ошибка при получении залов", http.StatusInternalServerError)
		return
	}

	if len(halls) == 0 {
		http.Error(w, "Кинозалы с указанным типом экрана не найдены", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(halls)
}

// @Summary Поиск залов по названию (guest | user | admin)
// @Description Возвращает список залов, названия которых содержат указанную строку.
// @Tags Кинозалы
// @Produce json
// @Param query query string true "Строка для поиска"
// @Success 200 {array} Hall "Список найденных кинозалов"
// @Failure 400 {object} ErrorResponse "Строка поиска пуста"
// @Failure 404 {object} ErrorResponse "Залы не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /halls/search [get]
func (h *HallHandler) SearchHallsByName(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	query = PrepareString(query)
	if query == "" {
		http.Error(w, "Строка поиска пуста", http.StatusBadRequest)
		return
	}

	halls, err := h.hallService.SearchByName(r.Context(), query)
	if err != nil {
		http.Error(w, "Ошибка при поиске залов", http.StatusInternalServerError)
		return
	}

	if len(halls) == 0 {
		http.Error(w, "Кинозалы по запросу не найдены", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(halls)
}
