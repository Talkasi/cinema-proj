package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func fetchGenresByMovieID(db *sql.DB, movieID string) []Genre {
	rows, err := db.Query(`
		SELECT g.id, g.name, g.description
		FROM genres g
		JOIN movies_genres mg ON g.id = mg.genre_id
		WHERE mg.movie_id = $1`, movieID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var genres []Genre
	for rows.Next() {
		var g Genre
		_ = rows.Scan(&g.ID, &g.Name, &g.Description)
		genres = append(genres, g)
	}
	return genres
}

func insertMovieGenres(db *sql.DB, movieID string, genres []Genre) error {
	for _, g := range genres {
		if _, err := db.Exec("INSERT INTO movies_genres (movie_id, genre_id) VALUES ($1, $2)", movieID, g.ID); err != nil {
			return err
		}
	}
	return nil
}

// @Summary Получить все фильмы
// @Tags movies
// @Produce json
// @Success 200 {array} Movie
// @Failure 500 {string} string "Ошибка сервера"
// @Router /movies [get]
func GetMovies(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
			SELECT m.id, m.title, m.duration, m.rating, m.description, m.age_limit, m.box_office_revenue, m.release_date
			FROM movies m`)
		if err != nil {
			http.Error(w, "Ошибка при получении фильмов", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var movies []Movie
		for rows.Next() {
			var m Movie
			if err := rows.Scan(&m.ID, &m.Title, &m.Duration, &m.Rating, &m.Description, &m.AgeLimit, &m.BoxOffice, &m.ReleaseDate); err != nil {
				http.Error(w, "Ошибка при сканировании фильма", http.StatusInternalServerError)
				return
			}
			// Жанры подтянем отдельно
			m.Genres = fetchGenresByMovieID(db, m.ID)
			movies = append(movies, m)
		}
		json.NewEncoder(w).Encode(movies)
	}
}

// @Summary Получить фильм по ID
// @Tags movies
// @Produce json
// @Param id path string true "ID фильма"
// @Success 200 {object} Movie
// @Failure 404 {string} string "Фильм не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /movies/{id} [get]
func GetMovieByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var m Movie
		err := db.QueryRow(`
			SELECT id, title, duration, rating, description, age_limit, box_office_revenue, release_date
			FROM movies WHERE id = $1`, id).
			Scan(&m.ID, &m.Title, &m.Duration, &m.Rating, &m.Description, &m.AgeLimit, &m.BoxOffice, &m.ReleaseDate)

		if err == sql.ErrNoRows {
			http.Error(w, "Фильм не найден", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Ошибка при получении фильма", http.StatusInternalServerError)
			return
		}
		m.Genres = fetchGenresByMovieID(db, m.ID)
		json.NewEncoder(w).Encode(m)
	}
}

// @Summary Создать фильм
// @Tags movies
// @Accept json
// @Produce json
// @Param movie body Movie true "Новый фильм"
// @Success 201 {object} Movie
// @Failure 400 {string} string "Неверный запрос"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /movies [post]
func CreateMovie(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m Movie
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		m.ID = uuid.New().String()

		_, err := db.Exec(`
			INSERT INTO movies (id, title, duration, rating, description, age_limit, box_office_revenue, release_date)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			m.ID, m.Title, m.Duration, m.Rating, m.Description, m.AgeLimit, m.BoxOffice, m.ReleaseDate)
		if err != nil {
			http.Error(w, "Ошибка при создании фильма", http.StatusInternalServerError)
			return
		}

		if err := insertMovieGenres(db, m.ID, m.Genres); err != nil {
			http.Error(w, "Ошибка при добавлении жанров", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(m)
	}
}

// @Summary Обновить фильм
// @Tags movies
// @Accept json
// @Produce json
// @Param id path string true "ID фильма"
// @Param movie body Movie true "Обновленные данные фильма"
// @Success 200 {object} Movie
// @Failure 400 {string} string "Неверный запрос"
// @Failure 404 {string} string "Фильм не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /movies/{id} [put]
func UpdateMovie(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var m Movie
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		m.ID = id

		res, err := db.Exec(`
			UPDATE movies SET title=$1, duration=$2, rating=$3, description=$4, age_limit=$5, box_office_revenue=$6, release_date=$7
			WHERE id=$8`,
			m.Title, m.Duration, m.Rating, m.Description, m.AgeLimit, m.BoxOffice, m.ReleaseDate, id)
		if err != nil {
			http.Error(w, "Ошибка при обновлении", http.StatusInternalServerError)
			return
		}
		count, _ := res.RowsAffected()
		if count == 0 {
			http.Error(w, "Фильм не найден", http.StatusNotFound)
			return
		}

		// пересохраняем жанры
		db.Exec("DELETE FROM movies_genres WHERE movie_id = $1", m.ID)
		insertMovieGenres(db, m.ID, m.Genres)

		json.NewEncoder(w).Encode(m)
	}
}

// @Summary Удалить фильм
// @Tags movies
// @Param id path string true "ID фильма"
// @Success 204 {string} string "Удалено"
// @Failure 404 {string} string "Фильм не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /movies/{id} [delete]
func DeleteMovie(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		db.Exec("DELETE FROM movies_genres WHERE movie_id = $1", id)

		res, err := db.Exec("DELETE FROM movies WHERE id = $1", id)
		if err != nil {
			http.Error(w, "Ошибка при удалении", http.StatusInternalServerError)
			return
		}
		count, _ := res.RowsAffected()
		if count == 0 {
			http.Error(w, "Фильм не найден", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
