package dto

type SeatTypeFilters struct {
	Name        string `form:"name" example:"vip"`
	Description string `form:"description" example:"komfort"`
}

type SeatTypeResponse struct {
	ID            string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name          string  `json:"name" example:"VIP"`
	Description   string  `json:"description" example:"Premialnye kresla s uvelichennym prostranstvom i dopolnitelnym servisom"`
	PriceModifier float64 `json:"price_modifier" example:"1.8"`
}

type CreateSeatTypeRequest struct {
	Name          string  `json:"name" validate:"required,min=1,max=100" example:"Standart"`
	Description   string  `json:"description" validate:"required,min=1,max=1000" example:"Standartnye komfortabelnye kresla"`
	PriceModifier float64 `json:"price_modifier" validate:"required,min=0.1,max=10.0" example:"1.0"`
}

type UpdateSeatTypeRequest struct {
	Name          string  `json:"name" validate:"required,min=1,max=100" example:"Standart Plyus"`
	Description   string  `json:"description" validate:"required,min=1,max=1000" example:"Uluchshennye standartnye kresla s dopolnitelnym komfortom"`
	PriceModifier float64 `json:"price_modifier" validate:"required,min=0.1,max=10.0" example:"1.2"`
}
