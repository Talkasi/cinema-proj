package domain

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
	ID          string
	MovieShowID string
	SeatID      string
	UserID      *string
	Status      TicketStatusEnumType
	Price       float64
}

type TicketStatusData struct {
	UserID  string
	Reserve bool
}
