package entity

import (
	domain "cw/internal/domain/models"
)

type SeatType struct {
	ID            string  `db:"id"`
	Name          string  `db:"name"`
	Description   string  `db:"description"`
	PriceModifier float64 `db:"price_modifier"`
}

func SeatTypeToDomain(entity SeatType) domain.SeatType {
	return domain.SeatType{
		ID:            entity.ID,
		Name:          entity.Name,
		Description:   entity.Description,
		PriceModifier: entity.PriceModifier,
	}
}

func SeatTypeFromDomain(domainSeatType domain.SeatType) SeatType {
	return SeatType{
		ID:            domainSeatType.ID,
		Name:          domainSeatType.Name,
		Description:   domainSeatType.Description,
		PriceModifier: domainSeatType.PriceModifier,
	}
}

func SeatTypesToDomain(entities []SeatType) []domain.SeatType {
	domainSeatTypes := make([]domain.SeatType, len(entities))
	for i, entity := range entities {
		domainSeatTypes[i] = SeatTypeToDomain(entity)
	}
	return domainSeatTypes
}
