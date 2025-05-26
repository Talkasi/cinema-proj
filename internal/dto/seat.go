package dto

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

type SeatTypeAdmin struct {
	Name          string  `json:"name" example:"Премиум"`
	Description   string  `json:"description" example:"Комфортабельные места с дополнительным пространством и удобствами"`
	PriceModifier float64 `json:"price_modifier" example:"1"`
}
