package domain

import (
	dto "cw/internal/dto/models"
	"time"
)

type MovieShow struct {
	ID        string
	MovieID   string
	HallID    string
	StartTime time.Time
	Language  string
}

type MovieShowFilters struct {
	MovieIDs      []string
	HallIDs       []string
	Languages     []string
	Date          string
	StartTimeFrom string
	StartTimeTo   string
}

func MovieShowFiltersFromDTO(dtoFilters dto.MovieShowFilters) MovieShowFilters {
	return MovieShowFilters{
		MovieIDs:      dtoFilters.MovieIDs,
		HallIDs:       dtoFilters.HallIDs,
		Languages:     dtoFilters.Languages,
		Date:          dtoFilters.Date,
		StartTimeFrom: dtoFilters.StartTimeFrom,
		StartTimeTo:   dtoFilters.StartTimeTo,
	}
}

func CreateMovieShowFromDTO(req dto.CreateMovieShowRequest) MovieShow {
	return MovieShow{
		MovieID:   req.MovieID,
		HallID:    req.HallID,
		StartTime: req.StartTime,
		Language:  req.Language,
	}
}

func UpdateMovieShowFromDTO(req dto.UpdateMovieShowRequest) MovieShow {
	return MovieShow{
		MovieID:   req.MovieID,
		HallID:    req.HallID,
		StartTime: req.StartTime,
		Language:  req.Language,
	}
}

func MovieShowToDTO(domainMovieShow MovieShow) dto.MovieShowResponse {
	return dto.MovieShowResponse{
		ID:        domainMovieShow.ID,
		MovieID:   domainMovieShow.MovieID,
		HallID:    domainMovieShow.HallID,
		StartTime: domainMovieShow.StartTime,
		Language:  domainMovieShow.Language,
	}
}

func MovieShowsToDTO(domainMovieShows []MovieShow) []dto.MovieShowResponse {
	dtos := make([]dto.MovieShowResponse, len(domainMovieShows))
	for i, movieShow := range domainMovieShows {
		dtos[i] = MovieShowToDTO(movieShow)
	}
	return dtos
}
