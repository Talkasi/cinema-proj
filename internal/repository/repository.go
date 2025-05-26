package repository

import (
	"context"
	"cw/internal/domain"
	"cw/internal/utils"
	"time"
)

type UserRepository interface {
	Login(ctx context.Context, user domain.User) (domain.User, *utils.Error)
	Register(ctx context.Context, user domain.User) (domain.User, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.User, *utils.Error)
	GetAll(ctx context.Context) ([]domain.User, *utils.Error)
	GetByEmail(ctx context.Context, email string) (domain.User, *utils.Error)
	GetNicknameByID(ctx context.Context, id string) (string, *utils.Error)
	GetAdminStatus(ctx context.Context) (domain.User, *utils.Error)
	UpdateAdminStatus(ctx context.Context, admin domain.User) (domain.User, *utils.Error)
	Update(ctx context.Context, u domain.User) (domain.User, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
}

type GenreRepository interface {
	GetAll(ctx context.Context) ([]domain.Genre, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.Genre, *utils.Error)
	Create(ctx context.Context, g domain.Genre) (domain.Genre, *utils.Error)
	Update(ctx context.Context, g domain.Genre) (domain.Genre, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
	SearchByName(ctx context.Context, name string) ([]domain.Genre, *utils.Error)
}

type HallRepository interface {
	GetAll(ctx context.Context) ([]domain.Hall, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.Hall, *utils.Error)
	Create(ctx context.Context, h domain.Hall) (domain.Hall, *utils.Error)
	Update(ctx context.Context, h domain.Hall) (domain.Hall, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
	SearchByName(ctx context.Context, name string) ([]domain.Hall, *utils.Error)
	GetByScreenType(ctx context.Context, screenTypeId string) ([]domain.Hall, *utils.Error)
}

type MovieShowRepository interface {
	GetAll(ctx context.Context) ([]domain.MovieShow, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.MovieShow, *utils.Error)
	Create(ctx context.Context, m domain.MovieShow) (domain.MovieShow, *utils.Error)
	Update(ctx context.Context, m domain.MovieShow) (domain.MovieShow, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
	GetByMovie(ctx context.Context, movieId string) ([]domain.MovieShow, *utils.Error)
	GetByDate(ctx context.Context, date time.Time) ([]domain.MovieShow, *utils.Error)
	GetUpcoming(ctx context.Context, hours int) ([]domain.MovieShow, *utils.Error)
}

type MovieRepository interface {
	GetAll(ctx context.Context) ([]domain.Movie, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.Movie, *utils.Error)
	Create(ctx context.Context, m domain.Movie) (domain.Movie, *utils.Error)
	Update(ctx context.Context, m domain.Movie) (domain.Movie, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
	SearchByName(ctx context.Context, name string) ([]domain.Movie, *utils.Error)
	GetByGenres(ctx context.Context, genresIds []string) ([]domain.Movie, *utils.Error)
}

type ScreenTypeRepository interface {
	GetAll(ctx context.Context) ([]domain.ScreenType, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.ScreenType, *utils.Error)
	Create(ctx context.Context, t domain.ScreenType) (domain.ScreenType, *utils.Error)
	Update(ctx context.Context, t domain.ScreenType) (domain.ScreenType, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
	SearchByName(ctx context.Context, name string) ([]domain.ScreenType, *utils.Error)
}

type SeatTypeRepository interface {
	GetAll(ctx context.Context) ([]domain.SeatType, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.SeatType, *utils.Error)
	Create(ctx context.Context, t domain.SeatType) (domain.SeatType, *utils.Error)
	Update(ctx context.Context, t domain.SeatType) (domain.SeatType, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
	SearchByName(ctx context.Context, name string) ([]domain.SeatType, *utils.Error)
}

type SeatRepository interface {
	GetAll(ctx context.Context) ([]domain.Seat, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.Seat, *utils.Error)
	Create(ctx context.Context, s domain.Seat) (domain.Seat, *utils.Error)
	Update(ctx context.Context, s domain.Seat) (domain.Seat, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
	GetByHall(ctx context.Context, id string) ([]domain.Seat, *utils.Error)
}

type TicketRepository interface {
	GetByMovieShowID(ctx context.Context, movieShowId string) ([]domain.Ticket, *utils.Error)
	GetAvailableByMovieShowID(ctx context.Context, movieShowId string) ([]domain.Ticket, *utils.Error)
	GetByUserId(ctx context.Context, userId string) ([]domain.Ticket, *utils.Error)
	GetByID(ctx context.Context, id string) (domain.Ticket, *utils.Error)
	Create(ctx context.Context, ticket domain.Ticket) (domain.Ticket, *utils.Error)
	Update(ctx context.Context, ticket domain.Ticket) (domain.Ticket, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
}
