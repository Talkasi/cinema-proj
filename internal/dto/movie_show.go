package dto

import "time"

type LanguageEnumType string

const (
	English LanguageEnumType = "English"
	Spanish LanguageEnumType = "Spanish"
	French  LanguageEnumType = "French"
	German  LanguageEnumType = "German"
	Italian LanguageEnumType = "Italian"
	Russian LanguageEnumType = "Русский"
)

func (l LanguageEnumType) IsValid() bool {
	switch l {
	case English, Spanish, French, German, Italian, Russian:
		return true
	}
	return false
}

type MovieShow struct {
	ID        string           `json:"id" example:"9b165097-1c9f-4ea3-bef0-e505baa4ff63"`
	MovieID   string           `json:"movie_id" example:"1a2b3c4d-5e6f-7g8h-9i0j-k1l2m3n4o5p6"`
	HallID    string           `json:"hall_id" example:"de01f085-dffa-4347-88da-168560207511"`
	StartTime time.Time        `json:"start_time" example:"2023-10-01T14:30:00Z"`
	Language  LanguageEnumType `json:"language" example:"Русский"`
}

type MovieShowAdmin struct {
	MovieID   string           `json:"movie_id" example:"1a2b3c4d-5e6f-7g8h-9i0j-k1l2m3n4o5p6"`
	HallID    string           `json:"hall_id" example:"de01f085-dffa-4347-88da-168560207511"`
	StartTime time.Time        `json:"start_time" example:"2023-10-01T14:30:00Z"`
	Language  LanguageEnumType `json:"language" example:"Русский"`
	BasePrice float64          `json:"base_price" example:"300"`
}

type MovieShowData struct {
	MovieID   string           `json:"movie_id" example:"1a2b3c4d-5e6f-7g8h-9i0j-k1l2m3n4o5p6"`
	HallID    string           `json:"hall_id" example:"de01f085-dffa-4347-88da-168560207511"`
	StartTime time.Time        `json:"start_time" example:"2023-10-01T14:30:00Z"`
	Language  LanguageEnumType `json:"language" example:"Русский"`
}
