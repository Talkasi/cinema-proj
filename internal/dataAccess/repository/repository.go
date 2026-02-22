// internal/repository/interfaces.go
package repository

import (
	"context"
	domain "cw/internal/domain/models"
	"cw/internal/utils"
)

type UserRepository interface {
	Login(ctx context.Context, credentials domain.User) (domain.AuthResponse, *utils.Error)
	Register(ctx context.Context, user domain.User) (domain.User, *utils.Error)
	GetAll(ctx context.Context, filters domain.UserFilters, page, limit int) ([]domain.User, int, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.User, *utils.Error)
	Update(ctx context.Context, id string, user domain.User) (domain.User, *utils.Error)
	UpdateAdminStatus(ctx context.Context, id string, isAdmin bool) (domain.User, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
}

type GenreRepository interface {
	GetAll(ctx context.Context, filters domain.GenreFilters, page, limit int) ([]domain.Genre, int, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.Genre, *utils.Error)
	Create(ctx context.Context, genre domain.Genre) (domain.Genre, *utils.Error)
	Update(ctx context.Context, id string, genre domain.Genre) (domain.Genre, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
}

type HallRepository interface {
	GetAll(ctx context.Context, filters domain.HallFilters, page, limit int) ([]domain.Hall, int, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.Hall, *utils.Error)
	Create(ctx context.Context, hall domain.Hall) (domain.Hall, *utils.Error)
	Update(ctx context.Context, id string, hall domain.Hall) (domain.Hall, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
}

type MovieShowRepository interface {
	GetAll(ctx context.Context, filters domain.MovieShowFilters, page, limit int) ([]domain.MovieShow, int, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.MovieShow, *utils.Error)
	Create(ctx context.Context, movieShow domain.MovieShow) (domain.MovieShow, *utils.Error)
	Update(ctx context.Context, id string, movieShow domain.MovieShow) (domain.MovieShow, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
}

type MovieRepository interface {
	GetAll(ctx context.Context, filters domain.MovieFilters, page, limit int) ([]domain.Movie, int, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.Movie, *utils.Error)
	Create(ctx context.Context, movie domain.Movie) (domain.Movie, *utils.Error)
	Update(ctx context.Context, id string, movie domain.Movie) (domain.Movie, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
	CalculateMovieRating(ctx context.Context, movieID string) (float64, *utils.Error)
}

type ScreenTypeRepository interface {
	GetAll(ctx context.Context, filters domain.ScreenTypeFilters, page, limit int) ([]domain.ScreenType, int, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.ScreenType, *utils.Error)
	Create(ctx context.Context, screenType domain.ScreenType) (domain.ScreenType, *utils.Error)
	Update(ctx context.Context, id string, screenType domain.ScreenType) (domain.ScreenType, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
}

type SeatTypeRepository interface {
	GetAll(ctx context.Context, filters domain.SeatTypeFilters, page, limit int) ([]domain.SeatType, int, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.SeatType, *utils.Error)
	Create(ctx context.Context, seatType domain.SeatType) (domain.SeatType, *utils.Error)
	Update(ctx context.Context, id string, seatType domain.SeatType) (domain.SeatType, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
}

type SeatRepository interface {
	GetByHall(ctx context.Context, hallId string, filters domain.SeatFilters, page, limit int) ([]domain.Seat, int, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.Seat, *utils.Error)
	Create(ctx context.Context, seat domain.Seat) (domain.Seat, *utils.Error)
	Update(ctx context.Context, id string, seat domain.Seat) (domain.Seat, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
}

type TicketRepository interface {
	GetAll(ctx context.Context, filters domain.TicketFilters, page, limit int) ([]domain.Ticket, int, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.Ticket, *utils.Error)
	CreateForMovieShow(ctx context.Context, movieShowId string, ticket domain.Ticket) (domain.Ticket, *utils.Error)
	UpdateStatus(ctx context.Context, id string, StatusData domain.Ticket) (domain.Ticket, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
}

type ReviewRepository interface {
	GetAll(ctx context.Context, filters domain.ReviewFilters, page, limit int) ([]domain.Review, int, *utils.Error)
	CreateForMovie(ctx context.Context, movieId string, review domain.Review) (domain.Review, *utils.Error)
	Update(ctx context.Context, id string, review domain.Review) (domain.Review, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
}
