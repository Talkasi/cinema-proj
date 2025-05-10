package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// @Summary Получить все отзывы
// @Tags reviews
// @Produce json
// @Success 200 {array} Review
// @Failure 500 {string} string "Ошибка сервера"
// @Router /reviews [get]
func GetReviews(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, user_id, movie_id, rating, review_comment FROM reviews")
		if err != nil {
			http.Error(w, "Ошибка при получении отзывов", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var reviews []Review
		for rows.Next() {
			var review Review
			if err := rows.Scan(&review.ID, &review.UserID, &review.MovieID, &review.Rating, &review.Comment); err != nil {
				http.Error(w, "Ошибка при сканировании отзыва", http.StatusInternalServerError)
				return
			}
			reviews = append(reviews, review)
		}
		json.NewEncoder(w).Encode(reviews)
	}
}

// @Summary Получить отзыв по ID
// @Tags reviews
// @Produce json
// @Param id path string true "ID отзыва"
// @Success 200 {object} Review
// @Failure 404 {string} string "Отзыв не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /reviews/{id} [get]
func GetReviewByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var review Review
		err := db.QueryRow("SELECT id, user_id, movie_id, rating, review_comment FROM reviews WHERE id = $1", id).
			Scan(&review.ID, &review.UserID, &review.MovieID, &review.Rating, &review.Comment)

		if err == sql.ErrNoRows {
			http.Error(w, "Отзыв не найден", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Ошибка при получении отзыва", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(review)
	}
}

// @Summary Создать отзыв
// @Tags reviews
// @Accept json
// @Produce json
// @Param review body Review true "Новый отзыв"
// @Success 201 {object} Review
// @Failure 400 {string} string "Неверный запрос"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /reviews [post]
func CreateReview(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var review Review
		if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		review.ID = uuid.New().String()

		_, err := db.Exec(`
			INSERT INTO reviews (id, user_id, movie_id, rating, review_comment)
			VALUES ($1, $2, $3, $4, $5)`,
			review.ID, review.UserID, review.MovieID, review.Rating, review.Comment)
		if err != nil {
			http.Error(w, "Ошибка при создании отзыва", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(review)
	}
}

// @Summary Обновить отзыв
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path string true "ID отзыва"
// @Param review body Review true "Обновленные данные отзыва"
// @Success 200 {object} Review
// @Failure 400 {string} string "Неверный запрос"
// @Failure 404 {string} string "Отзыв не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /reviews/{id} [put]
func UpdateReview(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var review Review
		if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		review.ID = id

		res, err := db.Exec(`
			UPDATE reviews SET user_id=$1, movie_id=$2, rating=$3, review_comment=$4
			WHERE id=$5`,
			review.UserID, review.MovieID, review.Rating, review.Comment, id)
		if err != nil {
			http.Error(w, "Ошибка при обновлении отзыва", http.StatusInternalServerError)
			return
		}
		count, _ := res.RowsAffected()
		if count == 0 {
			http.Error(w, "Отзыв не найден", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(review)
	}
}

// @Summary Удалить отзыв
// @Tags reviews
// @Param id path string true "ID отзыва"
// @Success 204 {string} string "Удалено"
// @Failure 404 {string} string "Отзыв не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /reviews/{id} [delete]
func DeleteReview(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		res, err := db.Exec("DELETE FROM reviews WHERE id = $1", id)
		if err != nil {
			http.Error(w, "Ошибка при удалении отзыва", http.StatusInternalServerError)
			return
		}
		count, _ := res.RowsAffected()
		if count == 0 {
			http.Error(w, "Отзыв не найден", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
