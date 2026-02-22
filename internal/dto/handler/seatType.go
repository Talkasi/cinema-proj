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

// @Summary Получить список типов мест
// @Description Возвращает пагинированный список всех типов мест с фильтрацией
// @Tags Типы мест
// @Produce json
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Количество элементов на странице" default(20) minimum(1) maximum(100)
// @Param name query string false "Поиск по названию типа места (регистронезависимый поиск вхождений)"
// @Param description query string false "Поиск по описанию типа места (регистронезависимый поиск вхождений)"
// @Success 200 {object} dto.PaginatedSeatTypeResponse "Список типов мест"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
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

// @Summary Получить тип места по ID
// @Description Возвращает информацию о типе места по его идентификатору
// @Tags Типы мест
// @Produce json
// @Param id path string true "UUID типа места"
// @Success 200 {object} dto.SeatTypeResponse "Информация о типе места"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
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

// @Summary Создать тип места
// @Description Создает новый тип места (только для администраторов)
// @Tags Типы мест
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param seat_type body dto.CreateSeatTypeRequest true "Данные типа места"
// @Success 201 {object} dto.CreateResponse "Тип места создан"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Router /seat-types [post]
func (st *SeatTypeHandler) CreateSeatType(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSeatTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
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

// @Summary Обновить тип места
// @Description Полностью обновляет информацию о типе места (только для администраторов)
// @Tags Типы мест
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID типа места"
// @Param seat_type body dto.UpdateSeatTypeRequest true "Новые данные типа места"
// @Success 200 "Тип места обновлен"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /seat-types/{id} [put]
func (st *SeatTypeHandler) UpdateSeatType(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateSeatTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainSeatType := domain.UpdateSeatTypeFromDTO(req)

	_, err := st.seatTypeService.Update(r.Context(), id, domainSeatType)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Удалить тип места
// @Description Удаляет тип места по идентификатору (только для администраторов)
// @Tags Типы мест
// @Security BearerAuth
// @Param id path string true "UUID типа места"
// @Success 204 "Тип места удален"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
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
