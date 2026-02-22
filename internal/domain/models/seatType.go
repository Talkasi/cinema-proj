package domain

import dto "cw/internal/dto/models"

type SeatType struct {
	ID            string
	Name          string
	Description   string
	PriceModifier float64
}

type SeatTypeFilters struct {
	Name        string
	Description string
}

func SeatTypeFiltersFromDTO(dtoFilters dto.SeatTypeFilters) SeatTypeFilters {
	return SeatTypeFilters{
		Name:        dtoFilters.Name,
		Description: dtoFilters.Description,
	}
}

func CreateSeatTypeFromDTO(req dto.CreateSeatTypeRequest) SeatType {
	return SeatType{
		Name:          req.Name,
		Description:   req.Description,
		PriceModifier: req.PriceModifier,
	}
}

func UpdateSeatTypeFromDTO(req dto.UpdateSeatTypeRequest) SeatType {
	return SeatType{
		Name:          req.Name,
		Description:   req.Description,
		PriceModifier: req.PriceModifier,
	}
}

func SeatTypeToDTO(domainSeatType SeatType) dto.SeatTypeResponse {
	return dto.SeatTypeResponse{
		ID:            domainSeatType.ID,
		Name:          domainSeatType.Name,
		Description:   domainSeatType.Description,
		PriceModifier: domainSeatType.PriceModifier,
	}
}

func SeatTypesToDTO(domainSeatTypes []SeatType) []dto.SeatTypeResponse {
	dtos := make([]dto.SeatTypeResponse, len(domainSeatTypes))
	for i, seatType := range domainSeatTypes {
		dtos[i] = SeatTypeToDTO(seatType)
	}
	return dtos
}
