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

// @Summary Получить список залов
// @Description Возвращает пагинированный список всех кинозалов с фильтрацией
// @Tags Кинозалы
// @Produce json
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Количество элементов на странице" default(20) minimum(1) maximum(100)
// @Param name query string false "Поиск по названию зала (регистронезависимый поиск вхождений)"
// @Param screen_type_id query string false "Фильтр по ID типа экрана"
// @Param description query string false "Поиск по описанию зала (регистронезависимый поиск вхождений)"
// @Success 200 {object} dto.PaginatedHallResponse "Список залов"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
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

// @Summary Получить зал по ID
// @Description Возвращает информацию о кинозале по его идентификатору
// @Tags Кинозалы
// @Produce json
// @Param id path string true "UUID зала"
// @Success 200 {object} dto.HallResponse "Информация о зале"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
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

// @Summary Создать зал
// @Description Создает новый кинозал (только для администраторов)
// @Tags Кинозалы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hall body dto.CreateHallRequest true "Данные зала"
// @Success 201 {object} dto.CreateResponse "Зал создан"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Router /halls [post]
func (h *HallHandler) CreateHall(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateHallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
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

// @Summary Обновить зал
// @Description Обновляет информацию о кинозале (только для администраторов)
// @Tags Кинозалы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID зала"
// @Param hall body dto.UpdateHallRequest true "Новые данные зала"
// @Success 200 "Зал обновлен"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /halls/{id} [put]
func (h *HallHandler) UpdateHall(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateHallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainHall := domain.UpdateHallFromDTO(req)

	_, err := h.hallService.Update(r.Context(), id, domainHall)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Удалить зал
// @Description Удаляет кинозал по идентификатору (только для администраторов)
// @Tags Кинозалы
// @Security BearerAuth
// @Param id path string true "UUID зала"
// @Success 204 "Зал удален"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
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
