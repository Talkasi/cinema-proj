package domain

type AuthResponse struct {
	Token        string
	UserID       string
	Message      string
	TwoFAEnabled bool
}

type PaginatedResponse[T any] struct {
	Data  []T
	Total int
	Page  int
	Limit int
}
