package domain

import dto "cw/internal/dto/models"

type ScreenType struct {
	ID            string
	Name          string
	Description   string
	PriceModifier float64
}

type ScreenTypeFilters struct {
	Name        string
	Description string
}

func ScreenTypeFiltersFromDTO(dtoFilters dto.ScreenTypeFilters) ScreenTypeFilters {
	return ScreenTypeFilters{
		Name:        dtoFilters.Name,
		Description: dtoFilters.Description,
	}
}

func CreateScreenTypeFromDTO(req dto.CreateScreenTypeRequest) ScreenType {
	return ScreenType{
		Name:          req.Name,
		Description:   req.Description,
		PriceModifier: req.PriceModifier,
	}
}

func UpdateScreenTypeFromDTO(req dto.UpdateScreenTypeRequest) ScreenType {
	return ScreenType{
		Name:          req.Name,
		Description:   req.Description,
		PriceModifier: req.PriceModifier,
	}
}

func ScreenTypeToDTO(domainScreenType ScreenType) dto.ScreenTypeResponse {
	return dto.ScreenTypeResponse{
		ID:            domainScreenType.ID,
		Name:          domainScreenType.Name,
		Description:   domainScreenType.Description,
		PriceModifier: domainScreenType.PriceModifier,
	}
}

func ScreenTypesToDTO(domainScreenTypes []ScreenType) []dto.ScreenTypeResponse {
	dtos := make([]dto.ScreenTypeResponse, len(domainScreenTypes))
	for i, screenType := range domainScreenTypes {
		dtos[i] = ScreenTypeToDTO(screenType)
	}
	return dtos
}
