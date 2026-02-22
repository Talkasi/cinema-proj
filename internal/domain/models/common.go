package domain

type AuthResponse struct {
	Token  string
	UserID string
}

type PaginatedResponse[T any] struct {
	Data  []T
	Total int
	Page  int
	Limit int
}
