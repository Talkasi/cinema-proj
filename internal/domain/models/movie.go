package domain

import (
	"time"

	dto "cw/internal/dto/models"
)

type Movie struct {
	ID               string
	Title            string
	Description      string
	Duration         string // В формате "15:04:05" или "02:30:00"
	AgeLimit         int
	Rating           float64
	BoxOfficeRevenue float64
	ReleaseDate      time.Time
	GenreIDs         []string
}

type MovieFilters struct {
	Title string
	Genre string
}

func MovieFiltersFromDTO(dtoFilters dto.MovieFilters) MovieFilters {
	return MovieFilters{
		Title: dtoFilters.Title,
		Genre: dtoFilters.Genre,
	}
}

func CreateMovieFromDTO(req dto.CreateMovieRequest) Movie {
	releaseDate, _ := time.Parse("2006-01-02", req.ReleaseDate)
	return Movie{
		Title:            req.Title,
		Description:      req.Description,
		Duration:         req.Duration, // Ожидается формат "02:30:00"
		AgeLimit:         req.AgeLimit,
		BoxOfficeRevenue: req.BoxOfficeRevenue,
		ReleaseDate:      releaseDate,
		GenreIDs:         req.GenreIDs,
	}
}

func UpdateMovieFromDTO(req dto.UpdateMovieRequest) Movie {
	releaseDate, _ := time.Parse("2006-01-02", req.ReleaseDate)
	return Movie{
		Title:            req.Title,
		Description:      req.Description,
		Duration:         req.Duration,
		AgeLimit:         req.AgeLimit,
		BoxOfficeRevenue: req.BoxOfficeRevenue,
		ReleaseDate:      releaseDate,
		GenreIDs:         req.GenreIDs,
	}
}

func MovieToDTO(domainMovie Movie) dto.MovieResponse {
	return dto.MovieResponse{
		ID:               domainMovie.ID,
		Title:            domainMovie.Title,
		Description:      domainMovie.Description,
		Duration:         domainMovie.Duration,
		AgeLimit:         domainMovie.AgeLimit,
		Rating:           domainMovie.Rating,
		BoxOfficeRevenue: domainMovie.BoxOfficeRevenue,
		ReleaseDate:      domainMovie.ReleaseDate.Format("2006-01-02"),
	}
}

func MoviesToDTO(domainMovies []Movie) []dto.MovieResponse {
	dtos := make([]dto.MovieResponse, len(domainMovies))
	for i, movie := range domainMovies {
		dtos[i] = MovieToDTO(movie)
	}
	return dtos
}
