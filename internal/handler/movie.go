package handler

import (
	"cw/internal/dto"
	"cw/internal/service"
	"cw/internal/utils"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type MovieHandler struct {
	movieService service.MovieService
}

func NewMovieHandler(ms service.MovieService) *MovieHandler {
	return &MovieHandler{movieService: ms}
}

// @Summary Получить все фильмы (guest | user | admin)
// @Description Возвращает список всех фильмов, содержащихся в базе данных.
// @Tags Фильмы
// @Produce json
// @Success 200 {array} Movie "Список фильмов"
// @Failure 404 {object} ErrorResponse "Фильмы не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies [get]
func (m *MovieHandler) GetMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := m.movieService.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(dto.MoviesFromDomainList(movies))
}

// @Summary Получить фильм по ID (guest | user | admin)
// @Description Возвращает фильм по ID.
// @Tags Фильмы
// @Produce json
// @Param id path string true "ID фильма"
// @Success 200 {object} Movie "Фильм"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 404 {object} ErrorResponse "Фильм не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies/{id} [get]
func (m *MovieHandler) GetMovieByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	movie, err := m.movieService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(dto.MovieFromDomain(movie))
}

// @Summary Создать фильм (admin)
// @Description Создаёт новый фильм.
// @Tags Фильмы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie body MovieData true "Данные фильма"
// @Success 201 {object} CreateResponse "ID созданного фильма"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies [post]
func (m *MovieHandler) CreateMovie(w http.ResponseWriter, r *http.Request) {
	var data dto.MovieRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	movie, err := m.movieService.Create(r.Context(), data.ToDomain())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(movie.ID)
}

// @Summary Обновить фильм (admin)
// @Description Обновляет существующий фильм.
// @Tags Фильмы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID фильма"
// @Param movie body MovieData true "Новые данные фильма"
// @Success 200 "Данные о фильме успешно обновлены"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Фильм не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies/{id} [put]
func (m *MovieHandler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var data dto.MovieRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	_, err := m.movieService.Update(r.Context(), id, data.ToDomain())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w)
}

// @Summary Удалить фильм (admin)
// @Description Удаляет фильм по ID.
// @Tags Фильмы
// @Param id path string true "ID фильма"
// @Security BearerAuth
// @Success 204 "Данные о фильме успешно удалены"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Фильм не найден"
// @Failure 409 {object} ErrorResponse "Конфликт при удалении фильма"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies/{id} [delete]
func (m *MovieHandler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := m.movieService.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Поиск фильмов по названию (guest | user | admin)
// @Description Возвращает фильмы, в названии которых содержится заданная строка.
// @Tags Фильмы
// @Produce json
// @Param query query string true "Поисковый запрос"
// @Success 200 {array} Movie "Найденные фильмы"
// @Failure 400 {object} ErrorResponse "Строка поиска пуста"
// @Failure 404 {object} ErrorResponse "Данные не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies/by-title/search [get]
func (m *MovieHandler) SearchMoviesByName(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	query = utils.PrepareString(query)
	if query == "" {
		http.Error(w, "Строка поиска пуста", http.StatusBadRequest)
		return
	}

	movies, err := m.movieService.SearchByName(r.Context(), query)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.MoviesFromDomainList(movies))
}

// @Summary Получить фильмы по списку жанров (guest | user | admin)
// @Description Возвращает фильмы, относящиеся ко всем указанным жанрам.
// @Tags Фильмы
// @Produce json
// @Param genre_ids query []string true "Список ID жанров" collectionFormat(multi)
// @Success 200 {array} Movie "Найденные фильмы"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
// @Failure 404 {object} ErrorResponse "Жанры не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies/by-genres/search [get]
func (m *MovieHandler) GetMoviesByGenres(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	ids := []string{query} // ?

	movies, err := m.movieService.GetByGenres(r.Context(), ids)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.MoviesFromDomainList(movies))
}
