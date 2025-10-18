package dto

type TicketFilters struct {
	Status      []string `form:"ticket_Status" example:"Available,Reserved"`
	MovieShowID string   `form:"movie_show_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PriceMin    float64  `form:"price_min" example:"300.0"`
	PriceMax    float64  `form:"price_max" example:"1000.0"`
	SeatID      []string `form:"seat_id" example:"550e8400-e29b-41d4-a716-446655440001,550e8400-e29b-41d4-a716-446655440002"`
	UserID      []string `form:"user_id" example:"550e8400-e29b-41d4-a716-446655440003,550e8400-e29b-41d4-a716-446655440004"`
}

type TicketResponse struct {
	ID          string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	MovieShowID string  `json:"movie_show_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	SeatID      string  `json:"seat_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	UserID      string  `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440003"`
	Price       float64 `json:"price" example:"450.50"`
	Status      string  `json:"ticket_Status" example:"Reserved"`
}

type CreateTicketRequest struct {
	MovieShowID string  `json:"movie_show_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	SeatID      string  `json:"seat_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	UserID      string  `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	Price       float64 `json:"price" validate:"required,min=0" example:"500.00"`
	Status      string  `json:"ticket_Status" validate:"required,oneof=Available Reserved Sold Used Cancelled" example:"Available"`
}

type UpdateTicketRequest struct {
	MovieShowID string  `json:"movie_show_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	SeatID      string  `json:"seat_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	UserID      string  `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	Price       float64 `json:"price" validate:"required,min=0" example:"550.00"`
	Status      string  `json:"ticket_Status" validate:"required,oneof=Available Reserved Sold Used Cancelled" example:"Sold"`
}

type UpdateStatusRequest struct {
	Status string `json:"ticket_Status" validate:"required,oneof=Available Reserved Sold Used Cancelled" example:"Sold"`
	UserID string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}
