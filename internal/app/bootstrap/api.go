package bootstrap

import (
	dbconfig "cw/internal/dataAccess/config"
	repository "cw/internal/dataAccess/repository/postgres"
	"cw/internal/domain/service"
	"cw/internal/dto/handler"
	authMiddleware "cw/internal/middleware"
	"cw/internal/observability"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

type RouterConfig struct {
	JWTSecret       string
	TokenDuration   time.Duration
	SwaggerDocPath  string
	Observability   observability.Config
	EnableChiLogger bool
}

type RouteRegistrars struct {
	UserNoAuth func(chi.Router)
	UserAuth   func(chi.Router)
	Genre      func(chi.Router)
	Hall       func(chi.Router)
	MovieShow  func(chi.Router)
	Movie      func(chi.Router)
	ScreenType func(chi.Router)
	SeatType   func(chi.Router)
	Seat       func(chi.Router)
	Ticket     func(chi.Router)
	Review     func(chi.Router)
}

func NewDatabase() (*pgxpool.Pool, error) {
	return dbconfig.NewDatabase()
}

func BuildAPIRouter(db *pgxpool.Pool, cfg RouterConfig) http.Handler {
	userRepo := repository.NewUserRepository(db, cfg.JWTSecret, cfg.TokenDuration)
	genreRepo := repository.NewGenreRepository(db)
	hallRepo := repository.NewHallRepository(db)
	movieShowRepo := repository.NewMovieShowRepository(db)
	movieRepo := repository.NewMovieRepository(db)
	screenTypeRepo := repository.NewScreenTypeRepository(db)
	seatTypeRepo := repository.NewSeatTypeRepository(db)
	seatRepo := repository.NewSeatRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	userHandler := handler.NewUserHandler(service.NewUserService(userRepo))
	genreHandler := handler.NewGenreHandler(service.NewGenreService(genreRepo))
	hallHandler := handler.NewHallHandler(service.NewHallService(hallRepo))
	movieShowHandler := handler.NewMovieShowHandler(service.NewMovieShowService(movieShowRepo))
	movieHandler := handler.NewMovieHandler(service.NewMovieService(movieRepo))
	screenTypeHandler := handler.NewScreenTypeHandler(service.NewScreenTypeService(screenTypeRepo))
	seatTypeHandler := handler.NewSeatTypeHandler(service.NewSeatTypeService(seatTypeRepo))
	seatHandler := handler.NewSeatHandler(service.NewSeatService(seatRepo))
	ticketHandler := handler.NewTicketHandler(service.NewTicketService(ticketRepo))
	reviewHandler := handler.NewReviewHandler(service.NewReviewService(reviewRepo))

	registrars := RouteRegistrars{
		UserNoAuth: userHandler.RegisterRoutesNoAuth,
		UserAuth:   userHandler.RegisterRoutesWithAuth,
		Genre:      genreHandler.RegisterRoutes,
		Hall:       hallHandler.RegisterRoutes,
		MovieShow:  movieShowHandler.RegisterRoutes,
		Movie:      movieHandler.RegisterRoutes,
		ScreenType: screenTypeHandler.RegisterRoutes,
		SeatType:   seatTypeHandler.RegisterRoutes,
		Seat:       seatHandler.RegisterRoutes,
		Ticket:     ticketHandler.RegisterRoutes,
		Review:     reviewHandler.RegisterRoutes,
	}

	return BuildRouter(cfg, registrars)
}

func BuildRouter(cfg RouterConfig, routes RouteRegistrars) *chi.Mux {
	r := chi.NewRouter()

	if cfg.Observability.TracesEnabled {
		r.Use(observability.HTTPMiddleware(cfg.Observability.ServiceName))
	}
	r.Use(observability.HTTPLoggingMiddleware(cfg.Observability.ServiceName))
	if cfg.EnableChiLogger {
		r.Use(chimiddleware.Logger)
	}
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)

	r.Get("/api/v1/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/api/v1/swagger/doc.json"),
	))
	r.Get("/api/v1/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, cfg.swaggerDocPath())
	})

	r.Mount("/metrics", promhttp.Handler())

	r.Route("/api/v1", func(r chi.Router) {
		if routes.UserNoAuth != nil {
			routes.UserNoAuth(r)
		}
		if routes.Genre != nil {
			routes.Genre(r)
		}
		if routes.Hall != nil {
			routes.Hall(r)
		}
		if routes.MovieShow != nil {
			routes.MovieShow(r)
		}
		if routes.Movie != nil {
			routes.Movie(r)
		}
		if routes.ScreenType != nil {
			routes.ScreenType(r)
		}
		if routes.SeatType != nil {
			routes.SeatType(r)
		}
		if routes.Seat != nil {
			routes.Seat(r)
		}
		if routes.Ticket != nil {
			routes.Ticket(r)
		}
		if routes.Review != nil {
			routes.Review(r)
		}

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.JWTMiddleware(cfg.JWTSecret))
			if routes.UserAuth != nil {
				routes.UserAuth(r)
			}
		})
	})

	return r
}

func (c RouterConfig) swaggerDocPath() string {
	if c.SwaggerDocPath == "" {
		return "./docs/swagger.json"
	}
	return c.SwaggerDocPath
}
