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

// @Summary Poluchit otzyvy
// @Description Vozvraschaet paginirovannyy spisok otzyvov s filtratsiey
// @Tags Otzyvy
// @Produce json
// @Param page query int false "Nomer stranitsy" default(1) minimum(1)
// @Param limit query int false "Kolichestvo elementov na stranitse" default(20) minimum(1) maximum(100)
// @Param movie_id query string false "Poisk po filmu"
// @Param user_id query string false "Poisk po polzovatelyu"
// @Param rating_min query int false "Minimalnyy reyting" minimum(1) maximum(10)
// @Param rating_max query int false "Maksimalnyy reyting" minimum(1) maximum(10)
// @Param comment query string false "Poisk po tekstu otzyva (registronezavisimyy poisk vkhozhdeniy)"
// @Success 200 {object} dto.PaginatedReviewResponse "Spisok otzyvov"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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

// @Summary Sozdat otzyv k filmu
// @Description Sozdaet novyy otzyv dlya ukazannogo filma (tolko dlya avtorizovannykh polzovateley)
// @Tags Otzyvy
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param movie_id path string true "UUID filma"
// @Param review body dto.CreateReviewRequest true "Dannye otzyva"
// @Success 201 {object} dto.CreateResponse "Otzyv sozdan"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
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

// @Summary Obnovit otzyv
// @Description Polnostyu obnovlyaet informatsiyu ob otzyve (tolko avtor otzyva ili administrator)
// @Tags Otzyvy
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param review_id path string true "UUID otzyva"
// @Param review body dto.UpdateReviewRequest true "Novye dannye otzyva"
// @Success 200 "Otzyv obnovlen"
// @Failure 400 {object} dto.ErrorResponse "Nevernyy zapros"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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

// @Summary Udalit otzyv
// @Description Udalyaet otzyv po identifikatoru (tolko avtor otzyva ili administrator)
// @Tags Otzyvy
// @Security BearerAuth
// @Param review_id path string true "UUID otzyva"
// @Success 204 "Otzyv udalen"
// @Failure 403 {object} dto.ErrorResponse "Dostup zapreschen"
// @Failure 404 {object} dto.ErrorResponse "Resurs ne nayden"
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
