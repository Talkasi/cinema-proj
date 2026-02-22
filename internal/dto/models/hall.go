package dto

type HallFilters struct {
	Name         string `form:"name" example:"osnovnoy"`
	ScreenTypeID string `form:"screen_type_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Description  string `form:"description" example:"prostornyy"`
}

type HallResponse struct {
	ID           string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name         string `json:"name" example:"Zal 1"`
	Description  string `json:"description" example:"Osnovnoy kinozal s komfortnymi kreslami"`
	ScreenTypeID string `json:"screen_type_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type CreateHallRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=100" example:"Zal IMAX"`
	Description  string `json:"description" validate:"max=1000" example:"Zal s sistemoy IMAX dlya polnogo pogruzheniya v film"`
	ScreenTypeID string `json:"screen_type_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type UpdateHallRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=100" example:"Zal IMAX Premium"`
	Description  string `json:"description" validate:"max=1000" example:"Obnovlennyy zal IMAX s uluchshennoy akustikoy"`
	ScreenTypeID string `json:"screen_type_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}
