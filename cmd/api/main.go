package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	postgresDb "cw/internal/dataAccess/config"
	repository "cw/internal/dataAccess/repository/postgres"
	"cw/internal/domain/service"
	"cw/internal/dto/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "cw/docs"
)

// @title Курсовая работа по базам данных - API управления кинотеатром
// @version 1.0
// @description Разработка базы данных для управления кинотеатром
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT токен в формате: Bearer <token>
func main() {
	addr := flag.String("addr", ":8080", "address for http server")
	jwtSecret := flag.String("jwt-secret", "your-secret-key", "JWT secret key")
	tokenDuration := flag.Duration("token-duration", 24*time.Hour, "JWT token duration")
	flag.Parse()

	db, err := postgresDb.NewDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	userRepo := repository.NewUserRepository(db, *jwtSecret, *tokenDuration)
	genreRepo := repository.NewGenreRepository(db)
	hallRepo := repository.NewHallRepository(db)
	movieShowRepo := repository.NewMovieShowRepository(db)
	movieRepo := repository.NewMovieRepository(db)
	screenTypeRepo := repository.NewScreenTypeRepository(db)
	seatTypeRepo := repository.NewSeatTypeRepository(db)
	seatRepo := repository.NewSeatRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	userService := service.NewUserService(userRepo)
	genreService := service.NewGenreService(genreRepo)
	hallService := service.NewHallService(hallRepo)
	movieShowService := service.NewMovieShowService(movieShowRepo)
	movieService := service.NewMovieService(movieRepo)
	screenTypeService := service.NewScreenTypeService(screenTypeRepo)
	seatTypeService := service.NewSeatTypeService(seatTypeRepo)
	seatService := service.NewSeatService(seatRepo)
	ticketService := service.NewTicketService(ticketRepo)
	reviewService := service.NewReviewService(reviewRepo)

	userHandler := handler.NewUserHandler(userService)
	genreHandler := handler.NewGenreHandler(genreService)
	hallHandler := handler.NewHallHandler(hallService)
	movieShowHandler := handler.NewMovieShowHandler(movieShowService)
	movieHandler := handler.NewMovieHandler(movieService)
	screenTypeHandler := handler.NewScreenTypeHandler(screenTypeService)
	seatTypeHandler := handler.NewSeatTypeHandler(seatTypeService)
	seatHandler := handler.NewSeatHandler(seatService)
	ticketHandler := handler.NewTicketHandler(ticketService)
	reviewHandler := handler.NewReviewHandler(reviewService)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.json")
	})

	r.Route("/api/v1", func(r chi.Router) {
		userHandler.RegisterRoutes(r)
		genreHandler.RegisterRoutes(r)
		hallHandler.RegisterRoutes(r)
		movieShowHandler.RegisterRoutes(r)
		movieHandler.RegisterRoutes(r)
		screenTypeHandler.RegisterRoutes(r)
		seatTypeHandler.RegisterRoutes(r)
		seatHandler.RegisterRoutes(r)
		ticketHandler.RegisterRoutes(r)
		reviewHandler.RegisterRoutes(r)
	})

	log.Printf("Starting server on %s", *addr)
	if err := http.ListenAndServe(*addr, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
