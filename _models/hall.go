package models

type HallData struct {
	ScreenTypeID string  `json:"screen_type_id" validate:"required,uuid4"`
	Name         string  `json:"name" validate:"required,min=2,max=100"`
	Capacity     int     `json:"capacity" validate:"required,min=1"`
	Description  *string `json:"description,omitempty" validate:"omitempty,min=10,max=1000"`
}

type HallResponse struct {
	ID          string     `json:"id"`
	ScreenType  ScreenType `json:"screen_type"`
	Name        string     `json:"name"`
	Capacity    int        `json:"capacity"`
	Description *string    `json:"description,omitempty"`
}

type Hall struct {
	ID           string  `json:"id"`
	ScreenTypeID string  `json:"screen_type_id"`
	Name         string  `json:"name"`
	Capacity     int     `json:"capacity"`
	Description  *string `json:"description,omitempty"`
}
