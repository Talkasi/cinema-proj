package dto

import "cw/internal/domain"

type SeatResponse struct {
	ID         string `json:"id" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	HallID     string `json:"hall_id" example:"de01f085-dffa-4347-88da-168560207511"`
	SeatTypeID string `json:"seat_type_id" example:"premium"`
	RowNumber  int    `json:"row_number" example:"5"`
	SeatNumber int    `json:"seat_number" example:"12"`
}

type SeatRequest struct {
	HallID     string `json:"hall_id" example:"de01f085-dffa-4347-88da-168560207511"`
	SeatTypeID string `json:"seat_type_id" example:"a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"`
	RowNumber  int    `json:"row_number" example:"5"`
	SeatNumber int    `json:"seat_number" example:"12"`
}

func (st SeatRequest) ToDomain() domain.Seat {
	return domain.Seat{
		HallID:     st.HallID,
		SeatTypeID: st.SeatTypeID,
		RowNumber:  st.RowNumber,
		SeatNumber: st.SeatNumber,
	}
}

func SeatFromDomain(st domain.Seat) SeatResponse {
	return SeatResponse{
		ID:         st.ID,
		HallID:     st.HallID,
		SeatTypeID: st.SeatTypeID,
		RowNumber:  st.RowNumber,
		SeatNumber: st.SeatNumber,
	}
}

func SeatFromDomainList(screenTypes []domain.Seat) []SeatResponse {
	result := make([]SeatResponse, len(screenTypes))
	for i, s := range screenTypes {
		result[i] = SeatFromDomain(s)
	}
	return result
}
