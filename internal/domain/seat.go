package domain

type Seat struct {
	ID         string
	HallID     string
	SeatTypeID string
	RowNumber  int
	SeatNumber int
}

type SeatType struct {
	ID            string
	Name          string
	Description   string
	PriceModifier float64
}
