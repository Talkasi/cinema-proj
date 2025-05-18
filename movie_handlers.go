package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Валидаторы для фильмов
func validateAllMovieData(w http.ResponseWriter, m MovieData) bool {
	m.Title = PrepareString(m.Title)
	m.Description = PrepareString(m.Description)

	if err := validateMovieTitle(m.Title); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateMovieDuration(m.Duration); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateMovieDescription(m.Description); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateMovieAgeLimit(m.AgeLimit); err != nil {
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
	if rating != nil && (*rating < 1 || *rating > 10) {
		return errors.New("рейтинг должен быть в промежутке от 1 до 10")
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

	// rows, err := db.Query(context.Background(), `
	//     SELECT g.id, g.name, g.description
	//     FROM genres g
	//     WHERE EXISTS (
	//         SELECT 1 FROM movies_genres mg
	//         WHERE mg.movie_id = $1 AND mg.genre_id = g.id
	//     )`, movieID)
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

func insertMovieGenres(tx pgx.Tx, movieID string, genreIDs []string) error {
	if len(genreIDs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, genreID := range genreIDs {
		batch.Queue("INSERT INTO movies_genres (movie_id, genre_id) VALUES ($1, $2)",
			movieID, genreID)
	}

	br := tx.SendBatch(context.Background(), batch)
	defer br.Close()

	// Проверяем результаты всех запросов в batch
	for range genreIDs {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("failed to insert movie-genre relation: %w", err)
		}
	}

	return br.Close()
}

func updateMovieGenres(tx pgx.Tx, ctx context.Context, movieID uuid.UUID, genreIDs []string) error {
	// удаляем только те жанры, которых нет в новом списке
	_, err := tx.Exec(ctx, `
        DELETE FROM movies_genres 
        WHERE movie_id = $1 
        AND genre_id NOT IN (SELECT unnest($2::uuid[]))`,
		movieID, genreIDs)
	if err != nil {
		return fmt.Errorf("failed to delete old genres: %w", err)
	}

	// добавляем только новые жанры
	_, err = tx.Exec(ctx, `
        INSERT INTO movies_genres (movie_id, genre_id)
        SELECT $1, genre_id
        FROM unnest($2::uuid[]) AS genre_id
        ON CONFLICT (movie_id, genre_id) DO NOTHING`,
		movieID, genreIDs)

	return err
}

// @Summary Получить все фильмы (guest | user | admin)
// @Description Возвращает список всех фильмов, содержащихся в базе данных.
// @Tags Фильмы
// @Produce json
// @Success 200 {array} Movie "Список фильмов"
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
				http.Error(w, fmt.Sprintf("Ошибка при сканировании фильма %v", err), http.StatusInternalServerError)
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

// @Summary Создать фильм (admin)
// @Description Создаёт новый фильм.
// @Tags Фильмы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie body MovieData true "Данные фильма"
// @Success 201 {object} CreateResponse "ID созданного фильма"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
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

		ctx := r.Context()
		tx, err := db.Begin(ctx)
		if IsError(w, err) {
			return
		}
		defer func() {
			if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				log.Printf("failed to rollback transaction: %v", err)
			}
		}()

		id := uuid.New()
		_, err = tx.Exec(ctx, `
            INSERT INTO movies (id, title, duration, description, age_limit, release_date)
            VALUES ($1, $2, $3, $4, $5, $6)`,
			id, data.Title, data.Duration, data.Description,
			data.AgeLimit, data.ReleaseDate)
		if IsError(w, err) {
			return
		}

		if err := insertMovieGenres(tx, id.String(), data.GenreIDs); IsError(w, err) {
			return
		}

		if err := tx.Commit(ctx); IsError(w, err) {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(id.String())
	}
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
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
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

		ctx := r.Context()

		tx, err := db.Begin(ctx)
		if IsError(w, err) {
			return
		}
		defer func() {
			if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				log.Printf("failed to rollback transaction: %v", err)
			}
		}()

		res, err := tx.Exec(ctx, `
            UPDATE movies 
            SET title = $1, 
                duration = $2, 
                description = $3, 
                age_limit = $4, 
                release_date = $5
            WHERE id = $6`,
			data.Title, data.Duration, data.Description,
			data.AgeLimit, data.ReleaseDate, id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		if err := updateMovieGenres(tx, ctx, id, data.GenreIDs); IsError(w, err) {
			return
		}

		if err := tx.Commit(ctx); IsError(w, err) {
			return
		}

		w.WriteHeader(http.StatusOK)
	}
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

// @Summary Получить фильмы по списку жанров (guest | user | admin)
// @Description Возвращает фильмы, относящиеся ко всем указанным жанрам.
// @Tags Фильмы
// @Produce json
// @Param genre_ids query []string true "Список ID жанров" collectionFormat(multi)
// @Success 200 {array} Movie "Найденные фильмы"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 404 {object} ErrorResponse "Жанры не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies/by-genres/search [get]
func GetMoviesByAllGenres(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		genreIDs := r.URL.Query()["genre_ids"]
		if len(genreIDs) == 0 {
			http.Error(w, "Не указаны ID жанров", http.StatusBadRequest)
			return
		}

		validGenreIDs := make([]string, 0, len(genreIDs))
		for _, id := range genreIDs {
			if _, err := uuid.Parse(id); err != nil {
				http.Error(w, fmt.Sprintf("Неверный формат ID: %s", id), http.StatusBadRequest)
				return
			}
			validGenreIDs = append(validGenreIDs, id)
		}

		rows, err := db.Query(context.Background(), `
			SELECT *
			FROM movies
			WHERE id IN (
				SELECT movie_id
				FROM movies_genres
				WHERE genre_id = ANY($1::uuid[])
				GROUP BY movie_id
        		HAVING COUNT(DISTINCT genre_id) = $2
			)
			ORDER BY title`, validGenreIDs, len(validGenreIDs))
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

		if len(movies) == 0 {
			http.Error(w, "Фильмы не найдены", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(movies)
	}
}
