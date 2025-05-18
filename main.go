package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "cw/docs"
)

// @title Курсовая работа по базам данных
// @version 1.0
// @description Разработка базы данных для управления кинотеатром
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	IsTestMode = false
	err := godotenv.Load("config.env")
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	if err := InitDB(); err != nil {
		log.Fatal("ошибка подключения к БД: ", err)
	}

	defer AdminDB.Close()
	defer UserDB.Close()
	defer GuestDB.Close()

	if err := SeedAll(AdminDB); err != nil {
		log.Fatal("ошибка вставки данных: ", err)
	}

	log.Println("Сервер запущен на http://localhost:8080")
	http.ListenAndServe(":8080", NewRouter())
}

func NewRouter() *http.ServeMux {
	mux := new(http.ServeMux)

	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	mux.HandleFunc("GET /screen-types/search", Midleware(RoleBasedHandler(SearchScreenTypes)))
	mux.HandleFunc("GET /screen-types", Midleware(RoleBasedHandler(GetScreenTypes)))
	mux.HandleFunc("GET /screen-types/{id}", Midleware(RoleBasedHandler(GetScreenTypeByID)))
	mux.HandleFunc("POST /screen-types", Midleware(RoleBasedHandler(CreateScreenType)))
	mux.HandleFunc("PUT /screen-types/{id}", Midleware(RoleBasedHandler(UpdateScreenType)))
	mux.HandleFunc("DELETE /screen-types/{id}", Midleware(RoleBasedHandler(DeleteScreenType)))

	mux.HandleFunc("GET /genres/search", Midleware(RoleBasedHandler(SearchGenres)))
	mux.HandleFunc("GET /genres", Midleware(RoleBasedHandler(GetGenres)))
	mux.HandleFunc("GET /genres/{id}", Midleware(RoleBasedHandler(GetGenreByID)))
	mux.HandleFunc("POST /genres", Midleware(RoleBasedHandler(CreateGenre)))
	mux.HandleFunc("PUT /genres/{id}", Midleware(RoleBasedHandler(UpdateGenre)))
	mux.HandleFunc("DELETE /genres/{id}", Midleware(RoleBasedHandler(DeleteGenre)))

	mux.HandleFunc("GET /halls/by-screen-type", Midleware(RoleBasedHandler(GetHallsByScreenType)))
	mux.HandleFunc("GET /halls/search", Midleware(RoleBasedHandler(SearchHallsByName)))
	mux.HandleFunc("GET /halls", Midleware(RoleBasedHandler(GetHalls)))
	mux.HandleFunc("GET /halls/{id}", Midleware(RoleBasedHandler(GetHallByID)))
	mux.HandleFunc("POST /halls", Midleware(RoleBasedHandler(CreateHall)))
	mux.HandleFunc("PUT /halls/{id}", Midleware(RoleBasedHandler(UpdateHall)))
	mux.HandleFunc("DELETE /halls/{id}", Midleware(RoleBasedHandler(DeleteHall)))

	mux.HandleFunc("GET /movies", Midleware(RoleBasedHandler(GetMovies)))
	mux.HandleFunc("GET /movies/by-title/search", Midleware(RoleBasedHandler(SearchMovies)))
	mux.HandleFunc("GET /movies/by-genres/search", Midleware(RoleBasedHandler(GetMoviesByAllGenres)))
	mux.HandleFunc("GET /movies/{id}", Midleware(RoleBasedHandler(GetMovieByID)))
	mux.HandleFunc("POST /movies", Midleware(RoleBasedHandler(CreateMovie)))
	mux.HandleFunc("PUT /movies/{id}", Midleware(RoleBasedHandler(UpdateMovie)))
	mux.HandleFunc("DELETE /movies/{id}", Midleware(RoleBasedHandler(DeleteMovie)))

	mux.HandleFunc("GET /movie-shows/upcoming", Midleware(RoleBasedHandler(GetUpcomingShows)))
	mux.HandleFunc("GET /movie-shows/by-date/{date}", Midleware(RoleBasedHandler(GetShowsByDate)))
	mux.HandleFunc("GET /movies/{movie_id}/shows", Midleware(RoleBasedHandler(GetShowsByMovie)))
	mux.HandleFunc("GET /movie-shows", Midleware(RoleBasedHandler(GetMovieShows)))
	mux.HandleFunc("GET /movie-shows/{id}", Midleware(RoleBasedHandler(GetMovieShowByID)))
	mux.HandleFunc("POST /movie-shows", Midleware(RoleBasedHandler(CreateMovieShow)))
	mux.HandleFunc("PUT /movie-shows/{id}", Midleware(RoleBasedHandler(UpdateMovieShow)))
	mux.HandleFunc("DELETE /movie-shows/{id}", Midleware(RoleBasedHandler(DeleteMovieShow)))

	mux.HandleFunc("GET /users/{user_id}/reviews", Midleware(RoleBasedHandler(GetReviewsByUserID)))
	mux.HandleFunc("GET /movies/{movie_id}/reviews", Midleware(RoleBasedHandler(GetReviewsByMovieID)))
	mux.HandleFunc("GET /reviews", Midleware(RoleBasedHandler(GetReviews)))
	mux.HandleFunc("GET /reviews/{id}", Midleware(RoleBasedHandler(GetReviewByID)))
	mux.HandleFunc("POST /reviews", Midleware(RoleBasedHandler(CreateReview)))
	mux.HandleFunc("PUT /reviews/{id}", Midleware(RoleBasedHandler(UpdateReview)))
	mux.HandleFunc("DELETE /reviews/{id}", Midleware(RoleBasedHandler(DeleteReview)))

	mux.HandleFunc("GET /halls/{hall_id}/seats", Midleware(RoleBasedHandler(GetSeatsByHallID)))
	mux.HandleFunc("GET /seats", Midleware(RoleBasedHandler(GetSeats)))
	mux.HandleFunc("GET /seats/{id}", Midleware(RoleBasedHandler(GetSeatByID)))
	mux.HandleFunc("POST /seats", Midleware(RoleBasedHandler(CreateSeat)))
	mux.HandleFunc("PUT /seats/{id}", Midleware(RoleBasedHandler(UpdateSeat)))
	mux.HandleFunc("DELETE /seats/{id}", Midleware(RoleBasedHandler(DeleteSeat)))

	mux.HandleFunc("GET /seat-types/search", Midleware(RoleBasedHandler(SearchSeatTypes)))
	mux.HandleFunc("GET /seat-types", Midleware(RoleBasedHandler(GetSeatTypes)))
	mux.HandleFunc("GET /seat-types/{id}", Midleware(RoleBasedHandler(GetSeatTypeByID)))
	mux.HandleFunc("POST /seat-types", Midleware(RoleBasedHandler(CreateSeatType)))
	mux.HandleFunc("PUT /seat-types/{id}", Midleware(RoleBasedHandler(UpdateSeatType)))
	mux.HandleFunc("DELETE /seat-types/{id}", Midleware(RoleBasedHandler(DeleteSeatType)))

	mux.HandleFunc("GET /tickets/movie-show/{movie_show_id}", Midleware(RoleBasedHandler(GetTicketsByMovieShowID)))
	mux.HandleFunc("GET /tickets/user/{user_id}", Midleware(RoleBasedHandler(GetTicketsByUserID)))
	mux.HandleFunc("GET /tickets/{id}", Midleware(RoleBasedHandler(GetTicketByID)))
	mux.HandleFunc("POST /tickets", Midleware(RoleBasedHandler(CreateTicket)))
	mux.HandleFunc("PUT /tickets/{id}", Midleware(RoleBasedHandler(UpdateTicket)))
	mux.HandleFunc("DELETE /tickets/{id}", Midleware(RoleBasedHandler(DeleteTicket)))

	mux.HandleFunc("POST /user/register", Midleware(RoleBasedHandler(RegisterUser)))
	mux.HandleFunc("POST /user/login", Midleware(RoleBasedHandler(LoginUser)))
	mux.HandleFunc("GET /users", Midleware(RoleBasedHandler(GetUsers)))
	mux.HandleFunc("GET /users/{id}", Midleware(RoleBasedHandler(GetUserByID)))
	mux.HandleFunc("PUT /users/{id}", Midleware(RoleBasedHandler(UpdateUser)))
	mux.HandleFunc("DELETE /users/{id}", Midleware(RoleBasedHandler(DeleteUser)))

	return mux
}

var (
	GuestDB *pgxpool.Pool
	UserDB  *pgxpool.Pool
	AdminDB *pgxpool.Pool
)

func InitDB() error {
	var err error

	ctx := context.Background()

	GuestDB, err = pgxpool.New(ctx, fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("GUEST_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("GUEST_PASSWORD")))
	if err != nil {
		return fmt.Errorf("ошибка подключения гостя: %v", err)
	}

	UserDB, err = pgxpool.New(ctx, fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("USER_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("USER_PASSWORD")))
	if err != nil {
		return fmt.Errorf("ошибка подключения пользователя: %v", err)
	}

	AdminDB, err = pgxpool.New(ctx, fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("ADMIN_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("ADMIN_PASSWORD")))
	if err != nil {
		return fmt.Errorf("ошибка подключения администратора: %v", err)
	}

	return nil
}

func InitTestDB() error {
	var err error
	ctx := context.Background()

	TestGuestDB, err = pgxpool.New(ctx, fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("TEST_GUEST_USER"),
		os.Getenv("TEST_DB_NAME"),
		os.Getenv("TEST_GUEST_PASSWORD")))
	if err != nil {
		return fmt.Errorf("ошибка подключения гостя: %v", err)
	}

	TestUserDB, err = pgxpool.New(ctx, fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("TEST_USER_USER"),
		os.Getenv("TEST_DB_NAME"),
		os.Getenv("TEST_USER_PASSWORD")))
	if err != nil {
		return fmt.Errorf("ошибка подключения пользователя: %v", err)
	}

	TestAdminDB, err = pgxpool.New(ctx, fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("TEST_ADMIN_USER"),
		os.Getenv("TEST_DB_NAME"),
		os.Getenv("TEST_ADMIN_PASSWORD")))
	if err != nil {
		return fmt.Errorf("ошибка подключения администратора: %v", err)
	}

	return nil
}

var SecretKey = []byte(os.Getenv("TOKEN_KEY"))

type Claims struct {
	Role   string `json:"role"`
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

func GenerateToken(role string, user_id string) (string, error) {
	claims := Claims{
		Role:   role,
		UserID: user_id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour / 2).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SecretKey)
}

func validRoleClaim(db *pgxpool.Pool, userID string, claimed_is_admin bool) (bool, error) {
	var is_admin bool
	err := db.QueryRow(context.Background(), "SELECT is_admin FROM users WHERE id = $1", userID).Scan(&is_admin)
	if err != nil {
		return false, err
	}
	return is_admin == claimed_is_admin, nil
}

func Midleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString != "" {
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return SecretKey, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Неверный токен", http.StatusForbidden)
				return
			}

			r.Header.Set("UserID", claims.UserID)
			r.Header.Set("Role", claims.Role)
			// println("claims.Role", claims.Role)
			// println("claims.UserID", claims.UserID)
		} else {
			r.Header.Set("Role", os.Getenv("CLAIM_ROLE_GUEST"))
			// println("claims.Role", os.Getenv("CLAIM_ROLE_GUEST"))
		}
		next.ServeHTTP(w, r)
	}
}

var IsTestMode bool
var (
	TestGuestDB *pgxpool.Pool
	TestUserDB  *pgxpool.Pool
	TestAdminDB *pgxpool.Pool
)

func RoleBasedHandler(handler func(db *pgxpool.Pool) http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role := r.Header.Get("Role")
		user_id := r.Header.Get("UserID")

		// println("RoleBasedHandler role", role)
		// println("RoleBasedHandler user_id", user_id)

		var db *pgxpool.Pool

		switch role {
		case os.Getenv("CLAIM_ROLE_ADMIN"):
			db = AdminDB
			if IsTestMode {
				db = TestAdminDB
			}

			ok, err := validRoleClaim(db, user_id, true)
			if err != nil {
				if isNoRows(err) {
					http.Error(w, "Token error", http.StatusForbidden)
					return
				}
				http.Error(w, fmt.Sprintf("Fatal database error %v", err), http.StatusInternalServerError)
				return
			}

			if !ok {
				http.Error(w, "Token error", http.StatusForbidden)
				return
			}
		case os.Getenv("CLAIM_ROLE_USER"):
			db = UserDB
			if IsTestMode {
				db = TestUserDB
			}

			ok, err := validRoleClaim(db, user_id, false)
			if err != nil {
				if isNoRows(err) {
					http.Error(w, "Token error", http.StatusForbidden)
					return
				}
				http.Error(w, fmt.Sprintf("Fatal database error %v", err), http.StatusInternalServerError)
				return
			}

			if !ok {
				http.Error(w, "Token error", http.StatusForbidden)
				return
			}
		default:
			db = GuestDB
			if IsTestMode {
				db = TestGuestDB
			}
		}

		if db == nil {
			http.Error(w, "Database connection not available", http.StatusInternalServerError)
			return
		}

		handler(db)(w, r)
	}
}
