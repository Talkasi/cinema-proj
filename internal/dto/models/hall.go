package dto

type HallFilters struct {
	Name         string `form:"name" example:"Main cinema hall with comfortable seats"`
	ScreenTypeID string `form:"screen_type_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Description  string `form:"description" example:"Spacious hall with premium seating"`
}

type HallResponse struct {
	ID           string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name         string `json:"name" example:"Main hall with improved acoustics"`
	Description  string `json:"description" example:"Main cinema hall with comfortable seats"`
	ScreenTypeID string `json:"screen_type_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type CreateHallRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=100" example:"IMAX Hall"`
	Description  string `json:"description" validate:"max=1000" example:"IMAX hall with immersive viewing experience"`
	ScreenTypeID string `json:"screen_type_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type UpdateHallRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=100" example:"IMAX Premium Hall"`
	Description  string `json:"description" validate:"max=1000" example:"Upgraded IMAX hall with improved acoustics"`
	ScreenTypeID string `json:"screen_type_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}
