package domain

import dto "cw/internal/dto/models"

type Ticket struct {
	ID          string
	MovieShowID string
	SeatID      string
	UserID      *string
	Price       float64
	Status      string
}

type TicketFilters struct {
	Status      []string
	MovieShowID string
	PriceMin    float64
	PriceMax    float64
	SeatID      []string
	UserID      []string
}

func TicketFiltersFromDTO(dtoFilters dto.TicketFilters) TicketFilters {
	return TicketFilters{
		Status:      dtoFilters.Status,
		MovieShowID: dtoFilters.MovieShowID,
		PriceMin:    dtoFilters.PriceMin,
		PriceMax:    dtoFilters.PriceMax,
		SeatID:      dtoFilters.SeatID,
		UserID:      dtoFilters.UserID,
	}
}

func CreateTicketFromDTO(req dto.CreateTicketRequest) Ticket {
	return Ticket{
		MovieShowID: req.MovieShowID,
		SeatID:      req.SeatID,
		UserID:      stringPtr(req.UserID),
		Price:       req.Price,
		Status:      req.Status,
	}
}

func UpdateTicketFromDTO(req dto.UpdateTicketRequest) Ticket {
	return Ticket{
		MovieShowID: req.MovieShowID,
		SeatID:      req.SeatID,
		UserID:      stringPtr(req.UserID),
		Price:       req.Price,
		Status:      req.Status,
	}
}

func UpdateStatusFromDTO(req dto.UpdateStatusRequest) Ticket {
	return Ticket{
		UserID: stringPtr(req.UserID),
		Status: req.Status,
	}
}

func TicketToDTO(domainTicket Ticket) dto.TicketResponse {
	return dto.TicketResponse{
		ID:          domainTicket.ID,
		MovieShowID: domainTicket.MovieShowID,
		SeatID:      domainTicket.SeatID,
		UserID:      stringValue(domainTicket.UserID),
		Price:       domainTicket.Price,
		Status:      domainTicket.Status,
	}
}

func TicketsToDTO(domainTickets []Ticket) []dto.TicketResponse {
	dtos := make([]dto.TicketResponse, len(domainTickets))
	for i, ticket := range domainTickets {
		dtos[i] = TicketToDTO(ticket)
	}
	return dtos
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
