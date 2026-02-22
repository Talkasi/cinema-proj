//go:build bdd
// +build bdd

package bdd

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"cw/internal/observability"
)

// TestMain wires observability when BDD tests run in CI so traces/metrics
// can be exported without touching individual scenarios.
func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logCloser, err := observability.SetupLogging(observability.LoadLogConfigFromEnv())
	if err != nil {
		log.Printf("failed to setup logging: %v", err)
	}

	providers, err := observability.SetupProviders(ctx, observability.LoadConfigFromEnv())
	if err != nil {
		log.Printf("failed to setup observability providers: %v", err)
	}

	code := m.Run()

	if providers != nil {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := providers.Shutdown(shutdownCtx); err != nil {
			log.Printf("failed to shutdown observability providers: %v", err)
		}
		shutdownCancel()
	}

	if logCloser != nil {
		if err := logCloser(); err != nil {
			log.Printf("failed to close logs: %v", err)
		}
	}

	os.Exit(code)
}
