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

// @Summary Get reviews
// @Description Returns a paginated list of reviews with filtering
// @Tags Reviews
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param movie_id query string false "Search by movie"
// @Param user_id query string false "Search by user"
// @Param rating_min query int false "Minimum rating" minimum(1) maximum(10)
// @Param rating_max query int false "Maximum rating" minimum(1) maximum(10)
// @Param comment query string false "Search by review text (case-insensitive substring search)"
// @Success 200 {object} dto.PaginatedReviewResponse "Review list"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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
	if err := json.NewEncoder(w).Encode(result); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Create a review for a movie
// @Description Creates a new review for the specified movie (authorized users only)
// @Tags Reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie_id path string true "Movie UUID"
// @Param review body dto.CreateReviewRequest true "Review data"
// @Success 201 {object} dto.CreateResponse "Review created"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Router /movies/{movie_id}/reviews [post]
func (rh *ReviewHandler) CreateReviewForMovie(w http.ResponseWriter, r *http.Request) {
	movieID := chi.URLParam(r, "movie_id")

	var req dto.CreateReviewRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
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
	if err := json.NewEncoder(w).Encode(dto.CreateResponse{ID: result.ID}); err != nil {
		utils.WriteError(w, utils.NewInternal("failed to encode JSON", err))
		return
	}
}

// @Summary Update a review
// @Description Fully updates information about a review (review author or administrator only)
// @Tags Reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param review_id path string true "Review UUID"
// @Param review body dto.UpdateReviewRequest true "New review data"
// @Success 200 "Review updated"
// @Failure 400 {object} dto.ErrorResponse "Bad request"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
// @Router /reviews/{review_id} [put]
func (rh *ReviewHandler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	reviewID := chi.URLParam(r, "review_id")

	var req dto.UpdateReviewRequest
	if err := decodeAndValidateJSONBody(r, &req); err != nil {
		utils.WriteError(w, err)
		return
	}

	domainReview := domain.UpdateReviewFromDTO(req)

	_, err := rh.reviewService.Update(r.Context(), reviewID, domainReview)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
}

// @Summary Delete a review
// @Description Deletes a review by identifier (review author or administrator only)
// @Tags Reviews
// @Security BearerAuth
// @Param review_id path string true "Review UUID"
// @Success 204 "Review deleted"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Resource not found"
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
