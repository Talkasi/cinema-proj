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

type ScreenTypeHandler struct {
	screenTypeService *service.ScreenTypeService
}

func NewScreenTypeHandler(sts *service.ScreenTypeService) *ScreenTypeHandler {
	return &ScreenTypeHandler{screenTypeService: sts}
}

// @Summary Получить список типов экранов
// @Description Возвращает пагинированный список всех типов экранов с фильтрацией
// @Tags Типы экранов
// @Produce json
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Количество элементов на странице" default(20) minimum(1) maximum(100)
// @Param name query string false "Поиск по названию типа экрана (регистронезависимый поиск вхождений)"
// @Param description query string false "Поиск по описанию типа экрана (регистронезависимый поиск вхождений)"
// @Success 200 {object} dto.PaginatedScreenTypeResponse "Список типов экранов"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /screen-types [get]
func (st *ScreenTypeHandler) GetScreenTypes(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	dtoFilters := dto.ScreenTypeFilters{
		Name:        r.URL.Query().Get("name"),
		Description: r.URL.Query().Get("description"),
	}

	domainFilters := domain.ScreenTypeFiltersFromDTO(dtoFilters)

	result, err := st.screenTypeService.GetAll(r.Context(), domainFilters, page, limit)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// @Summary Получить тип экрана по ID
// @Description Возвращает информацию о типе экрана по его идентификатору
// @Tags Типы экранов
// @Produce json
// @Param id path string true "UUID типа экрана"
// @Success 200 {object} dto.ScreenTypeResponse "Информация о типе экрана"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /screen-types/{id} [get]
func (st *ScreenTypeHandler) GetScreenTypeByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	screenType, err := st.screenTypeService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(screenType)
}

// @Summary Создать тип экрана
// @Description Создает новый тип экрана (только для администраторов)
// @Tags Типы экранов
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param screen_type body dto.CreateScreenTypeRequest true "Данные типа экрана"
// @Success 201 {object} dto.CreateResponse "Тип экрана создан"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Router /screen-types [post]
func (st *ScreenTypeHandler) CreateScreenType(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateScreenTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainScreenType := domain.CreateScreenTypeFromDTO(req)

	result, err := st.screenTypeService.Create(r.Context(), domainScreenType)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.CreateResponse{ID: result.ID})
}

// @Summary Обновить тип экрана
// @Description Полностью обновляет информацию о типе экрана (только для администраторов)
// @Tags Типы экранов
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID типа экрана"
// @Param screen_type body dto.UpdateScreenTypeRequest true "Новые данные типа экрана"
// @Success 200 "Тип экрана обновлен"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /screen-types/{id} [put]
func (st *ScreenTypeHandler) UpdateScreenType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateScreenTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainScreenType := domain.UpdateScreenTypeFromDTO(req)

	_, err := st.screenTypeService.Update(r.Context(), id, domainScreenType)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Удалить тип экрана
// @Description Удаляет тип экрана по идентификатору (только для администраторов)
// @Tags Типы экранов
// @Security BearerAuth
// @Param id path string true "UUID типа экрана"
// @Success 204 "Тип экрана удален"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /screen-types/{id} [delete]
func (st *ScreenTypeHandler) DeleteScreenType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := st.screenTypeService.Delete(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (st *ScreenTypeHandler) RegisterRoutes(r chi.Router) {
	r.Route("/screen-types", func(r chi.Router) {
		r.Get("/", st.GetScreenTypes)
		r.Post("/", st.CreateScreenType)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", st.GetScreenTypeByID)
			r.Put("/", st.UpdateScreenType)
			r.Delete("/", st.DeleteScreenType)
		})
	})
}
