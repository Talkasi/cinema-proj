package models

import "time"

type MovieCreate struct {
	Title       string    `json:"title" validate:"required,min=1,max=200"`
	Duration    string    `json:"duration" validate:"required,durationFormat"`
	Description string    `json:"description" validate:"required,min=10,max=1000"`
	AgeLimit    int       `json:"ageLimit" validate:"required,oneof=0 6 12 16 18"`
	ReleaseDate time.Time `json:"releaseDate" validate:"required"`
	GenreIDs    []string  `json:"genreIds" validate:"required,min=1,dive,uuid4"`
}

type MovieResponse struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Duration         string    `json:"duration"`
	Rating           *float64  `json:"rating,omitempty"`
	Description      string    `json:"description"`
	AgeLimit         int       `json:"ageLimit"`
	BoxOfficeRevenue float64   `json:"boxOfficeRevenue"`
	ReleaseDate      time.Time `json:"releaseDate"`
	Genres           []Genre   `json:"genres"`
}

type MovieUpdate struct {
	Title       *string    `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Duration    *string    `json:"duration,omitempty" validate:"omitempty,durationFormat"`
	Description *string    `json:"description,omitempty" validate:"omitempty,min=10,max=1000"`
	AgeLimit    *int       `json:"ageLimit,omitempty" validate:"omitempty,oneof=0 6 12 16 18"`
	ReleaseDate *time.Time `json:"releaseDate,omitempty" validate:"omitempty"`
	GenreIDs    []string   `json:"genreIds,omitempty" validate:"omitempty,min=1,dive,uuid4"`
}

type Movie struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Duration         string    `json:"duration"`
	Rating           *float64  `json:"rating,omitempty"`
	BoxOfficeRevenue float64   `json:"box_office_revenue,omitempty"`
	AgeLimit         int       `json:"ageLimit"`
	ReleaseDate      time.Time `json:"releaseDate,omitempty"`
}

type MovieData struct {
	Title            string    `json:"title"`
	Duration         string    `json:"duration"`
	Rating           *float64  `json:"rating,omitempty"`
	BoxOfficeRevenue float64   `json:"box_office_revenue,omitempty"`
	AgeLimit         int       `json:"ageLimit"`
	ReleaseDate      time.Time `json:"releaseDate,omitempty"`
}

type MovieFilter struct {
	Title           *string    `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	GenreIDs        []string   `json:"genreIds,omitempty" validate:"omitempty,dive,uuid4"`
	MinRating       *float64   `json:"minRating,omitempty" validate:"omitempty,min=0,max=10"`
	MaxRating       *float64   `json:"maxRating,omitempty" validate:"omitempty,min=0,max=10"`
	AgeLimit        *int       `json:"ageLimit,omitempty" validate:"omitempty,oneof=0 6 12 16 18"`
	ReleaseDateFrom *time.Time `json:"releaseDateFrom,omitempty"`
	ReleaseDateTo   *time.Time `json:"releaseDateTo,omitempty"`
}
