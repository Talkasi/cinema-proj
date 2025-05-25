package domain

type Movie struct {
	ID               string
	Title            string
	Duration         string
	Rating           *float64
	Description      string
	AgeLimit         int
	BoxOfficeRevenue float64
	ReleaseDate      string
	Genres           []Genre
}
