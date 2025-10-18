package entity

import (
	domain "cw/internal/domain/models"
)

type Review struct {
	ID      string `db:"id"`
	MovieID string `db:"movie_id"`
	UserID  string `db:"user_id"`
	Rating  int    `db:"rating"`
	Comment string `db:"comment"`
}

func ReviewToDomain(entity Review) domain.Review {
	return domain.Review{
		ID:      entity.ID,
		MovieID: entity.MovieID,
		UserID:  entity.UserID,
		Rating:  entity.Rating,
		Comment: entity.Comment,
	}
}

func ReviewFromDomain(domainReview domain.Review) Review {
	return Review{
		ID:      domainReview.ID,
		MovieID: domainReview.MovieID,
		UserID:  domainReview.UserID,
		Rating:  domainReview.Rating,
		Comment: domainReview.Comment,
	}
}

func ReviewsToDomain(entities []Review) []domain.Review {
	domainReviews := make([]domain.Review, len(entities))
	for i, entity := range entities {
		domainReviews[i] = ReviewToDomain(entity)
	}
	return domainReviews
}
