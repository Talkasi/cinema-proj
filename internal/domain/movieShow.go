package domain

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
	ID        string
	MovieID   string
	HallID    string
	StartTime time.Time
	Language  LanguageEnumType
	BasePrice float64
}
