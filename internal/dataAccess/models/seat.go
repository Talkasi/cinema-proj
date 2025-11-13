package entity

import (
	domain "cw/internal/domain/models"
)

type Seat struct {
	ID         string `db:"id"`
	HallID     string `db:"hall_id"`
	SeatTypeID string `db:"seat_type_id"`
	RowNumber  int    `db:"row_number"`
	SeatNumber int    `db:"seat_number"`
}

func SeatToDomain(entity Seat) domain.Seat {
	return domain.Seat{
		ID:         entity.ID,
		HallID:     entity.HallID,
		RowNumber:  entity.RowNumber,
		SeatNumber: entity.SeatNumber,
		SeatTypeID: entity.SeatTypeID,
	}
}

func SeatFromDomain(domainSeat domain.Seat) Seat {
	return Seat{
		ID:         domainSeat.ID,
		HallID:     domainSeat.HallID,
		RowNumber:  domainSeat.RowNumber,
		SeatNumber: domainSeat.SeatNumber,
		SeatTypeID: domainSeat.SeatTypeID,
	}
}

func SeatsToDomain(entities []Seat) []domain.Seat {
	domainSeats := make([]domain.Seat, len(entities))
	for i, entity := range entities {
		domainSeats[i] = SeatToDomain(entity)
	}
	return domainSeats
}
