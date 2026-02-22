package main

import (
	"context"
	appbootstrap "cw/internal/app/bootstrap"
	appconfig "cw/internal/config"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"cw/internal/observability"
)

// @title Database Course Project - Cinema Management API
// @version 1.0
// @description Database development for cinema management
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT token
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := appconfig.LoadAPIConfigFromEnv()

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

	db, err := appbootstrap.NewDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	router := appbootstrap.BuildAPIRouter(db, appbootstrap.RouterConfig{
		JWTSecret:       cfg.JWTSecret,
		TokenDuration:   cfg.TokenDuration,
		SwaggerDocPath:  "./docs/swagger.json",
		Observability:   obsCfg,
		EnableChiLogger: true,
	})

	observability.Infof("Starting server on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
