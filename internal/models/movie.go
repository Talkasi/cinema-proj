package models

type Movie struct {
	ID               string   `json:"id" example:"9b165097-1c9f-4ea3-bef0-e505baa4ff63"`
	Title            string   `json:"title" example:"Властелин колец"`
	Duration         string   `json:"duration" example:"02:58:00"`
	Rating           *float64 `json:"rating,omitempty" example:"8.8"`
	Description      string   `json:"description" example:"Эпическая история о кольце власти."`
	AgeLimit         int      `json:"age_limit" example:"12"`
	BoxOfficeRevenue float64  `json:"box_office_revenue" example:"300000000"`
	ReleaseDate      string   `json:"release_date" example:"2001-12-19"`
	Genres           []Genre  `json:"genres"`
}

type MovieData struct {
	Title       string   `json:"title" example:"Властелин колец"`
	Duration    string   `json:"duration" example:"02:58:00"`
	Description string   `json:"description" example:"Эпическая история о кольце власти."`
	AgeLimit    int      `json:"age_limit" example:"12"`
	ReleaseDate string   `json:"release_date" example:"2001-12-19"`
	GenreIDs    []string `json:"genre_ids" example:"[\"f297eeaf-e784-43bf-a068-eef84f75baa4\", \"c5c8e037-a073-4105-9941-21e1cb4e79dd\"]"`
}
