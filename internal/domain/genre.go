package domain

type Genre struct {
	ID          string `json:"id" example:"ad2805ab-bf4c-4f93-ac68-2e0a854022f8"`
	Name        string `json:"name" example:"Исторический"`
	Description string `json:"description" example:"Жанр игрового кинематографа, повествующий о той или иной эпохе, людях и событиях прошлых лет"`
}

type GenreData struct {
	Name        string `json:"name" example:"Исторический"`
	Description string `json:"description" example:"Жанр игрового кинематографа, повествующий о той или иной эпохе, людях и событиях прошлых лет"`
}
