package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// @Summary Получить все жанры
// @Description Возвращает список всех жанров
// @Tags genres
// @Produce json
// @Success 200 {array} Genre
// @Failure 500 {string} string "Внутренняя ошибка"
// @Router /genres [get]
func GetGenres(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(context.Background(), "SELECT id, name, description FROM genres")
		if err != nil {
			http.Error(w, "Ошибка при получении жанров", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var genres []Genre
		for rows.Next() {
			var g Genre
			if err := rows.Scan(&g.ID, &g.Name, &g.Description); err != nil {
				http.Error(w, "Ошибка при сканировании жанра", http.StatusInternalServerError)
				return
			}
			genres = append(genres, g)
		}
		json.NewEncoder(w).Encode(genres)
	}
}

// @Summary Получить жанр по ID
// @Description Возвращает жанр по его UUID
// @Tags genres
// @Produce json
// @Param id path string true "ID жанра"
// @Success 200 {object} Genre
// @Failure 404 {string} string "Жанр не найден"
// @Failure 500 {string} string "Внутренняя ошибка"
// @Router /genres/{id} [get]
func GetGenreByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var g Genre
		err := db.QueryRow(context.Background(), "SELECT id, name, description FROM genres WHERE id = $1", id).Scan(&g.ID, &g.Name, &g.Description)
		if err == sql.ErrNoRows {
			http.Error(w, "Жанр не найден", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Ошибка при получении жанра", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(g)
	}
}

// @Summary Создать жанр
// @Description Добавляет новый жанр в базу данных
// @Tags genres
// @Accept json
// @Produce json
// @Param genre body Genre true "Новый жанр"
// @Success 201 {object} Genre
// @Failure 400 {string} string "Некорректный запрос"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /genres [post]
func CreateGenre(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var g Genre
		if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		g.ID = uuid.New().String()

		_, err := db.Exec(context.Background(), "INSERT INTO genres (id, name, description) VALUES ($1, $2, $3)", g.ID, g.Name, g.Description)
		if err != nil {
			http.Error(w, "Ошибка при вставке жанра", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(g)
	}
}

// @Summary Обновить жанр
// @Description Обновляет существующий жанр по ID
// @Tags genres
// @Accept json
// @Produce json
// @Param id path string true "ID жанра"
// @Param genre body Genre true "Обновленные данные жанра"
// @Success 200 {object} Genre
// @Failure 400 {string} string "Неверный JSON"
// @Failure 404 {string} string "Жанр не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /genres/{id} [put]
func UpdateGenre(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var g Genre
		if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		g.ID = id

		res, err := db.Exec(context.Background(), "UPDATE genres SET name=$1, description=$2 WHERE id=$3", g.Name, g.Description, g.ID)
		if err != nil {
			http.Error(w, "Ошибка при обновлении жанра", http.StatusInternalServerError)
			return
		}
		affected := res.RowsAffected()
		if affected == 0 {
			http.Error(w, "Жанр не найден", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(g)
	}
}

// @Summary Удалить жанр
// @Description Удаляет жанр по ID
// @Tags genres
// @Param id path string true "ID жанра"
// @Success 204 {string} string "Удалено"
// @Failure 404 {string} string "Жанр не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /genres/{id} [delete]
func DeleteGenre(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		res, err := db.Exec(context.Background(), "DELETE FROM genres WHERE id = $1", id)
		if err != nil {
			http.Error(w, "Ошибка при удалении жанра", http.StatusInternalServerError)
			return
		}
		affected := res.RowsAffected()
		if affected == 0 {
			http.Error(w, "Жанр не найден", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
