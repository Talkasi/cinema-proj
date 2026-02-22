package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	postgresDb "cw/internal/dataAccess/config"
	repository "cw/internal/dataAccess/repository/postgres"
	"cw/internal/domain/service"
	"cw/internal/dto/handler"
	"cw/internal/observability"
	"cw/internal/utils"

	authMiddleware "cw/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Курсовая работа по базам данных - API управления кинотеатром
// @version 1.0
// @description Разработка базы данных для управления кинотеатром
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT токен
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logCloser, err := observability.SetupLogging(observability.LoadLogConfigFromEnv())
	if err != nil {
		log.Fatalf("failed to configure logging: %v", err)
	}
	defer func() {
		if closeErr := logCloser(); closeErr != nil {
			log.Printf("failed to close log writer: %v", closeErr)
		}
	}()

	obsCfg := observability.LoadConfigFromEnv()
	providers, err := observability.SetupProviders(ctx, obsCfg)
	if err != nil {
		log.Fatalf("failed to init observability providers: %v", err)
	}
	defer func() {
		if providers != nil {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := providers.Shutdown(shutdownCtx); err != nil {
				log.Printf("failed to shutdown observability providers: %v", err)
			}
		}
	}()

	addr := utils.GetEnv("ADDR")
	jwtSecret := utils.GetEnv("JWT_SECRET")
	tokenDurationStr := utils.GetEnv("TOKEN_DURATION")

	tokenDuration, err := time.ParseDuration(tokenDurationStr)
	if err != nil {
		log.Printf("Invalid token duration, using default 24h: %v", err)
		tokenDuration = 24 * time.Hour
	}

	db, err := postgresDb.NewDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	userRepo := repository.NewUserRepository(db, jwtSecret, tokenDuration)
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

	if obsCfg.TracesEnabled {
		r.Use(observability.HTTPMiddleware(obsCfg.ServiceName))
	}
	r.Use(observability.HTTPLoggingMiddleware(obsCfg.ServiceName))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/api/v1/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/api/v1/swagger/doc.json"),
	))

	r.Get("/api/v1/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.json")
	})

	r.Mount("/metrics", promhttp.Handler())

	r.Route("/api/v1", func(r chi.Router) {
		userHandler.RegisterRoutesNoAuth(r)
		genreHandler.RegisterRoutes(r)
		hallHandler.RegisterRoutes(r)
		movieShowHandler.RegisterRoutes(r)
		movieHandler.RegisterRoutes(r)
		screenTypeHandler.RegisterRoutes(r)
		seatTypeHandler.RegisterRoutes(r)
		seatHandler.RegisterRoutes(r)
		ticketHandler.RegisterRoutes(r)
		reviewHandler.RegisterRoutes(r)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.JWTMiddleware(jwtSecret))

			userHandler.RegisterRoutesWithAuth(r)
		})
	})

	observability.Infof("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
