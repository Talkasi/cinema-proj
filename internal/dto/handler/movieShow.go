package handler

import (
	domain "cw/internal/domain/models"
	"cw/internal/domain/service"
	dto "cw/internal/dto/models"
	"cw/internal/utils"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type MovieShowHandler struct {
	movieShowService *service.MovieShowService
}

func NewMovieShowHandler(ms *service.MovieShowService) *MovieShowHandler {
	return &MovieShowHandler{movieShowService: ms}
}

// @Summary Получить список киносеансов
// @Description Возвращает пагинированный список всех киносеансов с фильтрацией
// @Tags Киносеансы
// @Produce json
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Количество элементов на странице" default(20) minimum(1) maximum(100)
// @Param movie_id query string false "Фильтр по ID фильма (можно указать несколько через запятую)"
// @Param hall_id query string false "Фильтр по ID зала (можно указать несколько через запятую)"
// @Param language query string false "Фильтр по языкам показа (можно указать несколько через запятую)"
// @Param date query string false "Фильтр по дате (YYYY-MM-DD)"
// @Param start_time_from query string false "Фильтр по времени начала (от)"
// @Param start_time_to query string false "Фильтр по времени начала (до)"
// @Success 200 {object} dto.PaginatedMovieShowResponse "Список киносеансов"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /movie-shows [get]
func (ms *MovieShowHandler) GetMovieShows(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	dtoFilters := dto.MovieShowFilters{
		MovieIDs:      splitCommaSeparated(r.URL.Query().Get("movie_id")),
		HallIDs:       splitCommaSeparated(r.URL.Query().Get("hall_id")),
		Languages:     splitCommaSeparated(r.URL.Query().Get("language")),
		Date:          r.URL.Query().Get("date"),
		StartTimeFrom: r.URL.Query().Get("start_time_from"),
		StartTimeTo:   r.URL.Query().Get("start_time_to"),
	}

	domainFilters := domain.MovieShowFiltersFromDTO(dtoFilters)

	result, err := ms.movieShowService.GetAll(r.Context(), domainFilters, page, limit)
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

// @Summary Получить киносеанс по ID
// @Description Возвращает информацию о киносеансе по его идентификатору
// @Tags Киносеансы
// @Produce json
// @Param id path string true "UUID киносеанса"
// @Success 200 {object} dto.MovieShowResponse "Информация о киносеансе"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /movie-shows/{id} [get]
func (ms *MovieShowHandler) GetMovieShowByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	movieShow, err := ms.movieShowService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(movieShow); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Создать киносеанс
// @Description Создает новый киносеанс (только для администраторов)
// @Tags Киносеансы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie_show body dto.CreateMovieShowRequest true "Данные киносеанса"
// @Success 201 {object} dto.CreateResponse "Киносеанс создан"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Router /movie-shows [post]
func (ms *MovieShowHandler) CreateMovieShow(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateMovieShowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainMovieShow := domain.CreateMovieShowFromDTO(req)

	result, err := ms.movieShowService.Create(r.Context(), domainMovieShow)
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

// @Summary Обновить киносеанс
// @Description Обновляет информацию о киносеансе (только для администраторов)
// @Tags Киносеансы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID киносеанса"
// @Param movie_show body dto.UpdateMovieShowRequest true "Новые данные киносеанса"
// @Success 200 "Киносеанс обновлен"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /movie-shows/{id} [put]
func (ms *MovieShowHandler) UpdateMovieShow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateMovieShowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainMovieShow := domain.UpdateMovieShowFromDTO(req)

	_, err := ms.movieShowService.Update(r.Context(), id, domainMovieShow)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Удалить киносеанс
// @Description Удаляет киносеанс по идентификатору (только для администраторов)
// @Tags Киносеансы
// @Security BearerAuth
// @Param id path string true "UUID киносеанса"
// @Success 204 "Киносеанс удален"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /movie-shows/{id} [delete]
func (ms *MovieShowHandler) DeleteMovieShow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := ms.movieShowService.Delete(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func splitCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func (ms *MovieShowHandler) RegisterRoutes(r chi.Router) {
	r.Route("/movie-shows", func(r chi.Router) {
		r.Get("/", ms.GetMovieShows)
		r.Post("/", ms.CreateMovieShow)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", ms.GetMovieShowByID)
			r.Put("/", ms.UpdateMovieShow)
			r.Delete("/", ms.DeleteMovieShow)
		})
	})
}
