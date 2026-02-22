package dto

type GenreFilters struct {
	Name        string `form:"name" example:"драма"`
	Description string `form:"description" example:"эмоциональный"`
}

type CreateGenreRequest struct {
	Name        string `json:"name" validate:"required" example:"Фантастика"`
	Description string `json:"description" validate:"required" example:"Фильмы о будущем, технологиях и космосе"`
}

type UpdateGenreRequest struct {
	Name        string `json:"name" validate:"required" example:"Научная фантастика"`
	Description string `json:"description" validate:"required" example:"Фильмы, основанные на научных концепциях и технологиях будущего"`
}

type GenreResponse struct {
	ID          string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string `json:"name" example:"Фантастика"`
	Description string `json:"description" example:"Фильмы о будущем, технологиях и космосе"`
}

type GenreDataRequest struct {
	Name        string `json:"name" validate:"required" example:"Драма"`
	Description string `json:"description" example:"Эмоциональные фильмы о человеческих отношениях"`
}
