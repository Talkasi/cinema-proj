package handler

import (
	"cw/internal/dto"
	"cw/internal/service"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type MovieShowHandler struct {
	movieShowService service.MovieShowService
}

func NewMovieShowHandler(ms service.MovieShowService) *MovieShowHandler {
	return &MovieShowHandler{movieShowService: ms}
}

// @Summary Получить все киносеансы (guest | user | admin)
// @Description Возвращает список всех киносеансов, хранящихся в базе данных.
// @Tags Киносеансы
// @Produce json
// @Success 200 {array} MovieShow "Список киносеансов"
// @Failure 404 {object} ErrorResponse "Киносеансы не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows [get]
func (ms *MovieShowHandler) GetMovieShows(w http.ResponseWriter, r *http.Request) {
	movieShows, err := ms.movieShowService.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(movieShows)
}

// @Summary Получить киносеанс по ID (guest | user | admin)
// @Description Возвращает даныне о киносеансе по ID.
// @Tags Киносеансы
// @Produce json
// @Param id path string true "ID киносеанса фильма"
// @Success 200 {object} MovieShow "Данные киносеанса"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 404 {object} ErrorResponse "Киносеанс не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows/{id} [get]
func (ms *MovieShowHandler) GetMovieShowByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	movieShow, err := ms.movieShowService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(dto.MovieShowFromDomain(movieShow))
}

// @Summary Создать киносеанс (admin)
// @Description Создаёт новый киносеанс (а также билеты на него)
// @Tags Киносеансы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie_show body MovieShowAdmin true "Данные киносеанса"
// @Success 201 {object} CreateResponse "ID созданного киносеанса"
// @Failure 400 {object} ErrorResponse "Неверные данные"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows [post]
func (ms *MovieShowHandler) CreateMovieShow(w http.ResponseWriter, r *http.Request) {
	var data dto.MovieShowRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	movieShow, err := ms.movieShowService.Create(r.Context(), data.ToDomain())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(movieShow.ID)
}

// @Summary Обновить киносеанс (admin)
// @Description Обновляет данные о киносеансе.
// @Tags Киносеансы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID киносеанса"
// @Param movie_show body MovieShowData true "Новые данные киносеанса"
// @Success 200 "Данные киносеанса обновлены"
// @Failure 400 {object} ErrorResponse "Неверные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Киносеанс не найден"
// @Failure 409 {object} ErrorResponse "Конфликт при обновлении киносеанса"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows/{id} [put]
func (ms *MovieShowHandler) UpdateMovieShow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var data dto.MovieShowRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	_, err := ms.movieShowService.Update(r.Context(), id, data.ToDomain())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w)
}

// @Summary Удалить киносеанс фильма (admin)
// @Description Удаляет данные о киносеансе.
// @Tags Киносеансы
// @Param id path string true "ID киносеанса"
// @Security BearerAuth
// @Success 204 "Данные о киносеансе удалёны"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Киносеанс не найден"
// @Failure 409 {object} ErrorResponse "Конфликт при удалении киносеанса"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows/{id} [delete]
func (ms *MovieShowHandler) DeleteMovieShow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := ms.movieShowService.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Получить киносеансы по ID фильма (guest | user | admin)
// @Description Возвращает киносеансы для указанного фильма в ближайшие N часов.
// @Tags Киносеансы
// @Produce json
// @Param movie_id path string true "ID фильма"
// @Param hours query integer false "Период в часах (по умолчанию 24)"
// @Success 200 {array} MovieShow "Данные о найденных киносеансах"
// @Failure 400 {object} ErrorResponse "Неверный формат ID фильма или параметра hours"
// @Failure 404 {object} ErrorResponse "Киносеансы для данного фильма не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies/{movie_id}/shows [get]
func (ms *MovieShowHandler) GetShowsByMovie(w http.ResponseWriter, r *http.Request) {
	movieId := chi.URLParam(r, "movie_id")
	movieShows, err := ms.movieShowService.GetByMovie(r.Context(), movieId)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.MovieShowsFromDomainList(movieShows))
}

// @Summary Получить сеансы на указанную дату (guest | user | admin)
// @Description Возвращает сеансы, начинающиеся в указанный день.
// @Tags Киносеансы
// @Produce json
// @Param date path string true "Дата (YYYY-MM-DD)"
// @Success 200 {array} MovieShow "Данные о киносеансах"
// @Failure 400 {object} ErrorResponse "Неверный формат даты"
// @Failure 404 {object} ErrorResponse "Киносеансы в указанную дату не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows/by-date/{date} [get]
func (ms *MovieShowHandler) GetShowsByDate(w http.ResponseWriter, r *http.Request) {
	date, parseErr := time.Parse(time.DateOnly, chi.URLParam(r, "date"))
	if parseErr != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}
	movieShows, err := ms.movieShowService.GetByDate(r.Context(), date)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.MovieShowsFromDomainList(movieShows))
}

// @Summary Получить ближайшие сеансы (guest | user | admin)
// @Description Возвращает сеансы, начинающиеся в ближайшие N часов.
// @Tags Киносеансы
// @Produce json
// @Param hours query integer false "Период в часах (по умолчанию 24)"
// @Success 200 {array} MovieShow "Данные о киносеансах"
// @Failure 400 {object} ErrorResponse "Неверный формат даты"
// @Failure 404 {object} ErrorResponse "Киносеансы в указанную дату не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows/upcoming [get]
func (ms *MovieShowHandler) GetUpcomingShows(w http.ResponseWriter, r *http.Request) {
	hours := 24
	if h := r.URL.Query().Get("hours"); h != "" {
		if parsedHours, err := strconv.Atoi(h); err == nil && parsedHours > 0 {
			hours = parsedHours
		} else {
			http.Error(w, "Период в часах должен быть больше 0.", http.StatusBadRequest)
			return
		}
	}

	movieShows, err := ms.movieShowService.GetUpcoming(r.Context(), hours)
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.MovieShowsFromDomainList(movieShows))
}
