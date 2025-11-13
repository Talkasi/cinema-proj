package entity

import (
	domain "cw/internal/domain/models"
	"strings"
	"time"
)

type Movie struct {
	ID               string    `db:"id"`
	Title            string    `db:"title"`
	Duration         string    `db:"duration"`
	Description      string    `db:"description"`
	AgeLimit         int       `db:"age_limit"`
	Rating           float64   `db:"rating"`
	BoxOfficeRevenue float64   `db:"box_office_revenue"`
	ReleaseDate      time.Time `db:"release_date"`
	GenreIDs         []string
}

func MovieToDomain(entity Movie) domain.Movie {

	duration := entity.Duration
	if strings.Contains(duration, ".") {
		duration = strings.Split(duration, ".")[0]
	}

	return domain.Movie{
		ID:               entity.ID,
		Title:            entity.Title,
		Description:      entity.Description,
		Duration:         duration,
		AgeLimit:         entity.AgeLimit,
		Rating:           entity.Rating,
		BoxOfficeRevenue: entity.BoxOfficeRevenue,
		ReleaseDate:      entity.ReleaseDate,
		GenreIDs:         entity.GenreIDs,
	}
}

func MovieFromDomain(domainMovie domain.Movie) Movie {
	return Movie{
		ID:               domainMovie.ID,
		Title:            domainMovie.Title,
		Description:      domainMovie.Description,
		Duration:         domainMovie.Duration,
		AgeLimit:         domainMovie.AgeLimit,
		Rating:           domainMovie.Rating,
		BoxOfficeRevenue: domainMovie.BoxOfficeRevenue,
		ReleaseDate:      domainMovie.ReleaseDate,
		GenreIDs:         domainMovie.GenreIDs,
	}
}

func MoviesToDomain(entities []Movie) []domain.Movie {
	domainMovies := make([]domain.Movie, len(entities))
	for i, entity := range entities {
		domainMovies[i] = MovieToDomain(entity)
	}
	return domainMovies
}
