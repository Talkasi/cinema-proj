package models

type SeatData struct {
	HallID     string `json:"hallId" validate:"required,uuid4"`
	SeatTypeID string `json:"seatTypeId" validate:"required,uuid4"`
	RowNumber  int    `json:"rowNumber" validate:"required,min=1"`
	SeatNumber int    `json:"seatNumber" validate:"required,min=1"`
}

type SeatResponse struct {
	ID         string   `json:"id"`
	HallID     string   `json:"hallId"`
	SeatType   SeatType `json:"seatType"`
	RowNumber  int      `json:"rowNumber"`
	SeatNumber int      `json:"seatNumber"`
}

type Seat struct {
	ID         string `json:"id"`
	HallID     string `json:"hallId"`
	SeatTypeID string `json:"seatTypeId"`
	RowNumber  int    `json:"rowNumber"`
	SeatNumber int    `json:"seatNumber"`
}
