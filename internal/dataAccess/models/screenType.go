package entity

import (
	domain "cw/internal/domain/models"
)

type ScreenType struct {
	ID            string  `db:"id"`
	Name          string  `db:"name"`
	Description   string  `db:"description"`
	PriceModifier float64 `db:"price_modifier"`
}

func ScreenTypeToDomain(entity ScreenType) domain.ScreenType {
	return domain.ScreenType{
		ID:            entity.ID,
		Name:          entity.Name,
		Description:   entity.Description,
		PriceModifier: entity.PriceModifier,
	}
}

func ScreenTypeFromDomain(domainScreenType domain.ScreenType) ScreenType {
	return ScreenType{
		ID:            domainScreenType.ID,
		Name:          domainScreenType.Name,
		Description:   domainScreenType.Description,
		PriceModifier: domainScreenType.PriceModifier,
	}
}

func ScreenTypesToDomain(entities []ScreenType) []domain.ScreenType {
	domainScreenTypes := make([]domain.ScreenType, len(entities))
	for i, entity := range entities {
		domainScreenTypes[i] = ScreenTypeToDomain(entity)
	}
	return domainScreenTypes
}
