package models

type TicketStatusEnumType string

const (
	Purchased TicketStatusEnumType = "Purchased"
	Reserved  TicketStatusEnumType = "Reserved"
	Available TicketStatusEnumType = "Available"
)

func (t TicketStatusEnumType) IsValid() bool {
	switch t {
	case Purchased, Reserved, Available:
		return true
	}
	return false
}

type Ticket struct {
	ID          string               `json:"id" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	MovieShowID string               `json:"movie_show_id" example:"9b165097-1c9f-4ea3-bef0-e505baa4ff63"`
	SeatID      string               `json:"seat_id" example:"c1bf35fb-4e5f-46cb-914b-bc8d76aaca23"`
	UserID      *string              `json:"user_id" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	Status      TicketStatusEnumType `json:"ticket_status" example:"Purchased"`
	Price       float64              `json:"price" example:"800"`
}

type TicketData struct {
	MovieShowID string               `json:"movie_show_id" example:"9b165097-1c9f-4ea3-bef0-e505baa4ff63"`
	SeatID      string               `json:"seat_id" example:"c1bf35fb-4e5f-46cb-914b-bc8d76aaca23"`
	UserID      *string              `json:"user_id,omitempty" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	Status      TicketStatusEnumType `json:"ticket_status" example:"Available"`
	Price       float64              `json:"price" example:"800"`
}

type TicketStatusData struct {
	UserID  string `json:"user_id,omitempty" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	Reserve bool   `json:"reserve" example:"true"`
}
