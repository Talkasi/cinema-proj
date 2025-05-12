package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// @Summary Получить все жанры
// @Description Возвращает список всех жанров
// @Tags genres
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Genre "Список жанров"
// @Failure 404 {string} string "Жанры не найдены"
// @Failure 500 {string} string "Ошибка сервера"
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
// @Description Возвращает жанр по его UUID
// @Tags genres
// @Produce json
// @Param id path string true "UUID жанра"
// @Security BearerAuth
// @Success 200 {object} Genre "Жанр"
// @Failure 400 {string} string "Неверный формат UUID"
// @Failure 404 {string} string "Жанр не найден"
// @Failure 500 {string} string "Ошибка сервера"
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
// @Description Создает новый жанр
// @Tags genres
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param genre body GenreData true "Данные жанра"
// @Success 201 {object} string "UUID созданного жанра"
// @Failure 400 {string} string "Неверный формат JSON"
// @Failure 403 {string} string "Доступ запрещен"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /genres [post]
func CreateGenre(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var g GenreData
		if !DecodeJSONBody(w, r, &g) {
			return
		}

		if !ValidateRequiredFields(w, map[string]string{
			"name":        g.Name,
			"description": g.Description,
		}) {
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
// @Description Обновляет существующий жанр
// @Tags genres
// @Accept json
// @Produce json
// @Param id path string true "UUID жанра"
// @Param genre body GenreData true "Обновленные данные жанра"
// @Security BearerAuth
// @Success 200 "Жанр успешно обновлен"
// @Failure 400 {string} string "Неверный формат UUID/JSON или пустые поля"
// @Failure 403 {string} string "Доступ запрещен"
// @Failure 404 {string} string "Жанр не найден"
// @Failure 500 {string} string "Ошибка сервера"
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

		if !ValidateRequiredFields(w, map[string]string{
			"name":        g.Name,
			"description": g.Description,
		}) {
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
// @Description Удаляет жанр по его UUID
// @Tags genres
// @Param id path string true "UUID жанра"
// @Security BearerAuth
// @Success 204 "Жанр успешно удален"
// @Failure 400 {string} string "Неверный формат UUID"
// @Failure 403 {string} string "Доступ запрещен"
// @Failure 404 {string} string "Жанр не найден"
// @Failure 500 {string} string "Ошибка сервера"
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
