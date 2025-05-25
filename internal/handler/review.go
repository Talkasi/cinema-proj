package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReviewHandler struct {
	reviewService service.ReviewService
}

func NewReviewHandler(rs service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: rs}
}

// @Summary Получить все отзывы (admin)
// @Description Возвращает список всех отзывов, хранящихся в базе данных.
// @Tags Отзывы
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Review "Список отзывов"
// @Failure 404 {object} ErrorResponse "Отзывы не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /reviews [get]
func (rs *ReviewHandler) GetReviews(w http.ResponseWriter, r *http.Request) {
	review, err := rs.reviewService.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Message, err.Code)
		return
	}
	json.NewEncoder(w).Encode(review)
}

// @Summary Получить отзыв по ID (guset | user | admin)
// @Description Возвращает отзыв по ID.
// @Tags Отзывы
// @Produce json
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

// @Summary Создать отзыв (user* | admin)
// @Description Создаёт новый отзыв.
// @Tags Отзывы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param review body ReviewData true "Данные отзыва"
// @Success 201 {object} CreateResponse "ID созданного отзыва"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
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

		role := r.Header.Get("Role")
		user_id := r.Header.Get("UserID")
		// println("from token", user_id)
		// println("from body", reviewData.UserID)
		// println("from test", UsersData[len(UsersData)-1].ID)
		if (role != os.Getenv("CLAIM_ROLE_ADMIN")) && (reviewData.UserID != user_id) {
			http.Error(w, "Доступ запрещён", http.StatusForbidden)
			return
		}

		id := uuid.New()
		_, err := db.Exec(context.Background(),
			"INSERT INTO reviews (id, user_id, movie_id, rating, review_comment) VALUES ($1, $2, $3, $4, $5)",
			id, reviewData.UserID, reviewData.MovieID, reviewData.Rating, reviewData.Comment)

		if IsError(w, err) {
			// println(err.Error())
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(id.String())
	}
}

// @Summary Обновить отзыв (user* | admin)
// @Description Обновляет существующий отзыв.
// @Tags Отзывы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID отзыва"
// @Param review body ReviewData true "Новые данные отзыва"
// @Success 200 "Данные отзыва успешно обновлены"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
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

		role := r.Header.Get("Role")
		user_id := r.Header.Get("UserID")
		if (role != os.Getenv("CLAIM_ROLE_ADMIN")) && (reviewData.UserID != user_id) {
			http.Error(w, "Доступ запрещён", http.StatusForbidden)
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

// @Summary Удалить отзыв (user* | admin)
// @Description Удаляет отзыв по ID.
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

		role := r.Header.Get("Role")
		user_id := r.Header.Get("UserID")
		var review_owner_id string
		if role != os.Getenv("CLAIM_ROLE_ADMIN") {
			err := db.QueryRow(context.Background(),
				"SELECT user_id FROM reviews WHERE id = $1", id).
				Scan(&review_owner_id)
			if IsError(w, err) {
				return
			}

			if review_owner_id != user_id {
				http.Error(w, "Доступ запрещен", http.StatusForbidden)
				return
			}
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

// @Summary Получить отзывы по ID фильма (guest | user | admin)
// @Description Возвращает все отзывы для указанного фильма.
// @Tags Отзывы
// @Produce json
// @Param movie_id path string true "ID фильма"
// @Success 200 {array} Review "Список отзывов"
// @Failure 400 {object} ErrorResponse "Неверный формат ID фильма"
// @Failure 404 {object} ErrorResponse "Отзывы не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /movies/{movie_id}/reviews [get]
func GetReviewsByMovieID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		movieID, ok := ParseUUIDFromPath(w, r.PathValue("movie_id"))
		if !ok {
			return
		}

		rows, err := db.Query(context.Background(),
			"SELECT id, user_id, movie_id, rating, review_comment FROM reviews WHERE movie_id = $1",
			movieID)
		if IsError(w, err) {
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

// @Summary Получить отзывы пользователя (user* | admin)
// @Description Возвращает все отзывы указанного пользователя.
// @Tags Отзывы
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "ID пользователя"
// @Success 200 {array} Review "Список отзывов"
// @Failure 400 {object} ErrorResponse "Неверный формат ID пользователя"
// @Failure 404 {object} ErrorResponse "Отзывы не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /users/{user_id}/reviews [get]
func GetReviewsByUserID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := uuid.Parse(r.PathValue("user_id"))
		if err != nil {
			http.Error(w, "Неверный формат ID пользователя", http.StatusBadRequest)
			return
		}

		role := r.Header.Get("Role")
		user_id := r.Header.Get("UserID")
		if (role != os.Getenv("CLAIM_ROLE_ADMIN")) && (userID.String() != user_id) {
			http.Error(w, "Доступ запрещён", http.StatusForbidden)
			return
		}

		query := `
            SELECT r.id, r.user_id, r.movie_id, r.rating, r.review_comment
            FROM reviews r
            WHERE r.user_id = $1`

		rows, err := db.Query(context.Background(), query, userID)
		if IsError(w, err) {
			return
		}
		defer rows.Close()

		var reviews []Review
		for rows.Next() {
			var r Review
			if err := rows.Scan(
				&r.ID,
				&r.UserID,
				&r.MovieID,
				&r.Rating,
				&r.Comment,
			); err != nil {
				http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
				return
			}
			reviews = append(reviews, r)
		}

		if len(reviews) == 0 {
			http.Error(w, "Отзывы не найдены", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(reviews)
	}
}
