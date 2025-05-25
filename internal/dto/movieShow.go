package dto

import (
	"cw/internal/domain"
	"time"
)

type LanguageEnumTypeDto string

const (
	English LanguageEnumTypeDto = "English"
	Spanish LanguageEnumTypeDto = "Spanish"
	French  LanguageEnumTypeDto = "French"
	German  LanguageEnumTypeDto = "German"
	Italian LanguageEnumTypeDto = "Italian"
	Russian LanguageEnumTypeDto = "Русский"
)

func (l LanguageEnumTypeDto) IsValid() bool {
	switch l {
	case English, Spanish, French, German, Italian, Russian:
		return true
	}
	return false
}

type MovieShowResponse struct {
	ID        string              `json:"id" example:"9b165097-1c9f-4ea3-bef0-e505baa4ff63"`
	MovieID   string              `json:"movie_id" example:"1a2b3c4d-5e6f-7g8h-9i0j-k1l2m3n4o5p6"`
	HallID    string              `json:"hall_id" example:"de01f085-dffa-4347-88da-168560207511"`
	StartTime time.Time           `json:"start_time" example:"2023-10-01T14:30:00Z"`
	Language  LanguageEnumTypeDto `json:"language" example:"Русский"`
}

type MovieShowAdmin struct {
	MovieID   string              `json:"movie_id" example:"1a2b3c4d-5e6f-7g8h-9i0j-k1l2m3n4o5p6"`
	HallID    string              `json:"hall_id" example:"de01f085-dffa-4347-88da-168560207511"`
	StartTime time.Time           `json:"start_time" example:"2023-10-01T14:30:00Z"`
	Language  LanguageEnumTypeDto `json:"language" example:"Русский"`
	BasePrice float64             `json:"base_price" example:"300"`
}

type MovieShowRequest struct {
	MovieID   string              `json:"movie_id" example:"1a2b3c4d-5e6f-7g8h-9i0j-k1l2m3n4o5p6"`
	HallID    string              `json:"hall_id" example:"de01f085-dffa-4347-88da-168560207511"`
	StartTime time.Time           `json:"start_time" example:"2023-10-01T14:30:00Z"`
	Language  LanguageEnumTypeDto `json:"language" example:"Русский"`
}

func (m MovieShowRequest) ToDomain() domain.MovieShow {
	return domain.MovieShow{
		MovieID:   m.MovieID,
		HallID:    m.HallID,
		StartTime: m.StartTime,
		Language:  domain.LanguageEnumType(m.Language),
	}
}

func (m MovieShowAdmin) ToDomain() domain.MovieShow {
	return domain.MovieShow{
		MovieID:   m.MovieID,
		HallID:    m.HallID,
		StartTime: m.StartTime,
		Language:  domain.LanguageEnumType(m.Language),
		BasePrice: m.BasePrice,
	}
}

func MovieShowFromDomain(m domain.MovieShow) MovieShowResponse {
	return MovieShowResponse{
		ID:        m.ID,
		MovieID:   m.MovieID,
		HallID:    m.HallID,
		StartTime: m.StartTime,
		Language:  LanguageEnumTypeDto(m.Language),
	}
}

func MovieShowsFromDomainList(movieShows []domain.MovieShow) []MovieShowResponse {
	result := make([]MovieShowResponse, len(movieShows))
	for i, m := range movieShows {
		result[i] = MovieShowFromDomain(m)
	}
	return result
}
