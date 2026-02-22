package entity

import (
	domain "cw/internal/domain/models"
)

type Hall struct {
	ID           string `db:"id"`
	Name         string `db:"name"`
	Description  string `db:"description"`
	ScreenTypeID string `db:"screen_type_id"`
}

func HallToDomain(entity Hall) domain.Hall {
	return domain.Hall{
		ID:           entity.ID,
		Name:         entity.Name,
		Description:  entity.Description,
		ScreenTypeID: entity.ScreenTypeID,
	}
}

func HallFromDomain(domainHall domain.Hall) Hall {
	return Hall{
		ID:           domainHall.ID,
		Name:         domainHall.Name,
		Description:  domainHall.Description,
		ScreenTypeID: domainHall.ScreenTypeID,
	}
}

func HallsToDomain(entities []Hall) []domain.Hall {
	domainHalls := make([]domain.Hall, len(entities))
	for i, entity := range entities {
		domainHalls[i] = HallToDomain(entity)
	}
	return domainHalls
}
