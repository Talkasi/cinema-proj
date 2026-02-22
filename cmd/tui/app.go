package main

import (
	"bufio"
	"fmt"
	"os"

	"cw/internal/tui/client"
	"cw/internal/tui/services"
)

type TUIApp struct {
	authService       services.AuthService
	genreService      services.GenreService
	movieService      services.MovieService
	movieShowService  services.MovieShowService
	ticketService     services.TicketService
	reviewService     services.ReviewService
	hallService       services.HallService
	screenTypeService services.ScreenTypeService
	seatTypeService   services.SeatTypeService
	seatService       services.SeatService
	userService       services.UserService
	scanner           *bufio.Scanner
	isAuthenticated   bool
	isAdmin           bool
	token             string
}

func NewTUIApp(apiClient *client.APIClient) *TUIApp {
	return &TUIApp{
		authService:       services.NewAuthService(apiClient),
		genreService:      services.NewGenreService(apiClient),
		movieService:      services.NewMovieService(apiClient),
		movieShowService:  services.NewMovieShowService(apiClient),
		ticketService:     services.NewTicketService(apiClient),
		reviewService:     services.NewReviewService(apiClient),
		hallService:       services.NewHallService(apiClient),
		screenTypeService: services.NewScreenTypeService(apiClient),
		seatTypeService:   services.NewSeatTypeService(apiClient),
		seatService:       services.NewSeatService(apiClient),
		userService:       services.NewUserService(apiClient),
		scanner:           bufio.NewScanner(os.Stdin),
	}
}

func (app *TUIApp) Run() {
	fmt.Println("=== Cinema Management System TUI ===")

	for {
		if !app.isAuthenticated {
			app.showAuthMenu()
		} else {
			app.showMainMenu()
		}
	}
}
