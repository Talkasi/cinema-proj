package domain

import dto "cw/internal/dto/models"

type Review struct {
	ID      string
	MovieID string
	UserID  string
	Rating  int
	Comment string
}

type ReviewFilters struct {
	MovieID   string
	UserID    string
	RatingMin int
	RatingMax int
	Comment   string
}

func ReviewFiltersFromDTO(dtoFilters dto.ReviewFilters) ReviewFilters {
	return ReviewFilters{
		MovieID:   dtoFilters.MovieID,
		UserID:    dtoFilters.UserID,
		RatingMin: dtoFilters.RatingMin,
		RatingMax: dtoFilters.RatingMax,
		Comment:   dtoFilters.Comment,
	}
}

func CreateReviewFromDTO(req dto.CreateReviewRequest) Review {
	return Review{
		MovieID: req.MovieID,
		UserID:  req.UserID,
		Rating:  req.Rating,
		Comment: req.Comment,
	}
}

func UpdateReviewFromDTO(req dto.UpdateReviewRequest) Review {
	return Review{
		MovieID: req.MovieID,
		UserID:  req.UserID,
		Rating:  req.Rating,
		Comment: req.Comment,
	}
}

func ReviewToDTO(domainReview Review) dto.ReviewResponse {
	return dto.ReviewResponse{
		ID:      domainReview.ID,
		MovieID: domainReview.MovieID,
		UserID:  domainReview.UserID,
		Rating:  domainReview.Rating,
		Comment: domainReview.Comment,
	}
}

func ReviewsToDTO(domainReviews []Review) []dto.ReviewResponse {
	dtos := make([]dto.ReviewResponse, len(domainReviews))
	for i, review := range domainReviews {
		dtos[i] = ReviewToDTO(review)
	}
	return dtos
}
