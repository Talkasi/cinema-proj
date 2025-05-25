package dto

import "cw/internal/domain"

type GenreRequest struct {
	Name        string `json:"name" example:"Исторический"`
	Description string `json:"description" example:"Жанр игрового кинематографа, повествующий о той или иной эпохе, людях и событиях прошлых лет"`
}

type GenreResponse struct {
	ID          string `json:"id" example:"ad2805ab-bf4c-4f93-ac68-2e0a854022f8"`
	Name        string `json:"name" example:"Исторический"`
	Description string `json:"description" example:"Жанр игрового кинематографа, повествующий о той или иной эпохе, людях и событиях прошлых лет"`
}

func (g GenreRequest) ToDomain() domain.Genre {
	return domain.Genre{
		Name:        g.Name,
		Description: g.Description,
	}
}

func GenreFromDomain(g domain.Genre) GenreResponse {
	return GenreResponse{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
	}
}

func GenresFromDomainList(genres []domain.Genre) []GenreResponse {
	result := make([]GenreResponse, len(genres))
	for i, g := range genres {
		result[i] = GenreFromDomain(g)
	}
	return result
}
