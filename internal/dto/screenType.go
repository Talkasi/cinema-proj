package dto

import "cw/internal/domain"

type ScreenTypeResponse struct {
	ID          string `json:"id" example:"de01f085-dffa-4347-88da-168560207511"`
	Name        string `json:"name" example:"IMAX"`
	Description string `json:"description" example:"Экран с технологией IMAX для максимального погружения"`
}

type ScreenTypeRequest struct {
	Name        string `json:"name" example:"IMAX"`
	Description string `json:"description" example:"Экран с технологией IMAX для максимального погружения"`
}

type ScreenTypeAdmin struct {
	Name          string  `json:"name" example:"IMAX"`
	Description   string  `json:"description" example:"Экран с технологией IMAX для максимального погружения"`
	PriceModifier float64 `json:"price_modifier" example:"1"`
}

func (st ScreenTypeRequest) ToDomain() domain.ScreenType {
	return domain.ScreenType{
		Name:        st.Name,
		Description: st.Description,
	}
}

func (st ScreenTypeAdmin) ToDomain() domain.ScreenType {
	return domain.ScreenType{
		Name:          st.Name,
		Description:   st.Description,
		PriceModifier: st.PriceModifier,
	}
}

func ScreenTypeFromDomain(st domain.ScreenType) ScreenTypeResponse {
	return ScreenTypeResponse{
		ID:          st.ID,
		Name:        st.Name,
		Description: st.Description,
	}
}

func ScreenTypeFromDomainList(screenTypes []domain.ScreenType) []ScreenTypeResponse {
	result := make([]ScreenTypeResponse, len(screenTypes))
	for i, s := range screenTypes {
		result[i] = ScreenTypeFromDomain(s)
	}
	return result
}
