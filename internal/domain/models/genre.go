package domain

import dto "cw/internal/dto/models"

type Genre struct {
	ID          string
	Name        string
	Description string
}

type GenreFilters struct {
	Name        string
	Description string
}

func GenreFiltersFromDTO(dtoGenreFilters dto.GenreFilters) GenreFilters {
	return GenreFilters{
		Name:        dtoGenreFilters.Name,
		Description: dtoGenreFilters.Description,
	}
}

func GenreToDTO(domainGenre Genre) dto.GenreResponse {
	return dto.GenreResponse{
		ID:          domainGenre.ID,
		Name:        domainGenre.Name,
		Description: domainGenre.Description,
	}
}

func GenreFromDTO(dtoGenre dto.GenreResponse) Genre {
	return Genre{
		ID:          dtoGenre.ID,
		Name:        dtoGenre.Name,
		Description: dtoGenre.Description,
	}
}

func CreateGenreFromDTO(req dto.CreateGenreRequest) Genre {
	return Genre{
		Name:        req.Name,
		Description: req.Description,
	}
}

func UpdateGenreFromDTO(req dto.UpdateGenreRequest) Genre {
	return Genre{
		Name:        req.Name,
		Description: req.Description,
	}
}

func GenresToDTO(domainGenres []Genre) []dto.GenreResponse {
	dtos := make([]dto.GenreResponse, len(domainGenres))
	for i, genre := range domainGenres {
		dtos[i] = GenreToDTO(genre)
	}
	return dtos
}
