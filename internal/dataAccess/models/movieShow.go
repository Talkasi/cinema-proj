package entity

import (
	domain "cw/internal/domain/models"
	"time"
)

type MovieShow struct {
	ID        string    `db:"id"`
	MovieID   string    `db:"movie_id"`
	HallID    string    `db:"hall_id"`
	StartTime time.Time `db:"start_time"`
	Language  string    `db:"language"`
}

func MovieShowToDomain(entity MovieShow) domain.MovieShow {
	return domain.MovieShow{
		ID:        entity.ID,
		MovieID:   entity.MovieID,
		HallID:    entity.HallID,
		StartTime: entity.StartTime,
		Language:  entity.Language,
	}
}

func MovieShowFromDomain(domainMovieShow domain.MovieShow) MovieShow {
	return MovieShow{
		ID:        domainMovieShow.ID,
		MovieID:   domainMovieShow.MovieID,
		HallID:    domainMovieShow.HallID,
		StartTime: domainMovieShow.StartTime,
		Language:  domainMovieShow.Language,
	}
}

func MovieShowsToDomain(entities []MovieShow) []domain.MovieShow {
	domainMovieShows := make([]domain.MovieShow, len(entities))
	for i, entity := range entities {
		domainMovieShows[i] = MovieShowToDomain(entity)
	}
	return domainMovieShows
}
