package handler

import (
	domain "cw/internal/domain/models"
	"cw/internal/domain/service"
	dto "cw/internal/dto/models"
	"cw/internal/utils"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ReviewHandler struct {
	reviewService *service.ReviewService
}

func NewReviewHandler(rs *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: rs}
}

// @Summary Получить отзывы
// @Description Возвращает пагинированный список отзывов с фильтрацией
// @Tags Отзывы
// @Produce json
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Количество элементов на странице" default(20) minimum(1) maximum(100)
// @Param movie_id query string false "Поиск по фильму"
// @Param user_id query string false "Поиск по пользователю"
// @Param rating_min query int false "Минимальный рейтинг" minimum(1) maximum(10)
// @Param rating_max query int false "Максимальный рейтинг" minimum(1) maximum(10)
// @Param comment query string false "Поиск по тексту отзыва (регистронезависимый поиск вхождений)"
// @Success 200 {object} dto.PaginatedReviewResponse "Список отзывов"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /reviews [get]
func (rh *ReviewHandler) GetReviews(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	dtoFilters := dto.ReviewFilters{
		MovieID: r.URL.Query().Get("movie_id"),
		UserID:  r.URL.Query().Get("user_id"),
		Comment: r.URL.Query().Get("comment"),
	}

	if ratingMin := r.URL.Query().Get("rating_min"); ratingMin != "" {
		if val, err := strconv.Atoi(ratingMin); err == nil {
			dtoFilters.RatingMin = val
		}
	}

	if ratingMax := r.URL.Query().Get("rating_max"); ratingMax != "" {
		if val, err := strconv.Atoi(ratingMax); err == nil {
			dtoFilters.RatingMax = val
		}
	}

	domainFilters := domain.ReviewFiltersFromDTO(dtoFilters)

	result, err := rh.reviewService.GetAll(r.Context(), domainFilters, page, limit)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// @Summary Создать отзыв к фильму
// @Description Создает новый отзыв для указанного фильма (только для авторизованных пользователей)
// @Tags Отзывы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie_id path string true "UUID фильма"
// @Param review body dto.CreateReviewRequest true "Данные отзыва"
// @Success 201 {object} dto.CreateResponse "Отзыв создан"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Router /movies/{movie_id}/reviews [post]
func (rh *ReviewHandler) CreateReviewForMovie(w http.ResponseWriter, r *http.Request) {
	movieID := chi.URLParam(r, "movie_id")

	var req dto.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainReview := domain.CreateReviewFromDTO(req)

	result, err := rh.reviewService.CreateForMovie(r.Context(), movieID, domainReview)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.CreateResponse{ID: result.ID})
}

// @Summary Обновить отзыв
// @Description Полностью обновляет информацию об отзыве (только автор отзыва или администратор)
// @Tags Отзывы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param review_id path string true "UUID отзыва"
// @Param review body dto.UpdateReviewRequest true "Новые данные отзыва"
// @Success 200 "Отзыв обновлен"
// @Failure 400 {object} dto.ErrorResponse "Неверный запрос"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /reviews/{review_id} [put]
func (rh *ReviewHandler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	reviewID := chi.URLParam(r, "review_id")

	var req dto.UpdateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, utils.NewBadRequest("Некорректные данные", err))
		return
	}

	domainReview := domain.UpdateReviewFromDTO(req)

	_, err := rh.reviewService.Update(r.Context(), reviewID, domainReview)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Удалить отзыв
// @Description Удаляет отзыв по идентификатору (только автор отзыва или администратор)
// @Tags Отзывы
// @Security BearerAuth
// @Param review_id path string true "UUID отзыва"
// @Success 204 "Отзыв удален"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Ресурс не найден"
// @Router /reviews/{review_id} [delete]
func (rh *ReviewHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	reviewID := chi.URLParam(r, "review_id")
	err := rh.reviewService.Delete(r.Context(), reviewID)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (rh *ReviewHandler) RegisterRoutes(r chi.Router) {
	r.Route("/reviews", func(r chi.Router) {
		r.Get("/", rh.GetReviews)
	})

	r.Route("/movies/{movie_id}/reviews", func(r chi.Router) {
		r.Post("/", rh.CreateReviewForMovie)
	})

	r.Route("/reviews/{review_id}", func(r chi.Router) {
		r.Put("/", rh.UpdateReview)
		r.Delete("/", rh.DeleteReview)
	})
}
