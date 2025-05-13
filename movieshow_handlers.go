package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// @Summary Получить все показы фильмов
// @Tags movie-shows
// @Produce json
// @Success 200 {array} MovieShow
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows [get]
func GetMovieShows(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(context.Background(), "SELECT id, movie_id, hall_id, start_time, language FROM movie_shows")
		if err != nil {
			http.Error(w, "Ошибка при получении показов фильмов", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var shows []MovieShow
		for rows.Next() {
			var ms MovieShow
			if err := rows.Scan(&ms.ID, &ms.MovieID, &ms.HallID, &ms.StartTime, &ms.Language); err != nil {
				http.Error(w, "Ошибка при сканировании", http.StatusInternalServerError)
				return
			}
			shows = append(shows, ms)
		}
		json.NewEncoder(w).Encode(shows)
	}
}

// @Summary Получить показ фильма по ID
// @Tags movie-shows
// @Produce json
// @Param id path string true "ID показа фильма"
// @Success 200 {object} MovieShow
// @Failure 404 {object} ErrorResponse "Показ фильма не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows/{id} [get]
func GetMovieShowByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var ms MovieShow
		err := db.QueryRow(context.Background(), "SELECT id, movie_id, hall_id, start_time, language FROM movie_shows WHERE id = $1", id).
			Scan(&ms.ID, &ms.MovieID, &ms.HallID, &ms.StartTime, &ms.Language)

		if err == sql.ErrNoRows {
			http.Error(w, "Показ фильма не найден", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Ошибка при получении", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(ms)
	}
}

// @Summary Создать показ фильма
// @Tags movie-shows
// @Accept json
// @Produce json
// @Param movie_show body MovieShow true "Показ фильма"
// @Success 201 {object} MovieShow
// @Failure 400 {object} ErrorResponse "Неверный запрос"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows [post]
func CreateMovieShow(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ms MovieShow
		if err := json.NewDecoder(r.Body).Decode(&ms); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		ms.ID = uuid.New().String()

		_, err := db.Exec(context.Background(), "INSERT INTO movie_shows (id, movie_id, hall_id, start_time, language) VALUES ($1, $2, $3, $4, $5)",
			ms.ID, ms.MovieID, ms.HallID, ms.StartTime, ms.Language)
		if err != nil {
			http.Error(w, "Ошибка при вставке", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ms)
	}
}

// @Summary Обновить показ фильма
// @Tags movie-shows
// @Accept json
// @Produce json
// @Param id path string true "ID показа фильма"
// @Param movie_show body MovieShow true "Обновлённые данные показа"
// @Success 200 {object} MovieShow
// @Failure 400 {object} ErrorResponse "Неверный JSON"
// @Failure 404 {object} ErrorResponse "Показ не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows/{id} [put]
func UpdateMovieShow(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var ms MovieShow
		if err := json.NewDecoder(r.Body).Decode(&ms); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		ms.ID = id

		res, err := db.Exec(context.Background(), "UPDATE movie_shows SET movie_id=$1, hall_id=$2, start_time=$3, language=$4 WHERE id=$5",
			ms.MovieID, ms.HallID, ms.StartTime, ms.Language, ms.ID)
		if err != nil {
			http.Error(w, "Ошибка при обновлёнии", http.StatusInternalServerError)
			return
		}
		rows := res.RowsAffected()
		if rows == 0 {
			http.Error(w, "Показ не найден", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(ms)
	}
}

// @Summary Удалить показ фильма
// @Tags movie-shows
// @Param id path string true "ID показа фильма"
// @Success 204 {string} string "Удалено"
// @Failure 404 {object} ErrorResponse "Показ не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movie-shows/{id} [delete]
func DeleteMovieShow(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		res, err := db.Exec(context.Background(), "DELETE FROM movie_shows WHERE id = $1", id)
		if err != nil {
			http.Error(w, "Ошибка при удалёнии", http.StatusInternalServerError)
			return
		}
		rows := res.RowsAffected()
		if rows == 0 {
			http.Error(w, "Показ не найден", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
