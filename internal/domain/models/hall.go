package domain

import dto "cw/internal/dto/models"

type HallFilters struct {
	Name         string
	ScreenTypeID string
	Description  string
}

type Hall struct {
	ID           string
	Name         string
	Description  string
	ScreenTypeID string
}

func HallFiltersFromDTO(dtoFilters dto.HallFilters) HallFilters {
	return HallFilters{
		Name:         dtoFilters.Name,
		ScreenTypeID: dtoFilters.ScreenTypeID,
		Description:  dtoFilters.Description,
	}
}

func CreateHallFromDTO(req dto.CreateHallRequest) Hall {
	return Hall{
		Name:         req.Name,
		Description:  req.Description,
		ScreenTypeID: req.ScreenTypeID,
	}
}

func UpdateHallFromDTO(req dto.UpdateHallRequest) Hall {
	return Hall{
		Name:         req.Name,
		Description:  req.Description,
		ScreenTypeID: req.ScreenTypeID,
	}
}

func HallToDTO(domainHall Hall) dto.HallResponse {
	return dto.HallResponse{
		ID:           domainHall.ID,
		Name:         domainHall.Name,
		Description:  domainHall.Description,
		ScreenTypeID: domainHall.ScreenTypeID,
	}
}

func HallsToDTO(domainHalls []Hall) []dto.HallResponse {
	dtos := make([]dto.HallResponse, len(domainHalls))
	for i, hall := range domainHalls {
		dtos[i] = HallToDTO(hall)
	}
	return dtos
}
