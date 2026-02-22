package dto

type MovieFilters struct {
	Title string `form:"title" example:"interstellar"`
	Genre string `form:"genre" example:"fantastika"`
}

type MovieResponse struct {
	ID               string          `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title            string          `json:"title" example:"Interstellar"`
	Description      string          `json:"description" example:"Fantasticheskiy epos o puteshestvii cherez chervotochinu v poiskakh novogo doma dlya chelovechestva"`
	Duration         string          `json:"duration" example:"02:49:00"`
	AgeLimit         int             `json:"age_limit" example:"12"`
	Rating           float64         `json:"rating" example:"8.6"`
	BoxOfficeRevenue float64         `json:"box_office_revenue" example:"677471339"`
	ReleaseDate      string          `json:"release_date" example:"2014-10-26"`
	Genres           []GenreResponse `json:"genres"`
}

type CreateMovieRequest struct {
	Title            string   `json:"title" validate:"required,min=1,max=200" example:"Nachalo"`
	Description      string   `json:"description" validate:"required,min=1,max=1000" example:"Triller o proniknovenii v sny s tselyu krazhi idey"`
	Duration         string   `json:"duration" validate:"required" example:"02:28:00"`
	AgeLimit         int      `json:"age_limit" validate:"required,min=0,max=21" example:"12"`
	BoxOfficeRevenue float64  `json:"box_office_revenue" validate:"min=0" example:"836836967"`
	ReleaseDate      string   `json:"release_date" validate:"required" example:"2010-07-08"`
	GenreIDs         []string `json:"genre_ids" example:"550e8400-e29b-41d4-a716-446655440000,550e8400-e29b-41d4-a716-446655440001"`
}

type UpdateMovieRequest struct {
	Title            string   `json:"title" validate:"required,min=1,max=200" example:"Nachalo (rezhisserskaya versiya)"`
	Description      string   `json:"description" validate:"required,min=1,max=1000" example:"Rasshirennaya versiya trillera o proniknovenii v sny"`
	Duration         string   `json:"duration" validate:"required" example:"02:48:00"`
	AgeLimit         int      `json:"age_limit" validate:"required,min=0,max=21" example:"16"`
	BoxOfficeRevenue float64  `json:"box_office_revenue" validate:"min=0" example:"900000000"`
	ReleaseDate      string   `json:"release_date" validate:"required" example:"2010-07-16"`
	GenreIDs         []string `json:"genre_ids" example:"550e8400-e29b-41d4-a716-446655440000,550e8400-e29b-41d4-a716-446655440002"`
}
