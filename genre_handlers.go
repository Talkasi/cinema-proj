package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func validateAllGenreData(w http.ResponseWriter, g GenreData) bool {
	g.Name = PrepareString(g.Name)
	g.Description = PrepareString(g.Description)

	if err := validateGenreName(g.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateGenreDescription(g.Description); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

func validateGenreName(name string) error {
	validNameRegex := regexp.MustCompile(`^[A-Za-zА-Яа-яЁё\s-]+$`)
	if !validNameRegex.MatchString(name) {
		return errors.New("имя жанра может содержать только буквы, пробелы и дефисы")
	}

	if !regexp.MustCompile(`\S`).MatchString(name) {
		return errors.New("имя жанра не может состоять только из пробелов")
	}

	if len(name) == 0 || len(name) > 64 {
		return errors.New("имя жанра не может быть пустым и не может превышать 64 символа")
	}
	return nil
}

func validateGenreDescription(description string) error {
	validDescriptionRegex := regexp.MustCompile(`\S`)
	if !validDescriptionRegex.MatchString(description) {
		return errors.New("описание жанра не может быть пустым или состоять только из пробелов")
	}
	if len(description) > 1000 {
		return errors.New("описание жанра не может превышать 1000 символов")
	}
	return nil
}

// @Summary Получить все жанры
// @Description Возвращает список всех жанров, хранящихся в базе данных.
// @Tags Жанры фильмов
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Genre "Список жанров"
// @Failure 404 {object} ErrorResponse "Жанры не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /genres [get]
func GetGenres(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(context.Background(), "SELECT id, name, description FROM genres")
		if HandleDatabaseError(w, err, "жанрами") {
			return
		}
		defer rows.Close()

		var genres []Genre
		for rows.Next() {
			var g Genre
			if err := rows.Scan(&g.ID, &g.Name, &g.Description); HandleDatabaseError(w, err, "жанром") {
				return
			}
			genres = append(genres, g)
		}

		if len(genres) == 0 {
			http.Error(w, "Жанры не найдены", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(genres)
	}
}

// @Summary Получить жанр по ID
// @Description Возвращает жанр по ID.
// @Tags Жанры фильмов
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID жанра"
// @Success 200 {object} Genre "Жанр"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 404 {object} ErrorResponse "Жанр не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /genres/{id} [get]
func GetGenreByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var g Genre
		g.ID = id.String()
		err := db.QueryRow(context.Background(),
			"SELECT name, description FROM genres WHERE id = $1", id).
			Scan(&g.Name, &g.Description)

		if IsError(w, err) {
			return
		}

		json.NewEncoder(w).Encode(g)
	}
}

// @Summary Создать жанр
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
func CreateGenre(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var g GenreData
		if !DecodeJSONBody(w, r, &g) {
			return
		}

		if !validateAllGenreData(w, g) {
			return
		}

		id := uuid.New()
		_, err := db.Exec(context.Background(),
			"INSERT INTO genres (id, name, description) VALUES ($1, $2, $3)",
			id, g.Name, g.Description)

		if IsError(w, err) {
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(id.String())
	}
}

// @Summary Обновить жанр
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
func UpdateGenre(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var g GenreData
		if !DecodeJSONBody(w, r, &g) {
			return
		}

		if !validateAllGenreData(w, g) {
			return
		}

		res, err := db.Exec(context.Background(),
			"UPDATE genres SET name=$1, description=$2 WHERE id=$3",
			g.Name, g.Description, id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Удалить жанр
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
func DeleteGenre(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		res, err := db.Exec(context.Background(),
			"DELETE FROM genres WHERE id = $1", id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// @Summary Поиск жанров по имени
// @Description Возвращает список жанров, имена которых содержат указанную строку (регистронезависимый поиск).
// @Tags Жанры фильмов
// @Produce json
// @Security BearerAuth
// @Param query query string true "Строка для поиска"
// @Success 200 {array} Genre "Список найденных жанров"
// @Failure 400 {object} ErrorResponse "Строка поиска пуста"
// @Failure 404 {object} ErrorResponse "Жанры не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /genres/search [get]
func SearchGenres(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		query = PrepareString(query)
		if query == "" {
			http.Error(w, "Строка поиска пуста", http.StatusBadRequest)
			return
		}

		rows, err := db.Query(context.Background(),
			"SELECT id, name, description FROM genres WHERE name ILIKE $1",
			"%"+query+"%")
		if IsError(w, err) {
			return
		}
		defer rows.Close()

		var genres []Genre
		for rows.Next() {
			var g Genre
			if err := rows.Scan(&g.ID, &g.Name, &g.Description); HandleDatabaseError(w, err, "жанром") {
				return
			}

			genres = append(genres, g)
		}

		if len(genres) == 0 {
			http.Error(w, "Жанры по запросу не найдены", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(genres)
	}
}
