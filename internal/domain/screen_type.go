package domain

type ScreenType struct {
	ID            string  `json:"id" example:"de01f085-dffa-4347-88da-168560207511"`
	Name          string  `json:"name" example:"IMAX"`
	Description   string  `json:"description" example:"Экран с технологией IMAX для максимального погружения"`
	PriceModifier float64 `json:"price_modifier" example:"1"`
}
