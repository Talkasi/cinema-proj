package models

type ScreenTypeData struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"required,min=10,max=1000"`
}

type ScreenType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
