package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	dto "cw/internal/dto/models"
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

func main() {
	apiClient := client.NewAPIClient("http://localhost:8080")
	app := NewTUIApp(apiClient)
	app.Run()
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

func (app *TUIApp) showAuthMenu() {
	fmt.Println("\n--- Authentication ---")
	fmt.Println("1. Login")
	fmt.Println("2. Register")
	fmt.Println("3. Exit")
	fmt.Print("Choose option: ")

	app.scanner.Scan()
	choice := app.scanner.Text()

	switch choice {
	case "1":
		app.login()
	case "2":
		app.register()
	case "3":
		fmt.Println("Goodbye!")
		os.Exit(0)
	default:
		fmt.Println("Invalid option!")
	}
}

func (app *TUIApp) showMainMenu() {
	app.displayMainMenuOptions()
	choice := app.getMenuChoice()
	app.processMenuChoice(choice)
}

func (app *TUIApp) displayMainMenuOptions() {
	fmt.Println("\n--- Main Menu ---")
	fmt.Println("1. Manage Genres")
	fmt.Println("2. Manage Movies")
	fmt.Println("3. Manage Movie Shows")
	fmt.Println("4. Manage Tickets")
	fmt.Println("5. Manage Reviews")
	fmt.Println("6. Manage Halls")
	fmt.Println("7. Manage Screen Types")
	fmt.Println("8. Manage Seat Types")
	fmt.Println("9. Manage Seats")
	if app.isAdmin {
		fmt.Println("10. Manage Users")
	}
	fmt.Println("0. Logout")
	fmt.Print("Choose option: ")
}

func (app *TUIApp) getMenuChoice() string {
	app.scanner.Scan()
	return app.scanner.Text()
}

func (app *TUIApp) processMenuChoice(choice string) {
	action := app.getActionForChoice(choice)
	if action != nil {
		action()
	} else {
		fmt.Println("Invalid option!")
	}
}

func (app *TUIApp) getActionForChoice(choice string) func() {
	actions := map[string]func(){
		"1":  app.manageGenres,
		"2":  app.manageMovies,
		"3":  app.manageMovieShows,
		"4":  app.manageTickets,
		"5":  app.manageReviews,
		"6":  app.manageHalls,
		"7":  app.manageScreenTypes,
		"8":  app.manageSeatTypes,
		"9":  app.manageSeats,
		"10": func() { app.handleUserManagementOption() },
		"0":  app.logout,
	}
	return actions[choice]
}

func (app *TUIApp) handleUserManagementOption() {
	if app.isAdmin {
		app.manageUsers()
	} else {
		fmt.Println("Invalid option!")
	}
}

func (app *TUIApp) login() {
	var email, password string

	fmt.Print("Email: ")
	app.scanner.Scan()
	email = app.scanner.Text()

	fmt.Print("PasswordHash: ")
	app.scanner.Scan()
	password = app.scanner.Text()

	token, err := app.authService.Login(context.Background(), email, password)
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return
	}

	app.token = token
	app.isAuthenticated = true
	app.isAdmin = true
	fmt.Println("Login successful!")
}

func (app *TUIApp) register() {
	var req dto.CreateUserRequest

	fmt.Print("Name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("Email: ")
	app.scanner.Scan()
	req.Email = app.scanner.Text()

	fmt.Print("PasswordHash: ")
	app.scanner.Scan()
	req.PasswordHash = app.scanner.Text()

	id, err := app.authService.Register(context.Background(), req)
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
		return
	}

	fmt.Printf("Registration successful! User ID: %s\n", id)
}

func (app *TUIApp) logout() {
	app.isAuthenticated = false
	app.isAdmin = false
	app.token = ""
	fmt.Println("Logged out successfully!")
}

func (app *TUIApp) manageGenres() {
	for {
		fmt.Println("\n--- Genre Management ---")
		fmt.Println("1. List genres")
		fmt.Println("2. Get genre by ID")
		fmt.Println("3. Create genre")
		fmt.Println("4. Update genre")
		fmt.Println("5. Delete genre")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listGenres()
		case "2":
			app.getGenre()
		case "3":
			app.createGenre()
		case "4":
			app.updateGenre()
		case "5":
			app.deleteGenre()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listGenres() {
	page, limit := app.getPaginationParams()

	var filters dto.GenreFilters
	fmt.Print("Name filter: ")
	app.scanner.Scan()
	filters.Name = app.scanner.Text()

	fmt.Print("Description filter: ")
	app.scanner.Scan()
	filters.Description = app.scanner.Text()

	result, err := app.genreService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nFound %d genres (total: %d):\n", len(result.Data), result.Total)
	for _, genre := range result.Data {
		fmt.Printf("ID: %s, Name: %s, Description: %s\n",
			genre.ID, genre.Name, genre.Description)
	}
}

func (app *TUIApp) getGenre() {
	fmt.Print("Enter genre ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	genre, err := app.genreService.GetByID(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("ID: %s\n", genre.ID)
	fmt.Printf("Name: %s\n", genre.Name)
	fmt.Printf("Description: %s\n", genre.Description)
}

func (app *TUIApp) createGenre() {
	var req dto.CreateGenreRequest

	fmt.Print("Genre name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("Description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	id, err := app.genreService.Create(context.Background(), req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Genre created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateGenre() {
	fmt.Print("Enter genre ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateGenreRequest
	fmt.Print("New name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("New description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	err := app.genreService.Update(context.Background(), id, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Genre updated successfully!")
}

func (app *TUIApp) deleteGenre() {
	fmt.Print("Enter genre ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.genreService.Delete(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Genre deleted successfully!")
}

func (app *TUIApp) getPaginationParams() (int, int) {
	fmt.Print("Page (default 1): ")
	app.scanner.Scan()
	pageStr := app.scanner.Text()
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			page = p
		}
	}

	fmt.Print("Limit (default 20): ")
	app.scanner.Scan()
	limitStr := app.scanner.Text()
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	return page, limit
}

func (app *TUIApp) manageMovies() {
	for {
		fmt.Println("\n--- Movie Management ---")
		fmt.Println("1. List movies")
		fmt.Println("2. Get movie by ID")
		fmt.Println("3. Create movie")
		fmt.Println("4. Update movie")
		fmt.Println("5. Delete movie")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listMovies()
		case "2":
			app.getMovie()
		case "3":
			app.createMovie()
		case "4":
			app.updateMovie()
		case "5":
			app.deleteMovie()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listMovies() {
	page, limit := app.getPaginationParams()

	var filters dto.MovieFilters
	fmt.Print("Title filter: ")
	app.scanner.Scan()
	filters.Title = app.scanner.Text()

	fmt.Print("Genre filter: ")
	app.scanner.Scan()
	filters.Genre = app.scanner.Text()

	result, err := app.movieService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nFound %d movies (total: %d):\n", len(result.Data), result.Total)
	for _, movie := range result.Data {
		fmt.Printf("ID: %s, Title: %s, Duration: %s min\n",
			movie.ID, movie.Title, movie.Duration)
	}
}

func (app *TUIApp) getMovie() {
	fmt.Print("Enter movie ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	movie, err := app.movieService.GetByID(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("ID: %s\n", movie.ID)
	fmt.Printf("Title: %s\n", movie.Title)
	fmt.Printf("Description: %s\n", movie.Description)
	fmt.Printf("Duration: %s min\n", movie.Duration)
}

func (app *TUIApp) createMovie() {
	var req dto.CreateMovieRequest

	fmt.Print("Title: ")
	app.scanner.Scan()
	req.Title = app.scanner.Text()

	fmt.Print("Description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	fmt.Print("Duration (minutes): ")
	app.scanner.Scan()
	durationStr := app.scanner.Text()
	req.Duration = durationStr

	fmt.Print("Genre IDs (comma separated): ")
	app.scanner.Scan()
	genreIDs := app.scanner.Text()
	if genreIDs != "" {
		req.GenreIDs = strings.Split(genreIDs, ",")
	}

	id, err := app.movieService.Create(context.Background(), req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Movie created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateMovie() {
	fmt.Print("Enter movie ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateMovieRequest
	fmt.Print("New title: ")
	app.scanner.Scan()
	req.Title = app.scanner.Text()

	fmt.Print("New description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	fmt.Print("New duration (minutes): ")
	app.scanner.Scan()
	durationStr := app.scanner.Text()
	req.Duration = durationStr

	err := app.movieService.Update(context.Background(), id, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Movie updated successfully!")
}

func (app *TUIApp) deleteMovie() {
	fmt.Print("Enter movie ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.movieService.Delete(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Movie deleted successfully!")
}

func (app *TUIApp) manageMovieShows() {
	for {
		fmt.Println("\n--- Movie Show Management ---")
		fmt.Println("1. List movie shows")
		fmt.Println("2. Get movie show by ID")
		fmt.Println("3. Create movie show")
		fmt.Println("4. Update movie show")
		fmt.Println("5. Delete movie show")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listMovieShows()
		case "2":
			app.getMovieShow()
		case "3":
			app.createMovieShow()
		case "4":
			app.updateMovieShow()
		case "5":
			app.deleteMovieShow()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listMovieShows() {
	page, limit := app.getPaginationParams()

	var filters dto.MovieShowFilters
	fmt.Print("Movie IDs (comma separated): ")
	app.scanner.Scan()
	movieIDs := app.scanner.Text()
	if movieIDs != "" {
		filters.MovieIDs = strings.Split(movieIDs, ",")
	}

	fmt.Print("Hall IDs (comma separated): ")
	app.scanner.Scan()
	hallIDs := app.scanner.Text()
	if hallIDs != "" {
		filters.HallIDs = strings.Split(hallIDs, ",")
	}

	fmt.Print("Date (YYYY-MM-DD): ")
	app.scanner.Scan()
	filters.Date = app.scanner.Text()

	result, err := app.movieShowService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nFound %d movie shows (total: %d):\n", len(result.Data), result.Total)
	for _, show := range result.Data {
		fmt.Printf("ID: %s, Movie: %s, Hall: %s, Start: %s\n",
			show.ID, show.MovieID, show.HallID, show.StartTime)
	}
}

func (app *TUIApp) getMovieShow() {
	fmt.Print("Enter movie show ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	show, err := app.movieShowService.GetByID(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("ID: %s\n", show.ID)
	fmt.Printf("Movie ID: %s\n", show.MovieID)
	fmt.Printf("Hall ID: %s\n", show.HallID)
	fmt.Printf("Start Time: %s\n", show.StartTime)
	fmt.Printf("Language: %s\n", show.Language)
}

func (app *TUIApp) createMovieShow() {
	var req dto.CreateMovieShowRequest

	fmt.Print("Movie ID: ")
	app.scanner.Scan()
	req.MovieID = app.scanner.Text()

	fmt.Print("Hall ID: ")
	app.scanner.Scan()
	req.HallID = app.scanner.Text()

	fmt.Print("Start Time (YYYY-MM-DD HH:MM:SS): ")
	app.scanner.Scan()
	startTimeStr := app.scanner.Text()

	startTime, err := time.Parse("2006-01-02 15:04:05", startTimeStr)
	if err != nil {
		fmt.Printf("Invalid time format: %v\n", err)
		return
	}
	req.StartTime = startTime

	fmt.Print("Language: ")
	app.scanner.Scan()
	req.Language = app.scanner.Text()

	fmt.Print("Price: ")
	app.scanner.Scan()
	priceStr := app.scanner.Text()
	if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
		req.BasePrice = price
	}

	id, err := app.movieShowService.Create(context.Background(), req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Movie show created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateMovieShow() {
	fmt.Print("Enter movie show ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateMovieShowRequest
	fmt.Print("New movie ID: ")
	app.scanner.Scan()
	req.MovieID = app.scanner.Text()

	fmt.Print("New hall ID: ")
	app.scanner.Scan()
	req.HallID = app.scanner.Text()

	fmt.Print("New start time (YYYY-MM-DD HH:MM:SS): ")
	app.scanner.Scan()
	startTimeStr := app.scanner.Text()

	startTime, err := time.Parse("2006-01-02 15:04:05", startTimeStr)
	if err != nil {
		fmt.Printf("Invalid time format: %v\n", err)
		return
	}
	req.StartTime = startTime

	fmt.Print("New language: ")
	app.scanner.Scan()
	req.Language = app.scanner.Text()

	fmt.Print("New price: ")
	app.scanner.Scan()
	priceStr := app.scanner.Text()
	if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
		req.BasePrice = price
	}

	err = app.movieShowService.Update(context.Background(), id, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Movie show updated successfully!")
}
func (app *TUIApp) deleteMovieShow() {
	fmt.Print("Enter movie show ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.movieShowService.Delete(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Movie show deleted successfully!")
}

func (app *TUIApp) manageTickets() {
	for {
		fmt.Println("\n--- Ticket Management ---")
		fmt.Println("1. List tickets")
		fmt.Println("2. Get ticket by ID")
		fmt.Println("3. Create ticket")
		fmt.Println("4. Update ticket Status")
		fmt.Println("5. Delete ticket")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listTickets()
		case "2":
			app.getTicket()
		case "3":
			app.createTicket()
		case "4":
			app.updateStatus()
		case "5":
			app.deleteTicket()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listTickets() {
	page, limit := app.getPaginationParams()

	var filters dto.TicketFilters
	fmt.Print("Movie Show ID: ")
	app.scanner.Scan()
	filters.MovieShowID = app.scanner.Text()

	fmt.Print("Statuses (comma separated): ")
	app.scanner.Scan()
	Statuses := app.scanner.Text()
	if Statuses != "" {
		filters.Status = strings.Split(Statuses, ",")
	}

	result, err := app.ticketService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nFound %d tickets (total: %d):\n", len(result.Data), result.Total)
	for _, ticket := range result.Data {
		fmt.Printf("ID: %s, Movie Show: %s, Seat: %s, Status: %s\n",
			ticket.ID, ticket.MovieShowID, ticket.SeatID, ticket.Status)
	}
}

func (app *TUIApp) getTicket() {
	fmt.Print("Enter ticket ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	ticket, err := app.ticketService.GetByID(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("ID: %s\n", ticket.ID)
	fmt.Printf("Movie Show ID: %s\n", ticket.MovieShowID)
	fmt.Printf("Seat ID: %s\n", ticket.SeatID)
	fmt.Printf("User ID: %s\n", ticket.UserID)
	fmt.Printf("Status: %s\n", ticket.Status)
	fmt.Printf("Price: %.2f\n", ticket.Price)
}

func (app *TUIApp) createTicket() {
	var req dto.CreateTicketRequest

	fmt.Print("Movie Show ID: ")
	app.scanner.Scan()
	movieShowID := app.scanner.Text()

	fmt.Print("Seat ID: ")
	app.scanner.Scan()
	req.SeatID = app.scanner.Text()

	fmt.Print("Price: ")
	app.scanner.Scan()
	priceStr := app.scanner.Text()
	if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
		req.Price = price
	}

	id, err := app.ticketService.Create(context.Background(), req, movieShowID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Ticket created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateStatus() {
	fmt.Print("Enter ticket ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateStatusRequest
	fmt.Print("New Status: ")
	app.scanner.Scan()
	req.Status = app.scanner.Text()

	err := app.ticketService.UpdateStatus(context.Background(), id, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Ticket Status updated successfully!")
}

func (app *TUIApp) deleteTicket() {
	fmt.Print("Enter ticket ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.ticketService.Delete(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Ticket deleted successfully!")
}

func (app *TUIApp) manageReviews() {
	for {
		fmt.Println("\n--- Review Management ---")
		fmt.Println("1. List reviews")
		fmt.Println("2. Create review")
		fmt.Println("3. Update review")
		fmt.Println("4. Delete review")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listReviews()
		case "2":
			app.createReview()
		case "3":
			app.updateReview()
		case "4":
			app.deleteReview()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listReviews() {
	page, limit := app.getPaginationParams()

	var filters dto.ReviewFilters
	fmt.Print("Movie ID: ")
	app.scanner.Scan()
	filters.MovieID = app.scanner.Text()

	fmt.Print("User ID: ")
	app.scanner.Scan()
	filters.UserID = app.scanner.Text()

	result, err := app.reviewService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nFound %d reviews (total: %d):\n", len(result.Data), result.Total)
	for _, review := range result.Data {
		fmt.Printf("ID: %s, Movie: %s, User: %s, Rating: %d\n",
			review.ID, review.MovieID, review.UserID, review.Rating)
	}
}

func (app *TUIApp) createReview() {
	var req dto.CreateReviewRequest

	fmt.Print("Movie ID: ")
	app.scanner.Scan()
	movieID := app.scanner.Text()

	fmt.Print("Rating (1-10): ")
	app.scanner.Scan()
	ratingStr := app.scanner.Text()
	if rating, err := strconv.Atoi(ratingStr); err == nil {
		req.Rating = rating
	}

	fmt.Print("Comment: ")
	app.scanner.Scan()
	req.Comment = app.scanner.Text()

	id, err := app.reviewService.Create(context.Background(), req, movieID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Review created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateReview() {
	fmt.Print("Enter review ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateReviewRequest
	fmt.Print("New rating (1-10): ")
	app.scanner.Scan()
	ratingStr := app.scanner.Text()
	if rating, err := strconv.Atoi(ratingStr); err == nil {
		req.Rating = rating
	}

	fmt.Print("New comment: ")
	app.scanner.Scan()
	req.Comment = app.scanner.Text()

	err := app.reviewService.Update(context.Background(), id, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Review updated successfully!")
}

func (app *TUIApp) deleteReview() {
	fmt.Print("Enter review ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.reviewService.Delete(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Review deleted successfully!")
}

func (app *TUIApp) manageHalls() {
	for {
		fmt.Println("\n--- Hall Management ---")
		fmt.Println("1. List halls")
		fmt.Println("2. Get hall by ID")
		fmt.Println("3. Create hall")
		fmt.Println("4. Update hall")
		fmt.Println("5. Delete hall")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listHalls()
		case "2":
			app.getHall()
		case "3":
			app.createHall()
		case "4":
			app.updateHall()
		case "5":
			app.deleteHall()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listHalls() {
	page, limit := app.getPaginationParams()

	var filters dto.HallFilters
	fmt.Print("Name filter: ")
	app.scanner.Scan()
	filters.Name = app.scanner.Text()

	fmt.Print("Screen Type ID: ")
	app.scanner.Scan()
	filters.ScreenTypeID = app.scanner.Text()

	result, err := app.hallService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nFound %d halls (total: %d):\n", len(result.Data), result.Total)
	for _, hall := range result.Data {
		fmt.Printf("ID: %s, Name: %s, Screen Type: %s\n",
			hall.ID, hall.Name, hall.ScreenTypeID)
	}
}

func (app *TUIApp) getHall() {
	fmt.Print("Enter hall ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	hall, err := app.hallService.GetByID(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("ID: %s\n", hall.ID)
	fmt.Printf("Name: %s\n", hall.Name)
	fmt.Printf("Screen Type ID: %s\n", hall.ScreenTypeID)
	fmt.Printf("Description: %s\n", hall.Description)
}

func (app *TUIApp) createHall() {
	var req dto.CreateHallRequest

	fmt.Print("Name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("Screen Type ID: ")
	app.scanner.Scan()
	req.ScreenTypeID = app.scanner.Text()

	fmt.Print("Description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	id, err := app.hallService.Create(context.Background(), req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Hall created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateHall() {
	fmt.Print("Enter hall ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateHallRequest
	fmt.Print("New name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("New screen type ID: ")
	app.scanner.Scan()
	req.ScreenTypeID = app.scanner.Text()

	fmt.Print("New description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	err := app.hallService.Update(context.Background(), id, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Hall updated successfully!")
}

func (app *TUIApp) deleteHall() {
	fmt.Print("Enter hall ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.hallService.Delete(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Hall deleted successfully!")
}

func (app *TUIApp) manageScreenTypes() {
	for {
		fmt.Println("\n--- Screen Type Management ---")
		fmt.Println("1. List screen types")
		fmt.Println("2. Get screen type by ID")
		fmt.Println("3. Create screen type")
		fmt.Println("4. Update screen type")
		fmt.Println("5. Delete screen type")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listScreenTypes()
		case "2":
			app.getScreenType()
		case "3":
			app.createScreenType()
		case "4":
			app.updateScreenType()
		case "5":
			app.deleteScreenType()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listScreenTypes() {
	page, limit := app.getPaginationParams()

	var filters dto.ScreenTypeFilters
	fmt.Print("Name filter: ")
	app.scanner.Scan()
	filters.Name = app.scanner.Text()

	result, err := app.screenTypeService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nFound %d screen types (total: %d):\n", len(result.Data), result.Total)
	for _, st := range result.Data {
		fmt.Printf("ID: %s, Name: %s\n", st.ID, st.Name)
	}
}

func (app *TUIApp) getScreenType() {
	fmt.Print("Enter screen type ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	st, err := app.screenTypeService.GetByID(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("ID: %s\n", st.ID)
	fmt.Printf("Name: %s\n", st.Name)
	fmt.Printf("Description: %s\n", st.Description)
}

func (app *TUIApp) createScreenType() {
	var req dto.CreateScreenTypeRequest

	fmt.Print("Name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("Description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	id, err := app.screenTypeService.Create(context.Background(), req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Screen type created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateScreenType() {
	fmt.Print("Enter screen type ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateScreenTypeRequest
	fmt.Print("New name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("New description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	err := app.screenTypeService.Update(context.Background(), id, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Screen type updated successfully!")
}

func (app *TUIApp) deleteScreenType() {
	fmt.Print("Enter screen type ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.screenTypeService.Delete(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Screen type deleted successfully!")
}

func (app *TUIApp) manageSeatTypes() {
	for {
		fmt.Println("\n--- Seat Type Management ---")
		fmt.Println("1. List seat types")
		fmt.Println("2. Get seat type by ID")
		fmt.Println("3. Create seat type")
		fmt.Println("4. Update seat type")
		fmt.Println("5. Delete seat type")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listSeatTypes()
		case "2":
			app.getSeatType()
		case "3":
			app.createSeatType()
		case "4":
			app.updateSeatType()
		case "5":
			app.deleteSeatType()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listSeatTypes() {
	page, limit := app.getPaginationParams()

	var filters dto.SeatTypeFilters
	fmt.Print("Name filter: ")
	app.scanner.Scan()
	filters.Name = app.scanner.Text()

	result, err := app.seatTypeService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nFound %d seat types (total: %d):\n", len(result.Data), result.Total)
	for _, st := range result.Data {
		fmt.Printf("ID: %s, Name: %s, Price Multiplier: %.2f\n",
			st.ID, st.Name, st.PriceModifier)
	}
}

func (app *TUIApp) getSeatType() {
	fmt.Print("Enter seat type ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	st, err := app.seatTypeService.GetByID(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("ID: %s\n", st.ID)
	fmt.Printf("Name: %s\n", st.Name)
	fmt.Printf("Price Multiplier: %.2f\n", st.PriceModifier)
	fmt.Printf("Description: %s\n", st.Description)
}

func (app *TUIApp) createSeatType() {
	var req dto.CreateSeatTypeRequest

	fmt.Print("Name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("Price Multiplier: ")
	app.scanner.Scan()
	multiplierStr := app.scanner.Text()
	if multiplier, err := strconv.ParseFloat(multiplierStr, 64); err == nil {
		req.PriceModifier = multiplier
	}

	fmt.Print("Description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	id, err := app.seatTypeService.Create(context.Background(), req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Seat type created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateSeatType() {
	fmt.Print("Enter seat type ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateSeatTypeRequest
	fmt.Print("New name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("New price multiplier: ")
	app.scanner.Scan()
	multiplierStr := app.scanner.Text()
	if multiplier, err := strconv.ParseFloat(multiplierStr, 64); err == nil {
		req.PriceModifier = multiplier
	}

	fmt.Print("New description: ")
	app.scanner.Scan()
	req.Description = app.scanner.Text()

	err := app.seatTypeService.Update(context.Background(), id, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Seat type updated successfully!")
}

func (app *TUIApp) deleteSeatType() {
	fmt.Print("Enter seat type ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.seatTypeService.Delete(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Seat type deleted successfully!")
}

func (app *TUIApp) manageSeats() {
	for {
		fmt.Println("\n--- Seat Management ---")
		fmt.Println("1. List seats by hall")
		fmt.Println("2. Get seat by ID")
		fmt.Println("3. Create seat")
		fmt.Println("4. Update seat")
		fmt.Println("5. Delete seat")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listSeatsByHall()
		case "2":
			app.getSeat()
		case "3":
			app.createSeat()
		case "4":
			app.updateSeat()
		case "5":
			app.deleteSeat()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listSeatsByHall() {
	fmt.Print("Enter hall ID: ")
	app.scanner.Scan()
	hallID := app.scanner.Text()

	page, limit := app.getPaginationParams()

	var filters dto.SeatFilters
	fmt.Print("Seat Type ID: ")
	app.scanner.Scan()
	filters.SeatTypeID = app.scanner.Text()

	result, err := app.seatService.GetByHall(context.Background(), hallID, filters, page, limit)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nFound %d seats (total: %d):\n", len(result.Data), result.Total)
	for _, seat := range result.Data {
		fmt.Printf("ID: %s, Row: %d, Number: %d, Type: %s\n",
			seat.ID, seat.RowNumber, seat.SeatNumber, seat.SeatTypeID)
	}
}

func (app *TUIApp) getSeat() {
	fmt.Print("Enter seat ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	seat, err := app.seatService.GetByID(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("ID: %s\n", seat.ID)
	fmt.Printf("Hall ID: %s\n", seat.HallID)
	fmt.Printf("Seat Type ID: %s\n", seat.SeatTypeID)
	fmt.Printf("Row: %d\n", seat.RowNumber)
	fmt.Printf("Number: %d\n", seat.SeatNumber)
}

func (app *TUIApp) createSeat() {
	var req dto.CreateSeatRequest

	fmt.Print("Hall ID: ")
	app.scanner.Scan()
	req.HallID = app.scanner.Text()

	fmt.Print("Seat Type ID: ")
	app.scanner.Scan()
	req.SeatTypeID = app.scanner.Text()

	fmt.Print("Row Number: ")
	app.scanner.Scan()
	rowStr := app.scanner.Text()
	if row, err := strconv.Atoi(rowStr); err == nil {
		req.RowNumber = row
	}

	fmt.Print("Seat Number: ")
	app.scanner.Scan()
	seatStr := app.scanner.Text()
	if seat, err := strconv.Atoi(seatStr); err == nil {
		req.SeatNumber = seat
	}

	id, err := app.seatService.Create(context.Background(), req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Seat created successfully! ID: %s\n", id)
}

func (app *TUIApp) updateSeat() {
	fmt.Print("Enter seat ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateSeatRequest
	fmt.Print("New seat type ID: ")
	app.scanner.Scan()
	req.SeatTypeID = app.scanner.Text()

	fmt.Print("New row number: ")
	app.scanner.Scan()
	rowStr := app.scanner.Text()
	if row, err := strconv.Atoi(rowStr); err == nil {
		req.RowNumber = row
	}

	fmt.Print("New seat number: ")
	app.scanner.Scan()
	seatStr := app.scanner.Text()
	if seat, err := strconv.Atoi(seatStr); err == nil {
		req.SeatNumber = seat
	}

	err := app.seatService.Update(context.Background(), id, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Seat updated successfully!")
}

func (app *TUIApp) deleteSeat() {
	fmt.Print("Enter seat ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.seatService.Delete(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Seat deleted successfully!")
}

func (app *TUIApp) manageUsers() {
	for {
		fmt.Println("\n--- User Management ---")
		fmt.Println("1. List users")
		fmt.Println("2. Get user by ID")
		fmt.Println("3. Update user")
		fmt.Println("4. Update admin Status")
		fmt.Println("5. Delete user")
		fmt.Println("0. Back")
		fmt.Print("Choose option: ")

		app.scanner.Scan()
		choice := app.scanner.Text()

		switch choice {
		case "1":
			app.listUsers()
		case "2":
			app.getUser()
		case "3":
			app.updateUser()
		case "4":
			app.updateAdminStatus()
		case "5":
			app.deleteUser()
		case "0":
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

func (app *TUIApp) listUsers() {
	page, limit := app.getPaginationParams()

	var filters dto.UserFilters
	fmt.Print("Name filter: ")
	app.scanner.Scan()
	filters.Name = app.scanner.Text()

	fmt.Print("Email filter: ")
	app.scanner.Scan()
	filters.Email = app.scanner.Text()

	result, err := app.userService.GetAll(context.Background(), filters, page, limit)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nFound %d users (total: %d):\n", len(result.Data), result.Total)
	for _, user := range result.Data {
		fmt.Printf("ID: %s, Name: %s, Email: %s, Admin: %t\n",
			user.ID, user.Name, user.Email, user.IsAdmin)
	}
}

func (app *TUIApp) getUser() {
	fmt.Print("Enter user ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	user, err := app.userService.GetByID(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("ID: %s\n", user.ID)
	fmt.Printf("Name: %s\n", user.Name)
	fmt.Printf("Email: %s\n", user.Email)
	fmt.Printf("Is Admin: %t\n", user.IsAdmin)
}

func (app *TUIApp) updateUser() {
	fmt.Print("Enter user ID to update: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	var req dto.UpdateUserRequest
	fmt.Print("New name: ")
	app.scanner.Scan()
	req.Name = app.scanner.Text()

	fmt.Print("New email: ")
	app.scanner.Scan()
	req.Email = app.scanner.Text()

	err := app.userService.Update(context.Background(), id, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("User updated successfully!")
}

func (app *TUIApp) updateAdminStatus() {
	fmt.Print("Enter user ID: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	fmt.Print("Is admin (true/false): ")
	app.scanner.Scan()
	isAdminStr := app.scanner.Text()
	isAdmin := strings.ToLower(isAdminStr) == "true"

	err := app.userService.UpdateAdminStatus(context.Background(), id, isAdmin)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Admin Status updated successfully!")
}

func (app *TUIApp) deleteUser() {
	fmt.Print("Enter user ID to delete: ")
	app.scanner.Scan()
	id := app.scanner.Text()

	err := app.userService.Delete(context.Background(), id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("User deleted successfully!")
}
