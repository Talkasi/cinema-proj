package dto

import "time"

type MovieShowFilters struct {
	MovieIDs      []string `form:"movie_id" example:"550e8400-e29b-41d4-a716-446655440000,550e8400-e29b-41d4-a716-446655440001"`
	HallIDs       []string `form:"hall_id" example:"550e8400-e29b-41d4-a716-446655440002,550e8400-e29b-41d4-a716-446655440003"`
	Languages     []string `form:"language" example:"Russkiy,English"`
	Date          string   `form:"date" example:"2024-01-15"`
	StartTimeFrom string   `form:"start_time_from" example:"18:00"`
	StartTimeTo   string   `form:"start_time_to" example:"22:00"`
}

type MovieShowResponse struct {
	ID        string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	MovieID   string    `json:"movie_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	HallID    string    `json:"hall_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	StartTime time.Time `json:"start_time" example:"2024-01-15T20:00:00Z"`
	Language  string    `json:"language" example:"Russkiy"`
	BasePrice float64   `json:"base_price" example:"350.50"`
}

type CreateMovieShowRequest struct {
	MovieID   string    `json:"movie_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	HallID    string    `json:"hall_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	StartTime time.Time `json:"start_time" validate:"required" example:"2024-01-15T19:30:00Z"`
	Language  string    `json:"language" validate:"required,min=1,max=50" example:"English"`
	BasePrice float64   `json:"base_price" validate:"min=0" example:"400.00"`
}

type UpdateMovieShowRequest struct {
	MovieID   string    `json:"movie_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	HallID    string    `json:"hall_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	StartTime time.Time `json:"start_time" validate:"required" example:"2024-01-15T21:00:00Z"`
	Language  string    `json:"language" validate:"required,min=1,max=50" example:"Russkiy"`
	BasePrice float64   `json:"base_price" validate:"min=0" example:"450.00"`
}
