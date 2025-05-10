package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter() *http.ServeMux {
	mux := new(http.ServeMux)

	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	mux.HandleFunc("GET /equipment-types", Midleware(RoleBasedHandler(GetEquipmentTypes)))
	mux.HandleFunc("GET /equipment-types/{id}", Midleware(RoleBasedHandler(GetEquipmentTypeByID)))
	mux.HandleFunc("POST /equipment-types", Midleware(RoleBasedHandler(CreateEquipmentType)))
	mux.HandleFunc("PUT /equipment-types/{id}", Midleware(RoleBasedHandler(UpdateEquipmentType)))
	mux.HandleFunc("DELETE /equipment-types/{id}", Midleware(RoleBasedHandler(DeleteEquipmentType)))

	mux.HandleFunc("GET /genres", Midleware(RoleBasedHandler(GetGenres)))
	mux.HandleFunc("GET /genres/{id}", Midleware(RoleBasedHandler(GetGenreByID)))
	mux.HandleFunc("POST /genres", Midleware(RoleBasedHandler(CreateGenre)))
	mux.HandleFunc("PUT /genres/{id}", Midleware(RoleBasedHandler(UpdateGenre)))
	mux.HandleFunc("DELETE /genres/{id}", Midleware(RoleBasedHandler(DeleteGenre)))

	mux.HandleFunc("GET /halls", Midleware(RoleBasedHandler(GetHalls)))
	mux.HandleFunc("GET /halls/{id}", Midleware(RoleBasedHandler(GetHallByID)))
	mux.HandleFunc("POST /halls", Midleware(RoleBasedHandler(CreateHall)))
	mux.HandleFunc("PUT /halls/{id}", Midleware(RoleBasedHandler(UpdateHall)))
	mux.HandleFunc("DELETE /halls/{id}", Midleware(RoleBasedHandler(DeleteHall)))

	mux.HandleFunc("GET /movies", Midleware(RoleBasedHandler(GetMovies)))
	mux.HandleFunc("GET /movies/{id}", Midleware(RoleBasedHandler(GetMovieByID)))
	mux.HandleFunc("POST /movies", Midleware(RoleBasedHandler(CreateMovie)))
	mux.HandleFunc("PUT /movies/{id}", Midleware(RoleBasedHandler(UpdateMovie)))
	mux.HandleFunc("DELETE /movies/{id}", Midleware(RoleBasedHandler(DeleteMovie)))

	mux.HandleFunc("GET /movie-shows", Midleware(RoleBasedHandler(GetMovieShows)))
	mux.HandleFunc("GET /movie-shows/{id}", Midleware(RoleBasedHandler(GetMovieShowByID)))
	mux.HandleFunc("POST /movie-shows", Midleware(RoleBasedHandler(CreateMovieShow)))
	mux.HandleFunc("PUT /movie-shows/{id}", Midleware(RoleBasedHandler(UpdateMovieShow)))
	mux.HandleFunc("DELETE /movie-shows/{id}", Midleware(RoleBasedHandler(DeleteMovieShow)))

	mux.HandleFunc("GET /reviews", Midleware(RoleBasedHandler(GetReviews)))
	mux.HandleFunc("GET /reviews/{id}", Midleware(RoleBasedHandler(GetReviewByID)))
	mux.HandleFunc("POST /reviews", Midleware(RoleBasedHandler(CreateReview)))
	mux.HandleFunc("PUT /reviews/{id}", Midleware(RoleBasedHandler(UpdateReview)))
	mux.HandleFunc("DELETE /reviews/{id}", Midleware(RoleBasedHandler(DeleteReview)))

	mux.HandleFunc("GET /seats", Midleware(RoleBasedHandler(GetSeats)))
	mux.HandleFunc("GET /seats/{id}", Midleware(RoleBasedHandler(GetSeatByID)))
	mux.HandleFunc("POST /seats", Midleware(RoleBasedHandler(CreateSeat)))
	mux.HandleFunc("PUT /seats/{id}", Midleware(RoleBasedHandler(UpdateSeat)))
	mux.HandleFunc("DELETE /seats/{id}", Midleware(RoleBasedHandler(DeleteSeat)))

	mux.HandleFunc("GET /seat-types", Midleware(RoleBasedHandler(GetSeatTypes)))
	mux.HandleFunc("GET /seat-types/{id}", Midleware(RoleBasedHandler(GetSeatTypeByID)))
	mux.HandleFunc("POST /seat-types", Midleware(RoleBasedHandler(CreateSeatType)))
	mux.HandleFunc("PUT /seat-types/{id}", Midleware(RoleBasedHandler(UpdateSeatType)))
	mux.HandleFunc("DELETE /seat-types/{id}", Midleware(RoleBasedHandler(DeleteSeatType)))

	mux.HandleFunc("GET /tickets/movie-show/{movie_show_id}", Midleware(RoleBasedHandler(GetTicketsByMovieShowID)))
	mux.HandleFunc("GET /tickets/{id}", Midleware(RoleBasedHandler(GetTicketByID)))
	mux.HandleFunc("POST /tickets", Midleware(RoleBasedHandler(CreateTicket)))
	mux.HandleFunc("PUT /tickets/{id}", Midleware(RoleBasedHandler(UpdateTicket)))
	mux.HandleFunc("DELETE /tickets/{id}", Midleware(RoleBasedHandler(DeleteTicket)))

	mux.HandleFunc("GET /ticket-status", Midleware(RoleBasedHandler(GetTicketStatuses)))
	mux.HandleFunc("GET /ticket-status/{id}", Midleware(RoleBasedHandler(GetTicketStatusByID)))
	mux.HandleFunc("POST /ticket-status", Midleware(RoleBasedHandler(CreateTicketStatus)))
	mux.HandleFunc("PUT /ticket-status/{id}", Midleware(RoleBasedHandler(UpdateTicketStatus)))
	mux.HandleFunc("DELETE /ticket-status/{id}", Midleware(RoleBasedHandler(DeleteTicketStatus)))

	mux.HandleFunc("GET /users", Midleware(RoleBasedHandler(GetUsers)))
	mux.HandleFunc("GET /users/{id}", Midleware(RoleBasedHandler(GetUserByID)))
	mux.HandleFunc("POST /users", Midleware(RoleBasedHandler(CreateUser)))
	mux.HandleFunc("PUT /users/{id}", Midleware(RoleBasedHandler(UpdateUser)))
	mux.HandleFunc("DELETE /users/{id}", Midleware(RoleBasedHandler(DeleteUser)))

	mux.HandleFunc("/register", Midleware(RoleBasedHandler(RegisterUser)))
	mux.HandleFunc("/login", Midleware(RoleBasedHandler(LoginUser)))

	return mux
}

func CreateAll(db *pgxpool.Pool) error {
	schemaSQL, err := os.ReadFile("./schemas/create.sql")
	if err != nil {
		return fmt.Errorf("ошибка чтения SQL файла: %v", err)
	}

	_, err = db.Exec(context.Background(), string(schemaSQL))
	if err != nil {
		return fmt.Errorf("ошибка выполнения SQL схемы: %v", err)
	}

	return nil
}

func DeleteAll(db *pgxpool.Pool) error {
	schemaSQL, err := os.ReadFile("./schemas/delete.sql")
	if err != nil {
		return fmt.Errorf("ошибка чтения SQL файла: %v", err)
	}

	_, err = db.Exec(context.Background(), string(schemaSQL))
	if err != nil {
		return fmt.Errorf("ошибка выполнения SQL схемы: %v", err)
	}

	return nil
}

func SeedAll(db *pgxpool.Pool) error {
	if err := SeedGenres(db); err != nil {
		return fmt.Errorf("ошибка при вставке жанров: %v", err)
	}

	if err := SeedEquipmentTypes(db); err != nil {
		return fmt.Errorf("ошибка при вставке типов оборудования: %v", err)
	}

	if err := SeedSeatTypes(db); err != nil {
		return fmt.Errorf("ошибка при вставке типов мест: %v", err)
	}

	if err := SeedTicketStatuses(db); err != nil {
		return fmt.Errorf("ошибка при вставке статусов билетов: %v", err)
	}

	return nil
}

func SeedGenres(db *pgxpool.Pool) error {
	for _, g := range GenresData {
		_, err := db.Exec(context.Background(), `INSERT INTO genres (name, description)
			VALUES ($1, $2)
			ON CONFLICT (name) DO NOTHING`, g.Name, g.Description)
		if err != nil {
			return err
		}
	}
	return nil
}

func SeedEquipmentTypes(db *pgxpool.Pool) error {
	for _, e := range EquipmentTypesData {
		_, err := db.Exec(context.Background(), `INSERT INTO equipment_types (name, description)
			VALUES ($1, $2)
			ON CONFLICT (name) DO NOTHING`, e.Name, e.Description)
		if err != nil {
			return err
		}
	}
	return nil
}

func SeedSeatTypes(db *pgxpool.Pool) error {
	for _, s := range SeatTypesData {
		_, err := db.Exec(context.Background(), `INSERT INTO seat_types (name, description)
			VALUES ($1, $2)
			ON CONFLICT (name) DO NOTHING`, s.Name, s.Description)
		if err != nil {
			return err
		}
	}
	return nil
}

func SeedTicketStatuses(db *pgxpool.Pool) error {
	for _, status := range TicketStatusesData {
		_, err := db.Exec(context.Background(), `INSERT INTO ticket_status (name)
			VALUES ($1)
			ON CONFLICT (name) DO NOTHING`, status.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
