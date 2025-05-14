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

func validateAllReviewData(w http.ResponseWriter, r ReviewData) bool {
	r.Comment = PrepareString(r.Comment)

	if err := validateReviewUserID(r.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateReviewMovieID(r.MovieID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateReviewRating(r.Rating); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateReviewComment(r.Comment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

func validateReviewUserID(userID string) error {
	if _, err := uuid.Parse(userID); err != nil {
		return errors.New("неверный формат ID пользователя")
	}
	return nil
}

func validateReviewMovieID(movieID string) error {
	if _, err := uuid.Parse(movieID); err != nil {
		return errors.New("неверный формат ID фильма")
	}
	return nil
}

func validateReviewRating(rating int) error {
	if rating < 1 || rating > 10 {
		return errors.New("рейтинг должен быть от 1 до 10")
	}
	return nil
}

func validateReviewComment(comment string) error {
	validCommentRegex := regexp.MustCompile(`\S`)
	if !validCommentRegex.MatchString(comment) {
		return errors.New("комментарий не может быть пустым или состоять только из пробелов")
	}
	if len(comment) > 2000 {
		return errors.New("комментарий не может превышать 2000 символов")
	}
	return nil
}

// @Summary Получить все отзывы
// @Description Возвращает список всех отзывов, хранящихся в базе данных.
// @Tags Отзывы
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Review "Список отзывов"
// @Failure 404 {object} ErrorResponse "Отзывы не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /reviews [get]
func GetReviews(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(context.Background(), "SELECT id, user_id, movie_id, rating, review_comment FROM reviews")
		if HandleDatabaseError(w, err, "отзывами") {
			return
		}
		defer rows.Close()

		var reviews []Review
		for rows.Next() {
			var review Review
			if err := rows.Scan(&review.ID, &review.UserID, &review.MovieID, &review.Rating, &review.Comment); HandleDatabaseError(w, err, "отзывом") {
				return
			}
			reviews = append(reviews, review)
		}

		if len(reviews) == 0 {
			http.Error(w, "Отзывы не найдены", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(reviews)
	}
}

// @Summary Получить отзыв по его ID
// @Description Возвращает отзыв по ID.
// @Tags Отзывы
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID отзыва"
// @Success 200 {object} Review "Отзыв"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 404 {object} ErrorResponse "Отзыв не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /reviews/{id} [get]
func GetReviewByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var review Review
		review.ID = id.String()
		err := db.QueryRow(context.Background(),
			"SELECT user_id, movie_id, rating, review_comment FROM reviews WHERE id = $1", id).
			Scan(&review.UserID, &review.MovieID, &review.Rating, &review.Comment)

		if IsError(w, err) {
			return
		}

		json.NewEncoder(w).Encode(review)
	}
}

// @Summary Создать отзыв
// @Description Создаёт новый отзыв.
// @Tags Отзывы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param review body ReviewData true "Данные отзыва"
// @Success 201 {object} CreateResponse "ID созданного отзыва"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /reviews [post]
func CreateReview(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var reviewData ReviewData
		if !DecodeJSONBody(w, r, &reviewData) {
			return
		}

		if !validateAllReviewData(w, reviewData) {
			return
		}

		id := uuid.New()
		_, err := db.Exec(context.Background(),
			"INSERT INTO reviews (id, user_id, movie_id, rating, review_comment) VALUES ($1, $2, $3, $4, $5)",
			id, reviewData.UserID, reviewData.MovieID, reviewData.Rating, reviewData.Comment)

		if IsError(w, err) {
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(id.String())
	}
}

// @Summary Обновить отзыв
// @Description Обновляет существующий отзыв.
// @Tags Отзывы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID отзыва"
// @Param review body ReviewData true "Новые данные отзыва"
// @Success 200 "Данные отзыва успешно обновлены"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Отзыв не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /reviews/{id} [put]
func UpdateReview(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var reviewData ReviewData
		if !DecodeJSONBody(w, r, &reviewData) {
			return
		}

		if !validateAllReviewData(w, reviewData) {
			return
		}

		res, err := db.Exec(context.Background(),
			"UPDATE reviews SET user_id=$1, movie_id=$2, rating=$3, review_comment=$4 WHERE id=$5",
			reviewData.UserID, reviewData.MovieID, reviewData.Rating, reviewData.Comment, id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Удалить отзыв
// @Description Удаляет отзыв по его ID.
// @Tags Отзывы
// @Param id path string true "ID отзыва"
// @Security BearerAuth
// @Success 204 "Данные отзыва успешно удалены"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Отзыв не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /reviews/{id} [delete]
func DeleteReview(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		res, err := db.Exec(context.Background(),
			"DELETE FROM reviews WHERE id = $1", id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
