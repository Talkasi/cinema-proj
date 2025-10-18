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

type MovieHandler struct {
	movieService *service.MovieService
}

func NewMovieHandler(ms *service.MovieService) *MovieHandler {
	return &MovieHandler{movieService: ms}
}

// @Summary Получить список фильмов
// @Description Возвращает пагинированный список всех фильмов
// @Tags Фильмы
// @Produce json
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Количество элементов на странице" default(20) minimum(1) maximum(100)
// @Param genre query string false "Фильтр по жанру"
// @Param title query string false "Поиск по названию"
// @Success 200 {object} dto.PaginatedMovieResponse "Список фильмов"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /movies [get]
func (m *MovieHandler) GetMovies(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	dtoFilters := dto.MovieFilters{
		Title: r.URL.Query().Get("title"),
		Genre: r.URL.Query().Get("genre"),
	}

	domainFilters := domain.MovieFiltersFromDTO(dtoFilters)

	result, err := m.movieService.GetAll(r.Context(), domainFilters, page, limit)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// @Summary Получить фильм по ID
// @Description Возвращает информацию о фильме по его идентификатору
// @Tags Фильмы
// @Produce json
// @Param id path string true "UUID фильма"
// @Success 200 {object} dto.MovieResponse "Информация о фильме"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /movies/{id} [get]
func (m *MovieHandler) GetMovieByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	movie, err := m.movieService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}

// @Summary Создать новый фильм
// @Description Создаёт новый фильм (только для администраторов)
// @Tags Фильмы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie body dto.CreateMovieRequest true "Данные фильма"
// @Success 201 {object} dto.CreateResponse "Фильм создан"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Router /movies [post]
func (m *MovieHandler) CreateMovie(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateMovieRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainMovie := domain.CreateMovieFromDTO(req)

	result, err := m.movieService.Create(r.Context(), domainMovie)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.CreateResponse{ID: result.ID})
}

// @Summary Обновить фильм
// @Description Полностью обновляет информацию о фильме (только для администраторов)
// @Tags Фильмы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID фильма"
// @Param movie body dto.UpdateMovieRequest true "Новые данные фильма"
// @Success 200 "Фильм обновлен"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /movies/{id} [put]
func (m *MovieHandler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateMovieRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainMovie := domain.UpdateMovieFromDTO(req)

	_, err := m.movieService.Update(r.Context(), id, domainMovie)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Удалить фильм
// @Description Удаляет фильм по идентификатору (только для администраторов)
// @Tags Фильмы
// @Security BearerAuth
// @Param id path string true "UUID фильма"
// @Success 204 "Фильм удален"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /movies/{id} [delete]
func (m *MovieHandler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := m.movieService.Delete(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (m *MovieHandler) RegisterRoutes(r chi.Router) {
	r.Route("/movies", func(r chi.Router) {
		r.Get("/", m.GetMovies)
		r.Post("/", m.CreateMovie)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", m.GetMovieByID)
			r.Put("/", m.UpdateMovie)
			r.Delete("/", m.DeleteMovie)
		})
	})
}
