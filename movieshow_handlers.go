package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func validateMovieShowData(w http.ResponseWriter, ms MovieShowData) bool {
	if _, err := uuid.Parse(ms.MovieID); err != nil {
		http.Error(w, "Неверный формат ID фильма", http.StatusBadRequest)
		return false
	}

	if _, err := uuid.Parse(ms.HallID); err != nil {
		http.Error(w, "Неверный формат ID зала", http.StatusBadRequest)
		return false
	}

	startDate := time.Date(1895, 3, 22, 0, 0, 0, 0, time.UTC)
	if ms.StartTime.Before(startDate) {
		http.Error(w, "Время начала киносеанса должно быть позже 22 марта 1895 года (первый в мире киносеанс)", http.StatusBadRequest)
		return false
	}

	if !ms.Language.IsValid() {
		http.Error(w, "Неизвестный язык киносеанса", http.StatusBadRequest)
		return false
	}

	return true
}

// @Summary Получить все киносеансы
// @Description Возвращает список всех киносеансов, хранящихся в базе данных.
// @Tags Киносеансы
// @Produce json
// @Security BearerAuth
// @Success 200 {array} MovieShow "Список киносеансов"
// @Failure 404 {object} ErrorResponse "Киносеансы не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows [get]
func GetMovieShows(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(context.Background(),
			"SELECT id, movie_id, hall_id, start_time, language FROM movie_shows")
		if HandleDatabaseError(w, err, "киносеансами фильмов") {
			return
		}
		defer rows.Close()

		var shows []MovieShow
		for rows.Next() {
			var ms MovieShow
			if err := rows.Scan(&ms.ID, &ms.MovieID, &ms.HallID, &ms.StartTime, &ms.Language); HandleDatabaseError(w, err, "киносеансом фильма") {
				return
			}
			shows = append(shows, ms)
		}

		if len(shows) == 0 {
			http.Error(w, "киносеансы фильмов не найдены", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(shows)
	}
}

// @Summary Получить киносеанс по ID
// @Description Возвращает даныне о киносеансе по ID.
// @Tags Киносеансы
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID киносеанса фильма"
// @Success 200 {object} MovieShow "Данные киносеанса"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 404 {object} ErrorResponse "Киносеанс не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows/{id} [get]
func GetMovieShowByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var ms MovieShow
		ms.ID = id.String()
		err := db.QueryRow(context.Background(),
			"SELECT movie_id, hall_id, start_time, language FROM movie_shows WHERE id = $1", id).
			Scan(&ms.MovieID, &ms.HallID, &ms.StartTime, &ms.Language)

		if IsError(w, err) {
			return
		}

		json.NewEncoder(w).Encode(ms)
	}
}

// @Summary Создать киносеанс
// @Description Создаёт новый киносеанс.
// @Tags Киносеансы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie_show body MovieShowData true "Данные киносеанса"
// @Success 201 {object} CreateResponse "ID созданного киносеанса"
// @Failure 400 {object} ErrorResponse "Неверные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows [post]
func CreateMovieShow(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ms MovieShowData
		if !DecodeJSONBody(w, r, &ms) {
			return
		}
		id := uuid.New().String()

		if !validateMovieShowData(w, ms) {
			return
		}

		_, err := db.Exec(context.Background(),
			"INSERT INTO movie_shows (id, movie_id, hall_id, start_time, language) VALUES ($1, $2, $3, $4, $5)",
			id, ms.MovieID, ms.HallID, ms.StartTime, ms.Language)

		if IsError(w, err) {
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(id)
	}
}

// @Summary Обновить киносеанс
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
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows/{id} [put]
func UpdateMovieShow(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var ms MovieShowData
		if !DecodeJSONBody(w, r, &ms) {
			return
		}

		if !validateMovieShowData(w, ms) {
			return
		}

		res, err := db.Exec(context.Background(),
			"UPDATE movie_shows SET movie_id=$1, hall_id=$2, start_time=$3, language=$4 WHERE id=$5",
			ms.MovieID, ms.HallID, ms.StartTime, ms.Language, id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Удалить киносеанс фильма
// @Description Удаляет данные о киносеансе.
// @Tags Киносеансы
// @Param id path string true "ID киносеанса"
// @Security BearerAuth
// @Success 204 "Данные о киносеансе удалёны"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Киносеанс не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows/{id} [delete]
func DeleteMovieShow(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		res, err := db.Exec(context.Background(),
			"DELETE FROM movie_shows WHERE id = $1", id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
