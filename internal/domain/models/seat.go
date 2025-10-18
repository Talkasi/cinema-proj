package domain

import dto "cw/internal/dto/models"

type Seat struct {
	ID         string
	HallID     string
	RowNumber  int
	SeatNumber int
	SeatTypeID string
}

type SeatFilters struct {
	SeatTypeID    string
	RowNumberMin  int
	RowNumberMax  int
	SeatNumberMin int
	SeatNumberMax int
}

func SeatFiltersFromDTO(dtoFilters dto.SeatFilters) SeatFilters {
	return SeatFilters{
		SeatTypeID:    dtoFilters.SeatTypeID,
		RowNumberMin:  dtoFilters.RowNumberMin,
		RowNumberMax:  dtoFilters.RowNumberMax,
		SeatNumberMin: dtoFilters.SeatNumberMin,
		SeatNumberMax: dtoFilters.SeatNumberMax,
	}
}

func CreateSeatFromDTO(req dto.CreateSeatRequest) Seat {
	return Seat{
		HallID:     req.HallID,
		RowNumber:  req.RowNumber,
		SeatNumber: req.SeatNumber,
		SeatTypeID: req.SeatTypeID,
	}
}

func UpdateSeatFromDTO(req dto.UpdateSeatRequest) Seat {
	return Seat{
		HallID:     req.HallID,
		RowNumber:  req.RowNumber,
		SeatNumber: req.SeatNumber,
		SeatTypeID: req.SeatTypeID,
	}
}

func SeatToDTO(domainSeat Seat) dto.SeatResponse {
	return dto.SeatResponse{
		ID:         domainSeat.ID,
		HallID:     domainSeat.HallID,
		RowNumber:  domainSeat.RowNumber,
		SeatNumber: domainSeat.SeatNumber,
		SeatTypeID: domainSeat.SeatTypeID,
	}
}

func SeatsToDTO(domainSeats []Seat) []dto.SeatResponse {
	dtos := make([]dto.SeatResponse, len(domainSeats))
	for i, seat := range domainSeats {
		dtos[i] = SeatToDTO(seat)
	}
	return dtos
}
