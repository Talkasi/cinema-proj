package observability_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"cw/internal/observability"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Benchmarks a lightweight HTTP route (/genres) to capture middleware overhead
// with tracing/logging toggled. No external DB/services are touched.
func BenchmarkHTTPTracingDisabledLoggingDefault(b *testing.B) {
	benchmarkHTTP(b, false, "default")
}

func BenchmarkHTTPTracingDisabledLoggingExtended(b *testing.B) {
	benchmarkHTTP(b, false, "extended")
}

func BenchmarkHTTPTracingEnabledLoggingDefault(b *testing.B) {
	benchmarkHTTP(b, true, "default")
}

func BenchmarkHTTPTracingEnabledLoggingExtended(b *testing.B) {
	benchmarkHTTP(b, true, "extended")
}

func benchmarkHTTP(b *testing.B, tracing bool, logMode string) {
	cfg := baseConfig()
	cfg.TracesEnabled = tracing
	cfg.OutputPath = filepath.Join("artifacts", "observability", fmt.Sprintf("%s-traces.jsonl", b.Name()))

	logCfg := observability.LogConfig{
		Level:     "info",
		Mode:      logMode,
		Output:    filepath.Join("artifacts", "observability", fmt.Sprintf("%s.log", b.Name())),
		Timestamp: true,
	}

	cleanupLog, err := observability.SetupLogging(logCfg)
	if err != nil {
		b.Fatalf("setup logging: %v", err)
	}
	b.Cleanup(func() {
		if err := cleanupLog(); err != nil {
			b.Fatalf("cleanup logging: %v", err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	providers, err := observability.SetupProviders(ctx, cfg)
	if err != nil {
		b.Fatalf("setup providers: %v", err)
	}
	b.Cleanup(func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shutdownCancel()
		if err := providers.Shutdown(shutdownCtx); err != nil {
			b.Fatalf("shutdown providers: %v", err)
		}
	})

	router := chi.NewRouter()
	if tracing {
		router.Use(observability.HTTPMiddleware(cfg.ServiceName))
	}
	router.Use(observability.HTTPLoggingMiddleware(cfg.ServiceName))
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)

router.Get("/genres", func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`[{"id":1,"name":"Drama"},{"id":2,"name":"Comedy"}]`))
})

req := httptest.NewRequest(http.MethodGet, "/genres", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
	}
}
