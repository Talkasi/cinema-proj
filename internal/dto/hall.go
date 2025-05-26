package dto

type Hall struct {
	ID           string  `json:"id" example:"9b165097-1c9f-4ea3-bef0-e505baa4ff63"`
	Name         string  `json:"name" example:"Зал 1"`
	ScreenTypeID string  `json:"screen_type_id" example:"de01f085-dffa-4347-88da-168560207511"`
	Description  *string `json:"description,omitempty" example:"Комфортабельный зал с современным оборудованием"`
}

type HallData struct {
	Name         string  `json:"name" example:"Зал 1"`
	ScreenTypeID string  `json:"screen_type_id" example:"de01f085-dffa-4347-88da-168560207511"`
	Description  *string `json:"description,omitempty" example:"Комфортабельный зал с современным оборудованием"`
}
