package dto

type ScreenTypeFilters struct {
	Name        string `form:"name" example:"imax"`
	Description string `form:"description" example:"bolshoy ekran"`
}

type ScreenTypeResponse struct {
	ID            string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name          string  `json:"name" example:"IMAX"`
	Description   string  `json:"description" example:"Tekhnologiya kinopokaza s uvelichennym razresheniem i uluchshennym zvukom"`
	PriceModifier float64 `json:"price_modifier" example:"1.5"`
}

type CreateScreenTypeRequest struct {
	Name          string  `json:"name" validate:"required,min=1,max=100" example:"4DX"`
	Description   string  `json:"description" validate:"required,min=1,max=1000" example:"Tekhnologiya s dvizhuschimisya kreslami i spetseffektami"`
	PriceModifier float64 `json:"price_modifier" validate:"required,min=0.1,max=10.0" example:"2.0"`
}

type UpdateScreenTypeRequest struct {
	Name          string  `json:"name" validate:"required,min=1,max=100" example:"4DX Premium"`
	Description   string  `json:"description" validate:"required,min=1,max=1000" example:"Uluchshennaya versiya 4DX s dopolnitelnymi effektami"`
	PriceModifier float64 `json:"price_modifier" validate:"required,min=0.1,max=10.0" example:"2.2"`
}
