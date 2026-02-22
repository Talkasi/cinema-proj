package entity

import (
	domain "cw/internal/domain/models"
)

type Genre struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

func GenreToDomain(entity Genre) domain.Genre {
	return domain.Genre{
		ID:          entity.ID,
		Name:        entity.Name,
		Description: entity.Description,
	}
}

func GenreFromDomain(domainGenre domain.Genre) Genre {
	return Genre{
		ID:          domainGenre.ID,
		Name:        domainGenre.Name,
		Description: domainGenre.Description,
	}
}

func GenresToDomain(entities []Genre) []domain.Genre {
	domainGenres := make([]domain.Genre, len(entities))
	for i, entity := range entities {
		domainGenres[i] = GenreToDomain(entity)
	}
	return domainGenres
}
