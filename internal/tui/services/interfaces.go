package services

import (
	"context"
	dto "cw/internal/dto/models"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, error)
	Register(ctx context.Context, user dto.CreateUserRequest) (string, error)
}

type GenreService interface {
	GetAll(ctx context.Context, filters dto.GenreFilters, page, limit int) (*dto.PaginatedGenreResponse, error)
	GetByID(ctx context.Context, id string) (*dto.GenreResponse, error)
	Create(ctx context.Context, genre dto.CreateGenreRequest) (string, error)
	Update(ctx context.Context, id string, genre dto.UpdateGenreRequest) error
	Delete(ctx context.Context, id string) error
}

type MovieService interface {
	GetAll(ctx context.Context, filters dto.MovieFilters, page, limit int) (*dto.PaginatedMovieResponse, error)
	GetByID(ctx context.Context, id string) (*dto.MovieResponse, error)
	Create(ctx context.Context, movie dto.CreateMovieRequest) (string, error)
	Update(ctx context.Context, id string, movie dto.UpdateMovieRequest) error
	Delete(ctx context.Context, id string) error
}

type MovieShowService interface {
	GetAll(ctx context.Context, filters dto.MovieShowFilters, page, limit int) (*dto.PaginatedMovieShowResponse, error)
	GetByID(ctx context.Context, id string) (*dto.MovieShowResponse, error)
	Create(ctx context.Context, show dto.CreateMovieShowRequest) (string, error)
	Update(ctx context.Context, id string, show dto.UpdateMovieShowRequest) error
	Delete(ctx context.Context, id string) error
}

type TicketService interface {
	GetAll(ctx context.Context, filters dto.TicketFilters, page, limit int) (*dto.PaginatedTicketResponse, error)
	GetByID(ctx context.Context, id string) (*dto.TicketResponse, error)
	Create(ctx context.Context, ticket dto.CreateTicketRequest, movieShowID string) (string, error)
	UpdateStatus(ctx context.Context, id string, Status dto.UpdateStatusRequest) error
	Delete(ctx context.Context, id string) error
}

type ReviewService interface {
	GetAll(ctx context.Context, filters dto.ReviewFilters, page, limit int) (*dto.PaginatedReviewResponse, error)
	Create(ctx context.Context, review dto.CreateReviewRequest, movieID string) (string, error)
	Update(ctx context.Context, id string, review dto.UpdateReviewRequest) error
	Delete(ctx context.Context, id string) error
}

type HallService interface {
	GetAll(ctx context.Context, filters dto.HallFilters, page, limit int) (*dto.PaginatedHallResponse, error)
	GetByID(ctx context.Context, id string) (*dto.HallResponse, error)
	Create(ctx context.Context, hall dto.CreateHallRequest) (string, error)
	Update(ctx context.Context, id string, hall dto.UpdateHallRequest) error
	Delete(ctx context.Context, id string) error
}

type ScreenTypeService interface {
	GetAll(ctx context.Context, filters dto.ScreenTypeFilters, page, limit int) (*dto.PaginatedScreenTypeResponse, error)
	GetByID(ctx context.Context, id string) (*dto.ScreenTypeResponse, error)
	Create(ctx context.Context, screenType dto.CreateScreenTypeRequest) (string, error)
	Update(ctx context.Context, id string, screenType dto.UpdateScreenTypeRequest) error
	Delete(ctx context.Context, id string) error
}

type SeatTypeService interface {
	GetAll(ctx context.Context, filters dto.SeatTypeFilters, page, limit int) (*dto.PaginatedSeatTypeResponse, error)
	GetByID(ctx context.Context, id string) (*dto.SeatTypeResponse, error)
	Create(ctx context.Context, seatType dto.CreateSeatTypeRequest) (string, error)
	Update(ctx context.Context, id string, seatType dto.UpdateSeatTypeRequest) error
	Delete(ctx context.Context, id string) error
}

type SeatService interface {
	GetByHall(ctx context.Context, hallID string, filters dto.SeatFilters, page, limit int) (*dto.PaginatedSeatResponse, error)
	GetByID(ctx context.Context, id string) (*dto.SeatResponse, error)
	Create(ctx context.Context, seat dto.CreateSeatRequest) (string, error)
	Update(ctx context.Context, id string, seat dto.UpdateSeatRequest) error
	Delete(ctx context.Context, id string) error
}

type UserService interface {
	GetAll(ctx context.Context, filters dto.UserFilters, page, limit int) (*dto.PaginatedUserResponse, error)
	GetByID(ctx context.Context, id string) (*dto.UserResponse, error)
	Update(ctx context.Context, id string, user dto.UpdateUserRequest) error
	UpdateAdminStatus(ctx context.Context, id string, isAdmin bool) error
	Delete(ctx context.Context, id string) error
}
