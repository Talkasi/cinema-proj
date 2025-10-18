package dto

type HallFilters struct {
	Name         string `form:"name" example:"основной"`
	ScreenTypeID string `form:"screen_type_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Description  string `form:"description" example:"просторный"`
}

type HallResponse struct {
	ID           string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name         string `json:"name" example:"Зал 1"`
	Description  string `json:"description" example:"Основной кинозал с комфортными креслами"`
	ScreenTypeID string `json:"screen_type_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type CreateHallRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=100" example:"Зал IMAX"`
	Description  string `json:"description" validate:"max=1000" example:"Зал с системой IMAX для полного погружения в фильм"`
	ScreenTypeID string `json:"screen_type_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type UpdateHallRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=100" example:"Зал IMAX Premium"`
	Description  string `json:"description" validate:"max=1000" example:"Обновленный зал IMAX с улучшенной акустикой"`
	ScreenTypeID string `json:"screen_type_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}
