package models

import "time"

type MovieShowData struct {
	MovieID   string    `json:"movieId" validate:"required,uuid4"`
	HallID    string    `json:"hallId" validate:"required,uuid4"`
	StartTime time.Time `json:"startTime" validate:"required"`
	Language  string    `json:"language" validate:"required,oneof=English Spanish French German Italian Русский"`
}

type MovieShow struct {
	ID        string    `json:"id"`
	MovieID   string    `json:"movieId" validate:"required,uuid4"`
	HallID    string    `json:"hallId" validate:"required,uuid4"`
	StartTime time.Time `json:"startTime" validate:"required"`
	Language  string    `json:"language" validate:"required,oneof=English Spanish French German Italian Русский"`
}

type MovieShowResponse struct {
	ID             string        `json:"id"`
	Movie          MovieResponse `json:"movie"`
	Hall           HallResponse  `json:"hall"`
	StartTime      time.Time     `json:"startTime"`
	Language       string        `json:"language"`
	AvailableSeats int           `json:"availableSeats"`
}

type MovieShowFilter struct {
	MovieID  *string    `json:"movieId,omitempty" validate:"omitempty,uuid4"`
	HallID   *string    `json:"hallId,omitempty" validate:"omitempty,uuid4"`
	DateFrom *time.Time `json:"dateFrom,omitempty"`
	DateTo   *time.Time `json:"dateTo,omitempty"`
	Language *string    `json:"language,omitempty" validate:"omitempty,oneof=English Spanish French German Italian Русский"`
}
