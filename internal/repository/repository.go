package repository

import (
	"context"
	"cw/internal/domain"
	"time"
)

type UserRepository interface {
	Login(ctx context.Context, user domain.User) (domain.User, error)
	Register(ctx context.Context, user domain.User) (domain.User, error)
	GetByID(ctx context.Context, id string) (domain.User, error)
	GetAll(ctx context.Context) ([]domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetNicknameByID(ctx context.Context, id string) (string, error)
	GetAdminStatus(ctx context.Context) (domain.User, error)
	UpdateAdminStatus(ctx context.Context, admin domain.User) (domain.User, error)
	Update(ctx context.Context, u domain.User) (domain.User, error)
	Delete(ctx context.Context, id string) error
}

type GenreRepository interface {
	GetAll(ctx context.Context) ([]domain.Genre, error)
	GetByID(ctx context.Context, id string) (domain.Genre, error)
	Create(ctx context.Context, g domain.Genre) (domain.Genre, error)
	Update(ctx context.Context, g domain.Genre) (domain.Genre, error)
	Delete(ctx context.Context, id string) error
	SearchByName(ctx context.Context, name string) ([]domain.Genre, error)
}

type HallRepository interface {
	GetAll(ctx context.Context) ([]domain.Hall, error)
	GetByID(ctx context.Context, id string) (domain.Hall, error)
	Create(ctx context.Context, h domain.Hall) (domain.Hall, error)
	Update(ctx context.Context, h domain.Hall) (domain.Hall, error)
	Delete(ctx context.Context, id string) error
	SearchByName(ctx context.Context, name string) ([]domain.Hall, error)
	GetByScreenType(ctx context.Context, screenTypeId string) ([]domain.Hall, error)
}

type MovieShowRepository interface {
	GetAll(ctx context.Context) ([]domain.MovieShow, error)
	GetByID(ctx context.Context, id string) (domain.MovieShow, error)
	Create(ctx context.Context, m domain.MovieShow) (domain.MovieShow, error)
	Update(ctx context.Context, m domain.MovieShow) (domain.MovieShow, error)
	Delete(ctx context.Context, id string) error
	GetByMovie(ctx context.Context, movieId string) ([]domain.MovieShow, error)
	GetByDate(ctx context.Context, date time.Time) ([]domain.MovieShow, error)
	GetUpcoming(ctx context.Context, hours int) ([]domain.MovieShow, error)
}

type MovieRepository interface {
	GetAll(ctx context.Context) ([]domain.Movie, error)
	GetByID(ctx context.Context, id string) (domain.Movie, error)
	Create(ctx context.Context, m domain.Movie) (domain.Movie, error)
	Update(ctx context.Context, m domain.Movie) (domain.Movie, error)
	Delete(ctx context.Context, id string) error
	SearchByName(ctx context.Context, name string) ([]domain.Movie, error)
	GetByGenres(ctx context.Context, genresIds []string) ([]domain.Movie, error)
}

type ScreenTypeRepository interface {
	GetAll(ctx context.Context) ([]domain.ScreenType, error)
	GetByID(ctx context.Context, id string) (domain.ScreenType, error)
	Create(ctx context.Context, t domain.ScreenType) (domain.ScreenType, error)
	Update(ctx context.Context, t domain.ScreenType) (domain.ScreenType, error)
	Delete(ctx context.Context, id string) error
	SearchByName(ctx context.Context, name string) ([]domain.ScreenType, error)
}

type SeatTypeRepository interface {
	GetAll(ctx context.Context) ([]domain.SeatType, error)
	GetByID(ctx context.Context, id string) (domain.SeatType, error)
	Create(ctx context.Context, t domain.SeatType) (domain.SeatType, error)
	Update(ctx context.Context, t domain.SeatType) (domain.SeatType, error)
	Delete(ctx context.Context, id string) error
	SearchByName(ctx context.Context, name string) ([]domain.SeatType, error)
}

type SeatRepository interface {
	GetAll(ctx context.Context) ([]domain.Seat, error)
	GetByID(ctx context.Context, id string) (domain.Seat, error)
	Create(ctx context.Context, s domain.Seat) (domain.Seat, error)
	Update(ctx context.Context, s domain.Seat) (domain.Seat, error)
	Delete(ctx context.Context, id string) error
	GetByHall(ctx context.Context, id string) ([]domain.Seat, error)
}

type TicketRepository interface {
	GetByMovieShowID(ctx context.Context, movieShowId string) ([]domain.Ticket, error)
	GetAvailableByMovieShowID(ctx context.Context, movieShowId string) ([]domain.Ticket, error)
	GetByUserId(ctx context.Context, userId string) ([]domain.Ticket, error)
	GetByID(ctx context.Context, id string) (domain.Ticket, error)
	Create(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error)
	Update(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error)
	Delete(ctx context.Context, id string) error
}
