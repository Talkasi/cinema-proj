package handler

import (
	domain "cw/internal/domain/models"
	"cw/internal/domain/service"
	dto "cw/internal/dto/models"
	"cw/internal/utils"
	"strconv"

	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type GenreHandler struct {
	genreService *service.GenreService
}

func NewGenreHandler(gs *service.GenreService) *GenreHandler {
	return &GenreHandler{genreService: gs}
}

// @Summary Получить список жанров
// @Description Возвращает пагинированный список всех жанров с фильтрацией
// @Tags Жанры
// @Produce json
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Количество элементов на странице" default(20) minimum(1) maximum(100)
// @Param name query string false "Поиск по названию жанра (регистронезависимый поиск вхождений)"
// @Param description query string false "Поиск по описанию жанра (регистронезависимый поиск вхождений)"
// @Success 200 {object} dto.PaginatedGenreResponse "Список жанров"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Failure 500 {object} dto.ErrorResponse "Ошибка сервера"
// @Router /genres [get]
func (g *GenreHandler) GetGenres(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	dtoFilters := dto.GenreFilters{
		Name:        r.URL.Query().Get("name"),
		Description: r.URL.Query().Get("description"),
	}

	domainFilters := domain.GenreFiltersFromDTO(dtoFilters)

	result, err := g.genreService.GetAll(r.Context(), domainFilters, page, limit)
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

// @Summary Получить жанр по ID
// @Description Возвращает информацию о жанре по его идентификатору
// @Tags Жанры
// @Produce json
// @Param id path string true "UUID жанра"
// @Success 200 {object} dto.GenreResponse "Информация о жанре"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /genres/{id} [get]
func (g *GenreHandler) GetGenreByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	genre, err := g.genreService.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(genre); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Создать жанр
// @Description Создает новый жанр (только для администраторов)
// @Tags Жанры
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param genre body dto.CreateGenreRequest true "Данные жанра"
// @Success 201 {object} dto.CreateResponse "Жанр создан"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Router /genres [post]
func (g *GenreHandler) CreateGenre(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateGenreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainGenre := domain.CreateGenreFromDTO(req)

	result, err := g.genreService.Create(r.Context(), domainGenre)
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

// @Summary Обновить жанр
// @Description Обновляет информацию о жанре (только для администраторов)
// @Tags Жанры
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "UUID жанра"
// @Param genre body dto.UpdateGenreRequest true "Новые данные жанра"
// @Success 200 "Жанр обновлен"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /genres/{id} [put]
func (g *GenreHandler) UpdateGenre(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateGenreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainGenre := domain.UpdateGenreFromDTO(req)

	_, err := g.genreService.Update(r.Context(), id, domainGenre)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Удалить жанр
// @Description Удаляет жанр по идентификатору (только для администраторов)
// @Tags Жанры
// @Security BearerAuth
// @Param id path string true "UUID жанра"
// @Success 204 "Жанр удален"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /genres/{id} [delete]
func (g *GenreHandler) DeleteGenre(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := g.genreService.Delete(r.Context(), id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (g *GenreHandler) RegisterRoutes(r chi.Router) {
	r.Route("/genres", func(r chi.Router) {
		r.Get("/", g.GetGenres)
		r.Post("/", g.CreateGenre)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", g.GetGenreByID)
			r.Put("/", g.UpdateGenre)
			r.Delete("/", g.DeleteGenre)
		})
	})
}
