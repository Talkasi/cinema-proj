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

func SeedGenres(db *pgxpool.Pool) error {
	for _, g := range GenresData {
		_, err := db.Exec(context.Background(), `INSERT INTO genres (id, name, description)
            VALUES ($1, $2, $3)
            ON CONFLICT (name) DO UPDATE SET 
                description = EXCLUDED.description,
                id = EXCLUDED.id`,
			g.ID, g.Name, g.Description)
		if err != nil {
			return err
		}
	}
	return nil
}

func SeedHalls(db *pgxpool.Pool) error {
	for _, h := range HallsData {
		_, err := db.Exec(context.Background(), `INSERT INTO halls (id, name, capacity, equipment_type_id, description)
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (id) DO UPDATE SET
                capacity = EXCLUDED.capacity,
                equipment_type_id = EXCLUDED.equipment_type_id,
                description = EXCLUDED.description,
                name = EXCLUDED.name`,
			h.ID, h.Name, h.Capacity, h.EquipmentTypeID, h.Description)
		if err != nil {
			return err
		}
	}
	return nil
}

func SeedEquipmentTypes(db *pgxpool.Pool) error {
	for _, e := range EquipmentTypesData {
		_, err := db.Exec(context.Background(), `INSERT INTO equipment_types (id, name, description)
            VALUES ($1, $2, $3)
            ON CONFLICT (name) DO UPDATE SET
                description = EXCLUDED.description,
                id = EXCLUDED.id`,
			e.ID, e.Name, e.Description)
		if err != nil {
			return err
		}
	}
	return nil
}

func SeedSeatTypes(db *pgxpool.Pool) error {
	for _, s := range SeatTypesData {
		_, err := db.Exec(context.Background(), `INSERT INTO seat_types (id, name, description)
            VALUES ($1, $2, $3)
            ON CONFLICT (name) DO UPDATE SET
                description = EXCLUDED.description,
                id = EXCLUDED.id`,
			s.ID, s.Name, s.Description)
		if err != nil {
			return err
		}
	}
	return nil
}

func SeedMovies(db *pgxpool.Pool) error {
	for _, m := range MoviesData {
		_, err := db.Exec(context.Background(), `INSERT INTO movies 
            (id, title, duration, rating, description, age_limit, box_office_revenue, release_date)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            ON CONFLICT (id) DO UPDATE SET
                title = EXCLUDED.title,
                duration = EXCLUDED.duration,
                rating = EXCLUDED.rating,
                description = EXCLUDED.description,
                age_limit = EXCLUDED.age_limit,
                box_office_revenue = EXCLUDED.box_office_revenue,
                release_date = EXCLUDED.release_date`,
			m.ID, m.Title, m.Duration, m.Rating, m.Description, m.AgeLimit, m.BoxOfficeRevenue, m.ReleaseDate)
		if err != nil {
			return fmt.Errorf("ошибка при вставке фильма %s: %v", m.Title, err)
		}
	}
	return nil
}

func SeedMoviesGenres(db *pgxpool.Pool) error {
	for _, mg := range MoviesGenresData {
		_, err := db.Exec(context.Background(), `INSERT INTO movies_genres 
            (movie_id, genre_id) VALUES ($1, $2)
            ON CONFLICT (movie_id, genre_id) DO NOTHING`,
			mg[0], mg[1])
		if err != nil {
			return fmt.Errorf("ошибка при связывании фильма и жанра: %v", err)
		}
	}
	return nil
}

func SeedUsers(db *pgxpool.Pool) error {
	for _, u := range UsersData {
		_, err := db.Exec(context.Background(), `INSERT INTO users 
            (id, name, email, password_hash, birth_date, is_blocked, is_admin)
            VALUES ($1, $2, $3, $4, $5, $6, $7)
            ON CONFLICT (email) DO UPDATE SET
                name = EXCLUDED.name,
                password_hash = EXCLUDED.password_hash,
                birth_date = EXCLUDED.birth_date,
                is_blocked = EXCLUDED.is_blocked,
                is_admin = EXCLUDED.is_admin`,
			u.ID, u.Name, u.Email, u.PasswordHash, u.BirthDate, u.IsBlocked, u.IsAdmin)
		if err != nil {
			return fmt.Errorf("ошибка при вставке пользователя %s: %v", u.Email, err)
		}
	}
	return nil
}

func SeedMovieShows(db *pgxpool.Pool) error {
	for _, ms := range MovieShowsData {
		_, err := db.Exec(context.Background(), `INSERT INTO movie_shows 
            (id, movie_id, hall_id, start_time, language)
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (id) DO UPDATE SET
                movie_id = EXCLUDED.movie_id,
                hall_id = EXCLUDED.hall_id,
                start_time = EXCLUDED.start_time,
                language = EXCLUDED.language`,
			ms.ID, ms.MovieID, ms.HallID, ms.StartTime, ms.Language)
		if err != nil {
			return fmt.Errorf("ошибка при вставке киносеанса: %v", err)
		}
	}
	return nil
}

func SeedSeats(db *pgxpool.Pool) error {
	for _, s := range SeatsData {
		_, err := db.Exec(context.Background(), `INSERT INTO seats 
            (id, hall_id, seat_type_id, row_number, seat_number)
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (id) DO UPDATE SET
                hall_id = EXCLUDED.hall_id,
                seat_type_id = EXCLUDED.seat_type_id,
                row_number = EXCLUDED.row_number,
                seat_number = EXCLUDED.seat_number`,
			s.ID, s.HallID, s.SeatTypeID, s.RowNumber, s.SeatNumber)
		if err != nil {
			return fmt.Errorf("ошибка при вставке места: %v", err)
		}
	}
	return nil
}

func SeedTickets(db *pgxpool.Pool) error {
	for _, t := range TicketsData {
		_, err := db.Exec(context.Background(), `INSERT INTO tickets 
            (id, movie_show_id, seat_id, ticket_status, price)
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (id) DO UPDATE SET
                movie_show_id = EXCLUDED.movie_show_id,
                seat_id = EXCLUDED.seat_id,
                ticket_status = EXCLUDED.ticket_status,
                price = EXCLUDED.price`,
			t.ID, t.MovieShowID, t.SeatID, t.Status, t.Price)
		if err != nil {
			return fmt.Errorf("ошибка при вставке билета: %v", err)
		}
	}
	return nil
}

func SeedReviews(db *pgxpool.Pool) error {
	for _, r := range ReviewsData {
		_, err := db.Exec(context.Background(), `INSERT INTO reviews 
            (id, user_id, movie_id, rating, review_comment)
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (id) DO UPDATE SET
                rating = EXCLUDED.rating,
                review_comment = EXCLUDED.review_comment`,
			r.ID, r.UserID, r.MovieID, r.Rating, r.Comment)
		if err != nil {
			return fmt.Errorf("ошибка при вставке отзыва: %v", err)
		}
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

	if err := SeedHalls(db); err != nil {
		return fmt.Errorf("ошибка при вставке залов: %v", err)
	}

	if err := SeedMovies(db); err != nil {
		return fmt.Errorf("ошибка при вставке фильмов: %v", err)
	}

	if err := SeedMoviesGenres(db); err != nil {
		return fmt.Errorf("ошибка при связывании фильмов и жанров: %v", err)
	}

	if err := SeedUsers(db); err != nil {
		return fmt.Errorf("ошибка при вставке пользователей: %v", err)
	}

	if err := SeedMovieShows(db); err != nil {
		return fmt.Errorf("ошибка при вставке киносеансов: %v", err)
	}

	if err := SeedSeats(db); err != nil {
		return fmt.Errorf("ошибка при вставке мест: %v", err)
	}

	if err := SeedTickets(db); err != nil {
		return fmt.Errorf("ошибка при вставке билетов: %v", err)
	}

	if err := SeedReviews(db); err != nil {
		return fmt.Errorf("ошибка при вставке отзывов: %v", err)
	}

	return nil
}

func ClearTable(db *pgxpool.Pool, tableName string) error {
	query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tableName)

	_, err := db.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to clear table %s: %w", tableName, err)
	}
	return nil
}

func ClearAll(db *pgxpool.Pool) error {
	if err := ClearTable(db, "genres"); err != nil {
		return fmt.Errorf("ошибка при очищении жанров: %v", err)
	}

	if err := ClearTable(db, "equipment_types"); err != nil {
		return fmt.Errorf("ошибка при очищении типов оборудования: %v", err)
	}

	if err := ClearTable(db, "seat_types"); err != nil {
		return fmt.Errorf("ошибка при очищении типов мест: %v", err)
	}

	if err := ClearTable(db, "halls"); err != nil {
		return fmt.Errorf("ошибка при очищении залов: %v", err)
	}

	if err := ClearTable(db, "movies"); err != nil {
		return fmt.Errorf("ошибка при очищении фильмов: %v", err)
	}

	if err := ClearTable(db, "movies_genres"); err != nil {
		return fmt.Errorf("ошибка при очищении таблицы связи фильмов и жанров: %v", err)
	}

	if err := ClearTable(db, "users"); err != nil {
		return fmt.Errorf("ошибка при очищении пользователей: %v", err)
	}

	if err := ClearTable(db, "movie_shows"); err != nil {
		return fmt.Errorf("ошибка при очищении киносеансов: %v", err)
	}

	if err := ClearTable(db, "seats"); err != nil {
		return fmt.Errorf("ошибка при очищении мест: %v", err)
	}

	if err := ClearTable(db, "tickets"); err != nil {
		return fmt.Errorf("ошибка при очищении билетов: %v", err)
	}

	if err := ClearTable(db, "reviews"); err != nil {
		return fmt.Errorf("ошибка при очищении отзывов: %v", err)
	}

	return nil
}
