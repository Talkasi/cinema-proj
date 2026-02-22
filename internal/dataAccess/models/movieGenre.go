package entity

type MovieGenre struct {
	MovieID string `db:"movie_id"`
	GenreID string `db:"genre_id"`
}
