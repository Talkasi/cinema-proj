package dto

type SeatTypeFilters struct {
	Name        string `form:"name" example:"vip"`
	Description string `form:"description" example:"комфорт"`
}

type SeatTypeResponse struct {
	ID            string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name          string  `json:"name" example:"VIP"`
	Description   string  `json:"description" example:"Премиальные кресла с увеличенным пространством и дополнительным сервисом"`
	PriceModifier float64 `json:"price_modifier" example:"1.8"`
}

type CreateSeatTypeRequest struct {
	Name          string  `json:"name" validate:"required,min=1,max=100" example:"Стандарт"`
	Description   string  `json:"description" validate:"required,min=1,max=1000" example:"Стандартные комфортабельные кресла"`
	PriceModifier float64 `json:"price_modifier" validate:"required,min=0.1,max=10.0" example:"1.0"`
}

type UpdateSeatTypeRequest struct {
	Name          string  `json:"name" validate:"required,min=1,max=100" example:"Стандарт Плюс"`
	Description   string  `json:"description" validate:"required,min=1,max=1000" example:"Улучшенные стандартные кресла с дополнительным комфортом"`
	PriceModifier float64 `json:"price_modifier" validate:"required,min=0.1,max=10.0" example:"1.2"`
}
