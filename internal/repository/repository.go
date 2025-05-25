package repository

import (
	"context"
	"cw/internal/models"
	"cw/internal/utils"
	"time"
)

type UserRepository interface {
	Login(ctx context.Context, user models.UserLogin) (models.User, *utils.Error)
	Register(ctx context.Context, user models.UserRegister) (models.User, *utils.Error)
	GetByID(ctx context.Context, id string) (models.User, *utils.Error)
	GetAll(ctx context.Context) ([]models.User, *utils.Error)
	GetByEmail(ctx context.Context, email string) (models.User, *utils.Error)
	GetNicknameByID(ctx context.Context, id string) (string, *utils.Error)
	GetAdminStatus(ctx context.Context) (models.User, *utils.Error)
	UpdateAdminStatus(ctx context.Context, admin models.UserAdmin) (models.UserAdmin, *utils.Error)
	Update(ctx context.Context, u models.User) (models.User, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
}

type GenreRepository interface {
	GetAll(ctx context.Context) ([]models.Genre, *utils.Error)
	GetByID(ctx context.Context, id string) (models.Genre, *utils.Error)
	Create(ctx context.Context, g models.Genre) (models.Genre, *utils.Error)
	Update(ctx context.Context, g models.Genre) (models.Genre, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
	SearchByName(ctx context.Context, name string) ([]models.Genre, *utils.Error)
}

type HallRepository interface {
	GetAll(ctx context.Context) ([]models.Hall, *utils.Error)
	GetByID(ctx context.Context, id string) (models.Hall, *utils.Error)
	Create(ctx context.Context, h models.Hall) (models.Hall, *utils.Error)
	Update(ctx context.Context, h models.Hall) (models.Hall, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
	SearchByName(ctx context.Context, name string) ([]models.Hall, *utils.Error)
	GetByScreenType(ctx context.Context, screenTypeId string) ([]models.Hall, *utils.Error)
}

type MovieShowRepository interface {
	GetAll(ctx context.Context) ([]models.MovieShow, *utils.Error)
	GetByID(ctx context.Context, id string) (models.MovieShow, *utils.Error)
	Create(ctx context.Context, m models.MovieShow) (models.MovieShow, *utils.Error)
	Update(ctx context.Context, m models.MovieShow) (models.MovieShow, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
	GetByMovie(ctx context.Context, movieId string) ([]models.MovieShow, *utils.Error)
	GetByDate(ctx context.Context, date time.Time) ([]models.MovieShow, *utils.Error)
	GetUpcoming(ctx context.Context, hours int) ([]models.MovieShow, *utils.Error)
}

type MovieRepository interface {
	GetAll(ctx context.Context) ([]models.Movie, *utils.Error)
	GetByID(ctx context.Context, id string) (models.Movie, *utils.Error)
	Create(ctx context.Context, m models.Movie) (models.Movie, *utils.Error)
	Update(ctx context.Context, m models.Movie) (models.Movie, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
	SearchByName(ctx context.Context, name string) ([]models.Movie, *utils.Error)
	GetByGenres(ctx context.Context, genresIds []string) ([]models.Movie, *utils.Error)
}

type ScreenTypeRepository interface {
	GetAll(ctx context.Context) ([]models.ScreenType, *utils.Error)
	GetByID(ctx context.Context, id string) (models.ScreenType, *utils.Error)
	Create(ctx context.Context, t models.ScreenType) (models.ScreenType, *utils.Error)
	Update(ctx context.Context, t models.ScreenType) (models.ScreenType, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
	SearchByName(ctx context.Context, name string) ([]models.ScreenType, *utils.Error)
}

type SeatTypeRepository interface {
	GetAll(ctx context.Context) ([]models.SeatType, *utils.Error)
	GetByID(ctx context.Context, id string) (models.SeatType, *utils.Error)
	Create(ctx context.Context, t models.SeatType) (models.SeatType, *utils.Error)
	Update(ctx context.Context, t models.SeatType) (models.SeatType, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
	SearchByName(ctx context.Context, name string) ([]models.SeatType, *utils.Error)
}

type SeatRepository interface {
	GetAll(ctx context.Context) ([]models.Seat, *utils.Error)
	GetByID(ctx context.Context, id string) (models.Seat, *utils.Error)
	Create(ctx context.Context, s models.Seat) (models.Seat, *utils.Error)
	Update(ctx context.Context, s models.Seat) (models.Seat, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
	GetByHall(ctx context.Context, id string) ([]models.Seat, *utils.Error)
}

type TicketRepository interface {
	GetByMovieShowID(ctx context.Context, movieShowId string) ([]models.Ticket, *utils.Error)
	GetAvailableByMovieShowID(ctx context.Context, movieShowId string) ([]models.Ticket, *utils.Error)
	GetByUserId(ctx context.Context, userId string) ([]models.Ticket, *utils.Error)
	GetByID(ctx context.Context, id string) (models.Ticket, *utils.Error)
	Create(ctx context.Context, ticket models.Ticket) (models.Ticket, *utils.Error)
	Update(ctx context.Context, ticket models.Ticket) (models.Ticket, *utils.Error)
	Delete(ctx context.Context, id string) *utils.Error
}
