package handler

import (
	"cw/internal/dto"
	"cw/internal/service"
	"cw/internal/utils"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ScreenTypeHandler struct {
	screenTypeService service.ScreenTypeService
}

func NewScreenTypeHandler(sts service.ScreenTypeService) *ScreenTypeHandler {
	return &ScreenTypeHandler{screenTypeService: sts}
}

// @Summary Получить все типы экранов (guest | user | admin)
// @Description Возвращает список всех типов экранов, содержащихся в базе данных.
// @Tags Типы экранов
// @Produce json
// @Success 200 {array} ScreenType "Список типов экранов"
// @Failure 404 {object} ErrorResponse "Типы экранов не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /screen-types [get]
func (st *ScreenTypeHandler) GetScreenTypes(w http.ResponseWriter, r *http.Request) {
	screenType, err := st.screenTypeService.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(screenType)
}

// @Summary Получить тип экрана по ID (guest | user | admin)
// @Description Возвращает тип экрана по ID.
// @Tags Типы экранов
// @Produce json
// @Param id path string true "ID типа экрана"
// @Success 200 {object} ScreenType "Тип экрана"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 404 {object} ErrorResponse "Тип экрана не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /screen-types/{id} [get]
func (st *ScreenTypeHandler) GetScreenTypeByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	screenType, err := st.screenTypeService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(dto.ScreenTypeFromDomain(screenType))
}

// @Summary Создать тип экрана (admin)
// @Description Создаёт новый тип экрана.
// @Tags Типы экранов
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param screen_type body ScreenTypeAdmin true "Данные типа экрана"
// @Success 201 {object} CreateResponse "ID созданного типа экрана"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /screen-types [post]
func (st *ScreenTypeHandler) CreateScreenType(w http.ResponseWriter, r *http.Request) {
	var data dto.ScreenTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	screenType, err := st.screenTypeService.Create(r.Context(), data.ToDomain())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(screenType.ID)
}

// @Summary Обновить тип экрана (admin)
// @Description Обновляет существующий тип экрана.
// @Tags Типы экранов
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID типа экрана"
// @Param screen_type body ScreenTypeAdmin true "Обновлённые данные типа экрана"
// @Success 200 "Данные о типе экрана успешно обновлены"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Тип экранов не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /screen-types/{id} [put]
func (st *ScreenTypeHandler) UpdateScreenType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var data dto.ScreenTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	_, err := st.screenTypeService.Update(r.Context(), id, data.ToDomain())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w)
}

// @Summary Удалить тип экрана (admin)
// @Description Удаляет тип экрана по ID.
// @Tags Типы экранов
// @Param id path string true "ID типа экрана"
// @Security BearerAuth
// @Success 204 "Данные о типе экрана успешно удалены"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Тип экрана не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /screen-types/{id} [delete]
func (st *ScreenTypeHandler) DeleteScreenType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := st.screenTypeService.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Поиск типов экранов по названию (guest | user | admin)
// @Description Возвращает типы экранов, название которых содержит указанную строку.
// @Tags Типы экранов
// @Produce json
// @Param query query string true "Поисковый запрос"
// @Success 200 {array} ScreenType "Список типов экранов"
// @Failure 400 {object} ErrorResponse "Строка поиска пуста"
// @Failure 404 {object} ErrorResponse "Типы экранов не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /screen-types/search [get]
func (st *ScreenTypeHandler) SearchScreenTypes(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	query = utils.PrepareString(query)
	if query == "" {
		http.Error(w, "Строка поиска пуста", http.StatusBadRequest)
		return
	}

	screenTypes, err := st.screenTypeService.SearchByName(r.Context(), query)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.ScreenTypeFromDomainList(screenTypes))
}
