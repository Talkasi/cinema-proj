package dto

type ReviewFilters struct {
	MovieID   string `form:"movie_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID    string `form:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	RatingMin int    `form:"rating_min" example:"7"`
	RatingMax int    `form:"rating_max" example:"10"`
	Comment   string `form:"comment" example:"otlichnyy"`
}

type ReviewResponse struct {
	ID      string `json:"id" example:"550e8400-e29b-41d4-a716-446655440002"`
	MovieID string `json:"movie_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID  string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Rating  int    `json:"rating" example:"9"`
	Comment string `json:"comment" example:"Otlichnyy film s zakhvatyvayuschim syuzhetom i velikolepnoy akterskoy igroy!"`
}

type CreateReviewRequest struct {
	MovieID string `json:"movie_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID  string `json:"user_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	Rating  int    `json:"rating" validate:"required,min=1,max=10" example:"8"`
	Comment string `json:"comment" validate:"max=1000" example:"Ochen ponravilas operatorskaya rabota i saundtrek"`
}

type UpdateReviewRequest struct {
	MovieID string `json:"movie_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID  string `json:"user_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	Rating  int    `json:"rating" validate:"required,min=1,max=10" example:"9"`
	Comment string `json:"comment" validate:"max=1000" example:"Posle povtornogo prosmotra povyshayu otsenku - film stal esche luchshe!"`
}
