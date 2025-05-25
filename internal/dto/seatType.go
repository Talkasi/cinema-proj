package dto

import "cw/internal/domain"

type SeatTypeResponse struct {
	ID          string `json:"id" example:"de01f085-dffa-4347-88da-168560207511"`
	Name        string `json:"name" example:"Премиум"`
	Description string `json:"description" example:"Комфортабельные места с дополнительным пространством и удобствами"`
}

type SeatTypeRequest struct {
	Name        string `json:"name" example:"Премиум"`
	Description string `json:"description" example:"Комфортабельные места с дополнительным пространством и удобствами"`
}

type SeatTypeAdmin struct {
	Name          string  `json:"name" example:"Премиум"`
	Description   string  `json:"description" example:"Комфортабельные места с дополнительным пространством и удобствами"`
	PriceModifier float64 `json:"price_modifier" example:"1"`
}

func (st SeatTypeRequest) ToDomain() domain.SeatType {
	return domain.SeatType{
		Name:        st.Name,
		Description: st.Description,
	}
}

func (st SeatTypeAdmin) ToDomain() domain.SeatType {
	return domain.SeatType{
		Name:          st.Name,
		Description:   st.Description,
		PriceModifier: st.PriceModifier,
	}
}

func SeatTypeFromDomain(st domain.SeatType) SeatTypeResponse {
	return SeatTypeResponse{
		ID:          st.ID,
		Name:        st.Name,
		Description: st.Description,
	}
}

func SeatTypeFromDomainList(screenTypes []domain.SeatType) []SeatTypeResponse {
	result := make([]SeatTypeResponse, len(screenTypes))
	for i, s := range screenTypes {
		result[i] = SeatTypeFromDomain(s)
	}
	return result
}
