package dto

type SeatFilters struct {
	SeatTypeID    string `form:"seat_type_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	RowNumberMin  int    `form:"row_number_min" example:"1"`
	RowNumberMax  int    `form:"row_number_max" example:"10"`
	SeatNumberMin int    `form:"seat_number_min" example:"1"`
	SeatNumberMax int    `form:"seat_number_max" example:"20"`
}

type SeatResponse struct {
	ID         string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	HallID     string `json:"hall_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	RowNumber  int    `json:"row_number" example:"5"`
	SeatNumber int    `json:"seat_number" example:"12"`
	SeatTypeID string `json:"seat_type_id" example:"550e8400-e29b-41d4-a716-446655440002"`
}

type CreateSeatRequest struct {
	HallID     string `json:"hall_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	RowNumber  int    `json:"row_number" validate:"required,min=1" example:"3"`
	SeatNumber int    `json:"seat_number" validate:"required,min=1" example:"8"`
	SeatTypeID string `json:"seat_type_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
}

type UpdateSeatRequest struct {
	HallID     string `json:"hall_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	RowNumber  int    `json:"row_number" validate:"required,min=1" example:"4"`
	SeatNumber int    `json:"seat_number" validate:"required,min=1" example:"10"`
	SeatTypeID string `json:"seat_type_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440002"`
}
