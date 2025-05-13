package main

import (
	"time"
)

type LanguageEnumType string

const (
	English LanguageEnumType = "English"
	Spanish LanguageEnumType = "Spanish"
	French  LanguageEnumType = "French"
	German  LanguageEnumType = "German"
	Italian LanguageEnumType = "Italian"
	Russian LanguageEnumType = "Русский"
)

func (l LanguageEnumType) IsValid() bool {
	switch l {
	case English, Spanish, French, German, Italian, Russian:
		return true
	}
	return false
}

type TicketStatusEnumType string

const (
	Purchased TicketStatusEnumType = "Purchased"
	Reserved  TicketStatusEnumType = "Reserved"
	Available TicketStatusEnumType = "Available"
)

func (t TicketStatusEnumType) IsValid() bool {
	switch t {
	case Purchased, Reserved, Available:
		return true
	}
	return false
}

type Genre struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Movie struct {
	ID               string  `json:"id"`
	Title            string  `json:"title"`
	Duration         string  `json:"duration"`
	Rating           float64 `json:"rating"`
	Description      string  `json:"description"`
	AgeLimit         int     `json:"age_limit"`
	BoxOfficeRevenue float64 `json:"box_office_revenue"`
	ReleaseDate      string  `json:"release_date"`
	Genres           []Genre `json:"genres,omitempty"`
}

type Hall struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Capacity     int    `json:"capacity"`
	ScreenTypeID string `json:"screen_type_id"`
	Description  string `json:"description"`
}

type ScreenType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MovieShow struct {
	ID        string           `json:"id"`
	MovieID   string           `json:"movie_id"`
	HallID    string           `json:"hall_id"`
	StartTime time.Time        `json:"start_time"`
	Language  LanguageEnumType `json:"language"`
}

type Ticket struct {
	ID          string               `json:"id"`
	MovieShowID string               `json:"movie_show_id"`
	SeatID      string               `json:"seat_id"`
	Status      TicketStatusEnumType `json:"ticket_status"`
	Price       float64              `json:"price"`
}

type Seat struct {
	ID         string `json:"id"`
	HallID     string `json:"hall_id"`
	SeatTypeID string `json:"seat_type_id"`
	RowNumber  int    `json:"row_number"`
	SeatNumber int    `json:"seat_number"`
}

type SeatType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type User struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`          // скрывать при передаче
	BirthDate    string `json:"birth_date"` // формат "YYYY-MM-DD"
	IsBlocked    bool   `json:"is_blocked"`
	IsAdmin      bool   `json:"-"`
}

type UserLogin struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password-hash"`
}

type UserRegister struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"password-hash"`
	BirthDate    string `json:"birth_date"` // формат "YYYY-MM-DD"
}

type Review struct {
	ID      string  `json:"id"`
	UserID  string  `json:"user_id"`
	MovieID string  `json:"movie_id"`
	Rating  float64 `json:"rating"`
	Comment string  `json:"review_comment"`
}

type ErrorResponse struct {
	Message string `json:"message" example:"Error description"`
}

type GenreData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ScreenTypeData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SeatTypeData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TicketStatusData struct {
	Name string `json:"name"`
}

type HallData struct {
	Name         string `json:"name"`
	Capacity     int    `json:"capacity"`
	ScreenTypeID string `json:"screen_type_id"`
	Description  string `json:"description"`
}
