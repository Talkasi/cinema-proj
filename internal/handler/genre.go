package handler

import (
	"cw/internal/models"
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
func (h *GenreHandler) GetGenres(w http.ResponseWriter, r *http.Request) {
	halls, err := h.genreService.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(halls)
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
func (h *GenreHandler) GetGenreByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	hall, err := h.genreService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(hall)
}

// @Summary Создать жанр (admin)
// @Description Создаёт новый жанр.
// @Tags Жанры фильмов
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param genre body GenreData true "Данные жанра"
// @Success 201 {object} CreateResponse "ID созданного жанра"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /genres [post]
func (h *GenreHandler) CreateGenre(w http.ResponseWriter, r *http.Request) {
	var data models.GenreData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	hall, err := h.genreService.Create(r.Context(), data)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(hall)
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
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Жанр не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /genres/{id} [put]
func (h *GenreHandler) UpdateGenre(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var data models.GenreData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	hall, err := h.genreService.Update(r.Context(), id, data)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(hall)
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
func (h *GenreHandler) DeleteGenre(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.genreService.Delete(r.Context(), id)
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
// func SearchGenres(db *pgxpool.Pool) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		query := r.URL.Query().Get("query")
// 		query = PrepareString(query)
// 		if query == "" {
// 			http.Error(w, "Строка поиска пуста", http.StatusBadRequest)
// 			return
// 		}

// 		rows, err := db.Query(context.Background(),
// 			"SELECT id, name, description FROM genres WHERE name ILIKE $1",
// 			"%"+query+"%")
// 		if IsError(w, err) {
// 			return
// 		}
// 		defer rows.Close()

// 		var genres []Genre
// 		for rows.Next() {
// 			var g Genre
// 			if err := rows.Scan(&g.ID, &g.Name, &g.Description); HandleDatabaseError(w, err, "жанром") {
// 				return
// 			}

// 			genres = append(genres, g)
// 		}

// 		if len(genres) == 0 {
// 			http.Error(w, "Жанры по запросу не найдены", http.StatusNotFound)
// 			return
// 		}

// 		json.NewEncoder(w).Encode(genres)
// 	}
// }
