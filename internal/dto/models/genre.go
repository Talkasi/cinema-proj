package dto

type GenreFilters struct {
	Name        string `form:"name" example:"Example value"`
	Description string `form:"description" example:"emotsionalnyy"`
}

type CreateGenreRequest struct {
	Name        string `json:"name" validate:"required" example:"Fantastika"`
	Description string `json:"description" validate:"required" example:"Filmy o buduschem, tekhnologiyakh i kosmose"`
}

type UpdateGenreRequest struct {
	Name        string `json:"name" validate:"required" example:"Nauchnaya fantastika"`
	Description string `json:"description" validate:"required" example:"Filmy, osnovannye na nauchnykh kontseptsiyakh i tekhnologiyakh buduschego"`
}

type GenreResponse struct {
	ID          string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string `json:"name" example:"Fantastika"`
	Description string `json:"description" example:"Filmy o buduschem, tekhnologiyakh i kosmose"`
}

type GenreDataRequest struct {
	Name        string `json:"name" validate:"required" example:"Example value"`
	Description string `json:"description" example:"Emotsionalnye filmy o chelovecheskikh otnosheniyakh"`
}
