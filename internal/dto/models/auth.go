package dto

type AuthResponse struct {
	Token       string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNTUwZTg0MDAtZTI5Yi00MWQ0LWE3MTYtNDQ2NjU1NDQwMDAwIiwiaXNfYWRtaW4iOmZhbHNlLCJleHAiOjE3MDAwMDAwMDB9"`
	UserID      string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Message     string `json:"message" example:"Two-factor authentication required. Check your email for the verification code."`
	TwoFANeeded bool   `json:"two_fa_needed" example:"true"`
}

type FirstStepAuthResponse struct {
	Message string `json:"message" example:"Two-factor authentication required. Check your email for the verification code."`
}

type Verify2FARequest struct {
	Code   string `json:"code" validate:"required,len=6" example:"123456"`
	UserID string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type TwoFAInfoResponse struct {
	Enabled bool `json:"enabled" example:"true"`
}

type CreateResponse struct {
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type ErrorResponse struct {
	Message string `json:"message" example:"Proizoshla oshibka pri obrabotke zaprosa"`
}

type PaginatedResponse[T any] struct {
	Data  []T `json:"data"`
	Total int `json:"total" example:"150"`
	Page  int `json:"page" example:"1"`
	Limit int `json:"limit" example:"20"`
}

type PaginatedUserResponse = PaginatedResponse[UserResponse]
type PaginatedHallResponse = PaginatedResponse[HallResponse]
type PaginatedMovieResponse = PaginatedResponse[MovieResponse]
type PaginatedMovieShowResponse = PaginatedResponse[MovieShowResponse]
type PaginatedReviewResponse = PaginatedResponse[ReviewResponse]
type PaginatedScreenTypeResponse = PaginatedResponse[ScreenTypeResponse]
type PaginatedSeatResponse = PaginatedResponse[SeatResponse]
type PaginatedSeatTypeResponse = PaginatedResponse[SeatTypeResponse]
type PaginatedTicketResponse = PaginatedResponse[TicketResponse]
type PaginatedGenreResponse = PaginatedResponse[GenreResponse]
