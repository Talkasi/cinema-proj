package models

type SeatTypeData struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"required,min=10,max=1000"`
}

type SeatType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
