package models

// Структура для создания и обновлёния жанра
type GenreData struct {
	Name        string `json:"name" validate:"required,min=1,max=64,nameFormat"`
	Description string `json:"description" validate:"required,min=1,max=1000"`
}

type Genre struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
