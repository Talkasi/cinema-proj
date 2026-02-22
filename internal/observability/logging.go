package observability

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

var currentLevel = "error"

// SetupLogging configures the standard library logger so existing log.Printf calls
// are routed consistently for both default and extended verbosity modes.
func SetupLogging(cfg LogConfig) (func() error, error) {
	currentLevel = strings.ToLower(strings.TrimSpace(cfg.Level))
	if currentLevel == "" {
		currentLevel = "error"
	}

	if cfg.Output != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.Output), 0o755); err != nil {
			return nil, fmt.Errorf("create log directory: %w", err)
		}
		file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, fmt.Errorf("open log file: %w", err)
		}

		deferFunc := func() error { return file.Close() }

		if cfg.Mode == "silent" {
			log.SetOutput(file)
			log.SetFlags(log.LstdFlags)
			return deferFunc, nil
		}

		log.SetOutput(file)
		configureFlags(cfg)
		return deferFunc, nil
	}

	log.SetOutput(os.Stdout)
	configureFlags(cfg)

	return func() error { return nil }, nil
}

func configureFlags(cfg LogConfig) {
	flags := 0
	if cfg.Timestamp {
		flags |= log.LstdFlags
	}

	if cfg.Mode == "extended" {
		flags |= log.Lmicroseconds | log.Lshortfile
	}

	log.SetFlags(flags)
}

// Infof writes informational messages when LOG_LEVEL=info. It is a thin wrapper
// over the global stdlib logger to keep compatibility with existing log output.
func Infof(format string, args ...any) {
	if currentLevel == "info" {
		log.Printf("[INFO] "+format, args...)
	}
}

// Errorf writes error-level messages unconditionally.
func Errorf(format string, args ...any) {
	log.Printf("[ERROR] "+format, args...)
}

// HTTPLoggingMiddleware writes INFO logs for each HTTP request when LOG_LEVEL=info.
func HTTPLoggingMiddleware(service string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			Infof("HTTP service=%s method=%s path=%s status=%d duration_ms=%d bytes=%d",
				service,
				r.Method,
				r.URL.Path,
				ww.Status(),
				time.Since(start).Milliseconds(),
				ww.BytesWritten(),
			)
		})
	}
}
