package handler

import (
	"cw/internal/dto"
	"cw/internal/service"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type GenreHandler struct {
	genreService service.GenreService
}

func NewGenreHandler(gs service.GenreService) *GenreHandler {
	return &GenreHandler{genreService: gs}
}

// @Summary Получить все жанры (guest | user | admin)
// @Description Возвращает список всех жанров, хранящихся в базе данных.
// @Tags Жанры фильмов
// @Produce json
// @Success 200 {array} Genre "Список жанров"
// @Failure 404 {object} ErrorResponse "Жанры не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /genres [get]
func (g *GenreHandler) GetGenres(w http.ResponseWriter, r *http.Request) {
	genre, err := g.genreService.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(dto.GenresFromDomainList(genre))
}

// @Summary Получить жанр по ID (guest | user | admin)
// @Description Возвращает жанр по ID.
// @Tags Жанры фильмов
// @Produce json
// @Param id path string true "ID жанра"
// @Success 200 {object} Genre "Жанр"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 404 {object} ErrorResponse "Жанр не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /genres/{id} [get]
func (g *GenreHandler) GetGenreByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	genre, err := g.genreService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(dto.GenreFromDomain(genre))
}

// @Summary Создать жанр (admin)
// @Description Создаёт новый жанр.
// @Tags Жанры фильмов
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param genre body GenreData true "Данные жанра"
// @Success 201 {object} CreateResponse "ID созданного жанра"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /genres [post]
func (g *GenreHandler) CreateGenre(w http.ResponseWriter, r *http.Request) {
	var data dto.GenreRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	genre, err := g.genreService.Create(r.Context(), data.ToDomain())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(genre.ID)
}

// @Summary Обновить жанр (admin)
// @Description Обновляет существующий жанр.
// @Tags Жанры фильмов
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID жанра"
// @Param genre body GenreData true "Новые данные жанра"
// @Success 200 "Данные о жанре успешно обновлены"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Жанр не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /genres/{id} [put]
func (g *GenreHandler) UpdateGenre(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var data dto.GenreRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	_, err := g.genreService.Update(r.Context(), id, data.ToDomain())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w)
}

// @Summary Удалить жанр (admin)
// @Description Удаляет жанр по ID.
// @Tags Жанры фильмов
// @Param id path string true "ID жанра"
// @Security BearerAuth
// @Success 204 "Данные о жанре успешно удалены"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Жанр не найден"
// @Failure 409 {object} ErrorResponse "Конфликт при удалении жанра"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /genres/{id} [delete]
func (g *GenreHandler) DeleteGenre(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := g.genreService.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Поиск жанров по имени (guest | user | admin)
// @Description Возвращает список жанров, имена которых содержат указанную строку (регистронезависимый поиск).
// @Tags Жанры фильмов
// @Produce json
// @Param query query string true "Строка для поиска"
// @Success 200 {array} Genre "Список найденных жанров"
// @Failure 400 {object} ErrorResponse "Строка поиска пуста"
// @Failure 404 {object} ErrorResponse "Жанры не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /genres/search [get]
func (g *GenreHandler) SearchGenres(w http.ResponseWriter, r *http.Request) {
	query := chi.URLParam(r, "query")
	genres, err := g.genreService.SearchByName(r.Context(), query)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(dto.GenresFromDomainList(genres))
}
