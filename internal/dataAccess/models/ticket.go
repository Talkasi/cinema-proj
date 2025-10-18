package entity

import (
	domain "cw/internal/domain/models"
)

type Ticket struct {
	ID          string  `db:"id"`
	MovieShowID string  `db:"movie_show_id"`
	SeatID      string  `db:"seat_id"`
	UserID      *string `db:"user_id"`
	Status      string  `db:"ticket_Status"`
	Price       float64 `db:"price"`
}

func TicketToDomain(entity Ticket) domain.Ticket {
	return domain.Ticket{
		ID:          entity.ID,
		MovieShowID: entity.MovieShowID,
		SeatID:      entity.SeatID,
		UserID:      entity.UserID,
		Price:       entity.Price,
		Status:      entity.Status,
	}
}

func TicketFromDomain(domainTicket domain.Ticket) Ticket {
	return Ticket{
		ID:          domainTicket.ID,
		MovieShowID: domainTicket.MovieShowID,
		SeatID:      domainTicket.SeatID,
		UserID:      domainTicket.UserID,
		Price:       domainTicket.Price,
		Status:      domainTicket.Status,
	}
}

func TicketsToDomain(entities []Ticket) []domain.Ticket {
	domainTickets := make([]domain.Ticket, len(entities))
	for i, entity := range entities {
		domainTickets[i] = TicketToDomain(entity)
	}
	return domainTickets
}
