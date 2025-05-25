package dto

import "cw/internal/domain"

type HallRequest struct {
	Name         string  `json:"name" example:"Зал 1"`
	ScreenTypeID string  `json:"screen_type_id" example:"de01f085-dffa-4347-88da-168560207511"`
	Description  *string `json:"description,omitempty" example:"Комфортабельный зал с современным оборудованием"`
}

type HallResponse struct {
	ID           string  `json:"id" example:"9b165097-1c9f-4ea3-bef0-e505baa4ff63"`
	Name         string  `json:"name" example:"Зал 1"`
	ScreenTypeID string  `json:"screen_type_id" example:"de01f085-dffa-4347-88da-168560207511"`
	Description  *string `json:"description,omitempty" example:"Комфортабельный зал с современным оборудованием"`
}

func (h HallRequest) ToDomain() domain.Hall {
	return domain.Hall{
		Name:         h.Name,
		ScreenTypeID: h.ScreenTypeID,
		Description:  h.Description,
	}
}

func HallFromDomain(h domain.Hall) HallResponse {
	return HallResponse{
		ID:           h.ID,
		Name:         h.Name,
		ScreenTypeID: h.ScreenTypeID,
		Description:  h.Description,
	}
}

func HallsFromDomainList(halls []domain.Hall) []HallResponse {
	result := make([]HallResponse, len(halls))
	for i, h := range halls {
		result[i] = HallFromDomain(h)
	}
	return result
}
