package dto

import "cw/internal/domain"

type TicketStatusEnumTypeDto string

const (
	Purchased TicketStatusEnumTypeDto = "Purchased"
	Reserved  TicketStatusEnumTypeDto = "Reserved"
	Available TicketStatusEnumTypeDto = "Available"
)

func (t TicketStatusEnumTypeDto) IsValid() bool {
	switch t {
	case Purchased, Reserved, Available:
		return true
	}
	return false
}

type TicketResponse struct {
	ID          string                  `json:"id" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	MovieShowID string                  `json:"movie_show_id" example:"9b165097-1c9f-4ea3-bef0-e505baa4ff63"`
	SeatID      string                  `json:"seat_id" example:"c1bf35fb-4e5f-46cb-914b-bc8d76aaca23"`
	UserID      *string                 `json:"user_id" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	Status      TicketStatusEnumTypeDto `json:"ticket_status" example:"Purchased"`
	Price       float64                 `json:"price" example:"800"`
}

type TicketRequest struct {
	MovieShowID string                  `json:"movie_show_id" example:"9b165097-1c9f-4ea3-bef0-e505baa4ff63"`
	SeatID      string                  `json:"seat_id" example:"c1bf35fb-4e5f-46cb-914b-bc8d76aaca23"`
	UserID      *string                 `json:"user_id,omitempty" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	Status      TicketStatusEnumTypeDto `json:"ticket_status" example:"Available"`
	Price       float64                 `json:"price" example:"800"`
}

type TicketStatusData struct {
	UserID  string `json:"user_id,omitempty" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	Reserve bool   `json:"reserve" example:"true"`
}

func (t TicketRequest) ToDomain() domain.Ticket {
	return domain.Ticket{
		MovieShowID: t.MovieShowID,
		SeatID:      t.SeatID,
		UserID:      t.UserID,
		Status:      domain.TicketStatusEnumType(t.Status),
		Price:       t.Price,
	}
}

func TicketFromDomain(t domain.Ticket) TicketResponse {
	return TicketResponse{
		ID:          t.ID,
		MovieShowID: t.MovieShowID,
		SeatID:      t.SeatID,
		UserID:      t.UserID,
		Status:      TicketStatusEnumTypeDto(t.Status),
		Price:       t.Price,
	}
}

func TicketFromDomainList(screenTypes []domain.Ticket) []TicketResponse {
	result := make([]TicketResponse, len(screenTypes))
	for i, s := range screenTypes {
		result[i] = TicketFromDomain(s)
	}
	return result
}
