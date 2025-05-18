package main

import "time"

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
	ID          string `json:"id" example:"ad2805ab-bf4c-4f93-ac68-2e0a854022f8"`
	Name        string `json:"name" example:"Исторический"`
	Description string `json:"description" example:"Жанр игрового кинематографа, повествующий о той или иной эпохе, людях и событиях прошлых лет"`
}

type GenreData struct {
	Name        string `json:"name" example:"Исторический"`
	Description string `json:"description" example:"Жанр игрового кинематографа, повествующий о той или иной эпохе, людях и событиях прошлых лет"`
}

type MovieData struct {
	Title       string    `json:"title" example:"Властелин колец"`
	Duration    string    `json:"duration" example:"02:58:00"`
	Description string    `json:"description" example:"Эпическая история о кольце власти."`
	AgeLimit    int       `json:"age_limit" example:"12"`
	ReleaseDate time.Time `json:"release_date" example:"2001-12-19"`
	GenreIDs    []string  `json:"genre_ids" example:"[\"f297eeaf-e784-43bf-a068-eef84f75baa4\", \"c5c8e037-a073-4105-9941-21e1cb4e79dd\"]"`
}

type Movie struct {
	ID               string    `json:"id" example:"9b165097-1c9f-4ea3-bef0-e505baa4ff63"`
	Title            string    `json:"title" example:"Властелин колец"`
	Duration         string    `json:"duration" example:"02:58:00"`
	Rating           float64   `json:"rating,omitempty" example:"8.8"`
	Description      string    `json:"description" example:"Эпическая история о кольце власти."`
	AgeLimit         int       `json:"age_limit" example:"12"`
	BoxOfficeRevenue float64   `json:"box_office_revenue" example:"300000000"`
	ReleaseDate      time.Time `json:"release_date" example:"2001-12-19"`
	Genres           []Genre   `json:"genres"`
}

type Hall struct {
	ID           string  `json:"id" example:"9b165097-1c9f-4ea3-bef0-e505baa4ff63"`
	Name         string  `json:"name" example:"Зал 1"`
	ScreenTypeID string  `json:"screen_type_id" example:"de01f085-dffa-4347-88da-168560207511"`
	Description  *string `json:"description,omitempty" example:"Комфортабельный зал с современным оборудованием"`
}

type HallData struct {
	Name         string  `json:"name" example:"Зал 1"`
	ScreenTypeID string  `json:"screen_type_id" example:"de01f085-dffa-4347-88da-168560207511"`
	Description  *string `json:"description,omitempty" example:"Комфортабельный зал с современным оборудованием"`
}

type ScreenType struct {
	ID          string `json:"id" example:"de01f085-dffa-4347-88da-168560207511"`
	Name        string `json:"name" example:"IMAX"`
	Description string `json:"description" example:"Экран с технологией IMAX для максимального погружения"`
}

type ScreenTypeData struct {
	Name        string `json:"name" example:"IMAX"`
	Description string `json:"description" example:"Экран с технологией IMAX для максимального погружения"`
}

type MovieShow struct {
	ID        string           `json:"id" example:"9b165097-1c9f-4ea3-bef0-e505baa4ff63"`
	MovieID   string           `json:"movie_id" example:"1a2b3c4d-5e6f-7g8h-9i0j-k1l2m3n4o5p6"`
	HallID    string           `json:"hall_id" example:"de01f085-dffa-4347-88da-168560207511"`
	StartTime time.Time        `json:"start_time" example:"2023-10-01T14:30:00Z"`
	Language  LanguageEnumType `json:"language" example:"Русский"`
}

type MovieShowData struct {
	MovieID   string           `json:"movie_id" example:"1a2b3c4d-5e6f-7g8h-9i0j-k1l2m3n4o5p6"`
	HallID    string           `json:"hall_id" example:"de01f085-dffa-4347-88da-168560207511"`
	StartTime time.Time        `json:"start_time" example:"2023-10-01T14:30:00Z"`
	Language  LanguageEnumType `json:"language" example:"Русский"`
}

type Ticket struct {
	ID          string               `json:"id" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	MovieShowID string               `json:"movie_show_id" example:"9b165097-1c9f-4ea3-bef0-e505baa4ff63"`
	SeatID      string               `json:"seat_id" example:"c1bf35fb-4e5f-46cb-914b-bc8d76aaca23"`
	UserID      *string              `json:"user_id" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	Status      TicketStatusEnumType `json:"ticket_status" example:"Purchased"`
	Price       float64              `json:"price" example:"800"`
}

type TicketData struct {
	MovieShowID string               `json:"movie_show_id" example:"9b165097-1c9f-4ea3-bef0-e505baa4ff63"`
	SeatID      string               `json:"seat_id" example:"c1bf35fb-4e5f-46cb-914b-bc8d76aaca23"`
	UserID      *string              `json:"user_id,omitempty" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	Status      TicketStatusEnumType `json:"ticket_status" example:"Available"`
	Price       float64              `json:"price" example:"800"`
}

type Seat struct {
	ID         string `json:"id" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	HallID     string `json:"hall_id" example:"de01f085-dffa-4347-88da-168560207511"`
	SeatTypeID string `json:"seat_type_id" example:"premium"`
	RowNumber  int    `json:"row_number" example:"5"`
	SeatNumber int    `json:"seat_number" example:"12"`
}

type SeatData struct {
	HallID     string `json:"hall_id" example:"de01f085-dffa-4347-88da-168560207511"`
	SeatTypeID string `json:"seat_type_id" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	RowNumber  int    `json:"row_number" example:"5"`
	SeatNumber int    `json:"seat_number" example:"12"`
}

type SeatType struct {
	ID          string `json:"id" example:"de01f085-dffa-4347-88da-168560207511"`
	Name        string `json:"name" example:"Премиум"`
	Description string `json:"description" example:"Комфортабельные места с дополнительным пространством и удобствами"`
}

type SeatTypeData struct {
	Name        string `json:"name" example:"Премиум"`
	Description string `json:"description" example:"Комфортабельные места с дополнительным пространством и удобствами"`
}

type User struct {
	ID           string `json:"id" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	Name         string `json:"name" example:"Иван Иванов"`
	Email        string `json:"email" example:"ivan@example.com"`
	PasswordHash string `json:"-"`
	BirthDate    string `json:"birth_date" example:"1990-01-01"`
	IsBlocked    bool   `json:"is_blocked" example:"false"`
	IsAdmin      bool   `json:"-"`
}

type UserData struct {
	Name      string `json:"name" example:"Иван Иванов"`
	Email     string `json:"email" example:"ivan@example.com"`
	BirthDate string `json:"birth_date" example:"1990-01-01"`
}

type UserLogin struct {
	Email        string `json:"email" example:"ivan@example.com"`
	PasswordHash string `json:"password_hash" example:"hashed_password"`
}

type UserRegister struct {
	Name         string `json:"name" example:"Иван Иванов"`
	Email        string `json:"email" example:"ivan@example.com"`
	PasswordHash string `json:"password_hash" example:"hashed_password"`
	BirthDate    string `json:"birth_date" example:"1990-01-01"`
}

type Review struct {
	ID      string `json:"id" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	UserID  string `json:"user_id" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	MovieID string `json:"movie_id" example:"2002d9d0-80fa-4bc3-ab85-8525d1e9674c"`
	Rating  int    `json:"rating" example:"8"`
	Comment string `json:"review_comment" example:"Отличный фильм!"`
}

type ReviewData struct {
	UserID  string `json:"user_id" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	MovieID string `json:"movie_id" example:"2002d9d0-80fa-4bc3-ab85-8525d1e9674c"`
	Rating  int    `json:"rating" example:"8"`
	Comment string `json:"review_comment" example:"Отличный фильм!"`
}

type ErrorResponse struct {
	Message string `json:"message" example:"Описание ошибки"`
}

type CreateResponse struct {
	ID string `json:"id" example:"9b165097-1c9f-4ea3-bef0-e505baa4ff63"`
}

type AuthResponse struct {
	Token  string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
	UserID string `json:"user_id" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
}
