package main

type Genre struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Movie struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Duration    string  `json:"duration"` // формат "HH:MM:SS"
	Rating      float64 `json:"rating"`
	Description string  `json:"description"`
	AgeLimit    int     `json:"age_limit"`
	BoxOffice   float64 `json:"box_office_revenue"`
	ReleaseDate string  `json:"release_date"` // формат "YYYY-MM-DD"
	Genres      []Genre `json:"genres,omitempty"`
}

type Hall struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Capacity        int    `json:"capacity"`
	EquipmentTypeID string `json:"equipment_type_id"`
	Description     string `json:"description"`
}

type EquipmentType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MovieShow struct {
	ID        string `json:"id"`
	MovieID   string `json:"movie_id"`
	HallID    string `json:"hall_id"`
	StartTime string `json:"start_time"` // формат "HH:MM:SS"
	Language  string `json:"language"`
}

type Ticket struct {
	ID          string  `json:"id"`
	MovieShowID string  `json:"movie_show_id"`
	SeatID      string  `json:"seat_id"`
	StatusID    string  `json:"ticket_status_id"`
	Price       float64 `json:"price"`
}

type TicketStatus struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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

type EquipmentTypeData struct {
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
	Name            string `json:"name"`
	Capacity        int    `json:"capacity"`
	EquipmentTypeID string `json:"equipment_type_id"`
	Description     string `json:"description"`
}
