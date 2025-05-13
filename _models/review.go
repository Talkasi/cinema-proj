package models

import "time"

type ReviewData struct {
	MovieID string  `json:"movieId" validate:"required,uuid4"`
	Rating  float64 `json:"rating" validate:"required,min=0,max=10"`
	Comment *string `json:"comment,omitempty" validate:"omitempty,min=10"`
}

type ReviewResponse struct {
	ID        string        `json:"id"`
	User      UserProfile   `json:"user"`
	Movie     MovieResponse `json:"movie"`
	Rating    float64       `json:"rating"`
	Comment   *string       `json:"comment,omitempty"`
	CreatedAt time.Time     `json:"createdAt"`
}

type ReviewUpdate struct {
	Rating  *float64 `json:"rating,omitempty" validate:"omitempty,min=0,max=10"`
	Comment *string  `json:"comment,omitempty" validate:"omitempty,min=10"`
}
