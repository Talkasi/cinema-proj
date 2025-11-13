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

type TicketHandler struct {
	ticketService *service.TicketService
}

func NewTicketHandler(ts *service.TicketService) *TicketHandler {
	return &TicketHandler{ticketService: ts}
}

// @Summary Получить билеты
// @Description Возвращает пагинированный список билетов с фильтрацией
// @Tags Билеты
// @Produce json
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Количество элементов на странице" default(20) minimum(1) maximum(100)
// @Param ticket_Status query string false "Фильтр по статусу билета (можно указать несколько через запятую)"
// @Param movie_show_id query string false "UUID сеанса"
// @Param price_min query number false "Минимальная цена билета"
// @Param price_max query number false "Максимальная цена билета"
// @Param seat_id query string false "Фильтр по ID места (можно указать несколько через запятую)"
// @Param user_id query string false "Фильтр по ID пользователя (можно указать несколько через запятую)"
// @Success 200 {object} dto.PaginatedTicketResponse "Список билетов"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /tickets [get]
func (th *TicketHandler) GetTickets(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	dtoFilters := dto.TicketFilters{
		Status:      splitCommaSeparated(r.URL.Query().Get("ticket_Status")),
		MovieShowID: r.URL.Query().Get("movie_show_id"),
		SeatID:      splitCommaSeparated(r.URL.Query().Get("seat_id")),
		UserID:      splitCommaSeparated(r.URL.Query().Get("user_id")),
	}

	if priceMin := r.URL.Query().Get("price_min"); priceMin != "" {
		if val, err := strconv.ParseFloat(priceMin, 64); err == nil {
			dtoFilters.PriceMin = val
		}
	}

	if priceMax := r.URL.Query().Get("price_max"); priceMax != "" {
		if val, err := strconv.ParseFloat(priceMax, 64); err == nil {
			dtoFilters.PriceMax = val
		}
	}

	domainFilters := domain.TicketFiltersFromDTO(dtoFilters)

	result, err := th.ticketService.GetAll(r.Context(), domainFilters, page, limit)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// @Summary Получить билет по ID
// @Description Возвращает информацию о билете по его идентификатору
// @Tags Билеты
// @Produce json
// @Param id path string true "UUID билета"
// @Success 200 {object} dto.TicketResponse "Информация о билете"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /tickets/{id} [get]
func (th *TicketHandler) GetTicketByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ticket, err := th.ticketService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

// @Summary Создать билет на сеанс
// @Description Создает новый билет для указанного киносеанса (только для авторизованных пользователей)
// @Tags Билеты
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie_show_id path string true "UUID киносеанса"
// @Param ticket body dto.CreateTicketRequest true "Данные билета"
// @Success 201 {object} dto.CreateResponse "Билет создан"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Router /movie-shows/{movie_show_id}/tickets [post]
func (th *TicketHandler) CreateTicketForMovieShow(w http.ResponseWriter, r *http.Request) {
	movieShowID := chi.URLParam(r, "movie_show_id")

	var req dto.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainTicket := domain.CreateTicketFromDTO(req)

	result, err := th.ticketService.CreateForMovieShow(r.Context(), movieShowID, domainTicket)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.CreateResponse{ID: result.ID})
}

// @Summary Обновить статус билета
// @Description Обновляет статус билета (бронирование/покупка)
// @Tags Билеты
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID билета"
// @Param Status body dto.UpdateStatusRequest true "Данные статуса билета"
// @Success 200 "Статус билета обновлен"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /tickets/{id} [patch]
func (th *TicketHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainTicket := domain.UpdateStatusFromDTO(req)

	_, err := th.ticketService.UpdateStatus(r.Context(), id, domainTicket)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Удалить билет
// @Description Удаляет билет по идентификатору (только для администраторов или владельца)
// @Tags Билеты
// @Security BearerAuth
// @Param id path string true "UUID билета"
// @Success 204 "Билет удален"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /tickets/{id} [delete]
func (th *TicketHandler) DeleteTicket(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := th.ticketService.Delete(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (th *TicketHandler) RegisterRoutes(r chi.Router) {
	r.Route("/tickets", func(r chi.Router) {
		r.Get("/", th.GetTickets)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", th.GetTicketByID)
			r.Patch("/", th.UpdateStatus)
			r.Delete("/", th.DeleteTicket)
		})
	})

	r.Route("/movie-shows/{movie_show_id}/tickets", func(r chi.Router) {
		r.Post("/", th.CreateTicketForMovieShow)
	})
}
