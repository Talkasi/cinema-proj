package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Валидаторы для фильмов
func validateAllMovieData(w http.ResponseWriter, m MovieData) bool {
	if err := validateMovieTitle(m.Title); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateMovieDuration(m.Duration); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	// if err := validateMovieRating(m.Rating); err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return false
	// }

	if err := validateMovieDescription(m.Description); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateMovieAgeLimit(m.AgeLimit); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateMovieRevenue(m.BoxOfficeRevenue); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

func ValidateQuery(query string) error {
	if !regexp.MustCompile(`\S`).MatchString(query) {
		return errors.New("запрос не может быть пустым")
	}
	return nil
}

func validateMovieTitle(title string) error {
	if !regexp.MustCompile(`\S`).MatchString(title) {
		return errors.New("название фильма не может быть пустым")
	}
	if len(title) > 200 {
		return errors.New("название фильма не может превышать 200 символов")
	}
	return nil
}

func validateMovieDuration(duration string) error {
	parsedDuration, err := time.Parse("15:04:05", duration)
	if err != nil {
		return errors.New("неверный формат длительности (HH:MM:SS)")
	}

	if parsedDuration.Hour() == 0 && parsedDuration.Minute() == 0 && parsedDuration.Second() == 0 {
		return errors.New("длительность должна быть больше 0")
	}

	return nil
}

func validateMovieRating(rating *float64) error {
	if rating != nil && (*rating < 0 || *rating > 10) {
		return errors.New("рейтинг должен быть в промежутке от 0 до 10")
	}

	return nil
}

func validateMovieDescription(description string) error {
	if !regexp.MustCompile(`\S`).MatchString(description) {
		return errors.New("описание фильма не может быть пустым")
	}
	if len(description) > 1000 {
		return errors.New("описание фильма не может превышать 1000 символов")
	}
	return nil
}

func validateMovieAgeLimit(ageLimit int) error {
	validLimits := map[int]bool{0: true, 6: true, 12: true, 16: true, 18: true}
	if !validLimits[ageLimit] {
		return errors.New("возрастное ограничение должно быть одним из: 0, 6, 12, 16, 18")
	}
	return nil
}

func validateMovieRevenue(revenue float64) error {
	if revenue < 0 {
		return errors.New("кассовые сборы не могут быть отрицательными")
	}
	return nil
}

// Вспомогательные функции
func fetchGenresByMovieID(db *pgxpool.Pool, movieID string) ([]Genre, error) {
	rows, err := db.Query(context.Background(), `
		SELECT g.id, g.name, g.description
		FROM genres g
		JOIN movies_genres mg ON g.id = mg.genre_id
		WHERE mg.movie_id = $1`, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []Genre
	for rows.Next() {
		var g Genre
		if err := rows.Scan(&g.ID, &g.Name, &g.Description); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	return genres, nil
}

func insertMovieGenres(db *pgxpool.Pool, movieID string, genreIDs []string) error {
	for _, genreID := range genreIDs {
		if _, err := db.Exec(context.Background(),
			"INSERT INTO movies_genres (movie_id, genre_id) VALUES ($1, $2)",
			movieID, genreID); err != nil {
			return fmt.Errorf("failed to insert genre %s: %v", genreID, err)
		}
	}
	return nil
}

// @Summary Получить все фильмы
// @Tags movies
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Movie
// @Failure 404 {object} ErrorResponse "Фильмы не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies [get]
func GetMovies(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(context.Background(), `
			SELECT id, title, duration, rating, description, age_limit, box_office_revenue, release_date
			FROM movies`)
		if err != nil {
			http.Error(w, "Ошибка при получении фильмов", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var movies []Movie
		for rows.Next() {
			var m Movie
			if err := rows.Scan(&m.ID, &m.Title, &m.Duration, &m.Rating, &m.Description, &m.AgeLimit, &m.BoxOfficeRevenue, &m.ReleaseDate); err != nil {
				http.Error(w, "Ошибка при сканировании фильма", http.StatusInternalServerError)
				return
			}

			genres, err := fetchGenresByMovieID(db, m.ID)
			if err != nil {
				http.Error(w, "Ошибка при получении жанров", http.StatusInternalServerError)
				return
			}
			m.Genres = genres

			movies = append(movies, m)
		}

		if len(movies) == 0 {
			http.Error(w, "Фильмы не найдены", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(movies)
	}
}

// @Summary Получить фильм по ID
// @Tags movies
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID фильма"
// @Success 200 {object} Movie
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 404 {object} ErrorResponse "Фильм не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies/{id} [get]
func GetMovieByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var m Movie
		err := db.QueryRow(context.Background(), `
			SELECT id, title, duration, rating, description, age_limit, box_office_revenue, release_date
			FROM movies WHERE id = $1`, id).
			Scan(&m.ID, &m.Title, &m.Duration, &m.Rating, &m.Description, &m.AgeLimit, &m.BoxOfficeRevenue, &m.ReleaseDate)

		if IsError(w, err) {
			return
		}

		genres, err := fetchGenresByMovieID(db, m.ID)
		if IsError(w, err) {
			return
		}
		m.Genres = genres

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}
}

// @Summary Создать фильм
// @Tags movies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie body MovieData true "Данные фильма"
// @Success 201 {object} Movie
// @Failure 400 {object} ErrorResponse "Неверный формат JSON или данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies [post]
func CreateMovie(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data MovieData
		if !DecodeJSONBody(w, r, &data) {
			return
		}

		if !validateAllMovieData(w, data) {
			return
		}

		id := uuid.New()
		_, err := db.Exec(context.Background(), `
			INSERT INTO movies (id, title, duration, description, age_limit, box_office_revenue, release_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			id, data.Title, data.Duration, data.Description, data.AgeLimit, data.BoxOfficeRevenue, data.ReleaseDate)
		if IsError(w, err) {
			return
		}

		err = insertMovieGenres(db, id.String(), data.GenreIDs)
		if IsError(w, err) {
			return
		}

		movie := Movie{
			ID:               id.String(),
			Title:            data.Title,
			Duration:         data.Duration,
			Description:      data.Description,
			AgeLimit:         data.AgeLimit,
			BoxOfficeRevenue: data.BoxOfficeRevenue,
			ReleaseDate:      data.ReleaseDate,
		}

		genres, err := fetchGenresByMovieID(db, id.String())
		if IsError(w, err) {
			return
		}
		movie.Genres = genres

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(movie)
	}
}

// @Summary Обновить фильм
// @Tags movies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID фильма"
// @Param movie body MovieData true "Обновлённые данные фильма"
// @Success 200 {object} Movie
// @Failure 400 {object} ErrorResponse "Неверный формат ID/JSON или данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Фильм не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies/{id} [put]
func UpdateMovie(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var data MovieData
		if !DecodeJSONBody(w, r, &data) {
			return
		}

		if !validateAllMovieData(w, data) {
			return
		}

		res, err := db.Exec(context.Background(), `
			UPDATE movies SET title=$1, duration=$2, description=$3, age_limit=$4, box_office_revenue=$5, release_date=$6
			WHERE id=$7`,
			data.Title, data.Duration, data.Description, data.AgeLimit, data.BoxOfficeRevenue, data.ReleaseDate, id)
		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		_, err = db.Exec(context.Background(), "DELETE FROM movies_genres WHERE movie_id = $1", id)
		if IsError(w, err) {
			return
		}

		err = insertMovieGenres(db, id.String(), data.GenreIDs)
		if IsError(w, err) {
			return
		}

		movie := Movie{
			ID:               id.String(),
			Title:            data.Title,
			Duration:         data.Duration,
			Description:      data.Description,
			AgeLimit:         data.AgeLimit,
			BoxOfficeRevenue: data.BoxOfficeRevenue,
			ReleaseDate:      data.ReleaseDate,
		}

		genres, err := fetchGenresByMovieID(db, id.String())
		if IsError(w, err) {
			return
		}
		movie.Genres = genres

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(movie)
	}
}

// @Summary Удалить фильм
// @Tags movies
// @Param id path string true "ID фильма"
// @Security BearerAuth
// @Success 204 "Фильм успешно удалён"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Фильм не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies/{id} [delete]
func DeleteMovie(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		res, err := db.Exec(context.Background(), "DELETE FROM movies WHERE id = $1", id)
		if IsError(w, err) {
			return
		}
		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// @Summary Поиск фильмов по названию
// @Tags movies
// @Produce json
// @Security BearerAuth
// @Param query query string true "Поисковый запрос"
// @Success 200 {array} Movie
// @Failure 400 {object} ErrorResponse "Пустой поисковый запрос"
// @Failure 404 {object} ErrorResponse "Данные не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies/by-title/search [get]
func SearchMovies(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		if err := ValidateQuery(query); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		rows, err := db.Query(context.Background(), `
			SELECT id, title, duration, rating, description, age_limit, box_office_revenue, release_date
			FROM movies WHERE title ILIKE $1`, "%"+query+"%")
		if IsError(w, err) {
			return
		}
		defer rows.Close()

		var movies []Movie
		for rows.Next() {
			var m Movie
			err := rows.Scan(&m.ID, &m.Title, &m.Duration, &m.Rating, &m.Description, &m.AgeLimit, &m.BoxOfficeRevenue, &m.ReleaseDate)
			if IsError(w, err) {
				return
			}

			genres, err := fetchGenresByMovieID(db, m.ID)
			if IsError(w, err) {
				return
			}
			m.Genres = genres

			movies = append(movies, m)
		}

		if len(movies) == 0 {
			http.Error(w, "Фильмы не найдены", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(movies)
	}
}

// @Summary Получить фильмы по списку жанров (строгий поиск)
// @Description Возвращает только фильмы, которые относятся ко всем указанным жанрам
// @Tags movies
// @Produce json
// @Security BearerAuth
// @Param genre_ids query []string true "Список ID жанров" collectionFormat(multi)
// @Success 200 {array} Movie
// @Failure 400 {object} ErrorResponse "Неверный формат ID или не указаны жанры"
// @Failure 404 {object} ErrorResponse "Жанры не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies/by-genres/search [get]
func GetMoviesByAllGenres(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем список ID жанров из query-параметра
		genreIDs := r.URL.Query()["genre_ids"]
		if len(genreIDs) == 0 {
			http.Error(w, "Не указаны ID жанров", http.StatusBadRequest)
			return
		}

		conn, err := db.Acquire(context.Background())
		if err != nil {
			http.Error(w, "Ошибка подключения к БД", http.StatusInternalServerError)
			return
		}
		defer conn.Release()

		// Валидируем каждый UUID
		validGenreIDs := make([]string, 0, len(genreIDs))
		for _, id := range genreIDs {
			if _, err := uuid.Parse(id); err != nil {
				http.Error(w, fmt.Sprintf("Неверный формат ID: %s", id), http.StatusBadRequest)
				return
			}
			validGenreIDs = append(validGenreIDs, id)
		}

		// Проверяем существование всех жанров
		var existingGenresCount int
		err = db.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM genres WHERE id = ANY($1)", validGenreIDs).
			Scan(&existingGenresCount)
		if err != nil {
			http.Error(w, "Ошибка при проверке жанров", http.StatusInternalServerError)
			return
		}
		if existingGenresCount != len(validGenreIDs) {
			http.Error(w, "Некоторые жанры не найдены", http.StatusNotFound)
			return
		}

		// Получаем фильмы, которые относятся ко всем указанным жанрам
		query := `
            SELECT m.id, m.title, m.duration, m.rating, m.description, 
                   m.age_limit, m.box_office_revenue, m.release_date
            FROM movies m
            WHERE NOT EXISTS (
                SELECT id FROM unnest($1::uuid[]) AS id
                EXCEPT
                SELECT genre_id FROM movies_genres WHERE movie_id = m.id
            )
            ORDER BY m.title`

		rows, err := db.Query(context.Background(), query, validGenreIDs)
		if err != nil {
			http.Error(w, "Ошибка при получении фильмов: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var movies []Movie
		for rows.Next() {
			var m Movie
			if err := rows.Scan(&m.ID, &m.Title, &m.Duration, &m.Rating, &m.Description,
				&m.AgeLimit, &m.BoxOfficeRevenue, &m.ReleaseDate); err != nil {
				http.Error(w, "Ошибка при сканировании фильма", http.StatusInternalServerError)
				return
			}

			genres, err := fetchGenresByMovieID(db, m.ID)
			if err != nil {
				http.Error(w, "Ошибка при получении жанров", http.StatusInternalServerError)
				return
			}
			m.Genres = genres

			movies = append(movies, m)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(movies)
	}
}
