package observability

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// HTTPMiddleware starts a span for every incoming HTTP request. It is intentionally
// lightweight so it can run even when OTEL exporters are disabled (noop provider).
func HTTPMiddleware(service string) func(http.Handler) http.Handler {
	tracer := otel.Tracer("cinema.http")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			operation := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			start := time.Now()
			wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			ctx, span := tracer.Start(
				r.Context(),
				operation,
				trace.WithAttributes(
					attribute.String("service.name", service),
					attribute.String("http.method", r.Method),
					attribute.String("http.target", r.URL.Path),
				),
			)
			defer span.End()

			next.ServeHTTP(wrapped, r.WithContext(ctx))

			span.SetAttributes(
				attribute.Int("http.status_code", wrapped.Status()),
				attribute.Int64("http.response_size", int64(wrapped.BytesWritten())),
				attribute.Int64("http.duration_ms", time.Since(start).Milliseconds()),
			)
		})
	}
}
