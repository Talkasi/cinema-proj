package observability_test

import (
	"path/filepath"

	"cw/internal/observability"
)

// func BenchmarkTracingDisabled(b *testing.B) {
// 	cfg := baseConfig()
// 	cfg.TracesEnabled = false
// 	cfg.MetricsEnabled = false
// 	benchmarkTracing(b, cfg)
// }

// func BenchmarkTracingEnabled(b *testing.B) {
// 	cfg := baseConfig()
// 	cfg.TracesEnabled = true
// 	cfg.OutputPath = filepath.Join("artifacts", "observability", fmt.Sprintf("%s-traces.jsonl", b.Name()))
// 	benchmarkTracing(b, cfg)
// }

// func BenchmarkLoggingDefault(b *testing.B) {
// 	cfg := observability.LogConfig{
// 		Mode:      "default",
// 		Timestamp: true,
// 		Output:    filepath.Join("artifacts", "observability", fmt.Sprintf("%s.log", b.Name())),
// 	}
// 	benchmarkLogging(b, cfg)
// }

// func BenchmarkLoggingExtended(b *testing.B) {
// 	cfg := observability.LogConfig{
// 		Mode:      "extended",
// 		Timestamp: true,
// 		Output:    filepath.Join("artifacts", "observability", fmt.Sprintf("%s.log", b.Name())),
// 	}
// 	benchmarkLogging(b, cfg)
// }

// func benchmarkTracing(b *testing.B, cfg observability.Config) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	providers, err := observability.SetupProviders(ctx, cfg)
// 	if err != nil {
// 		b.Fatalf("failed to setup providers: %v", err)
// 	}
// 	b.Cleanup(func() {
// 		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
// 		defer shutdownCancel()
// 		if err := providers.Shutdown(shutdownCtx); err != nil {
// 			log.Printf("shutdown error: %v", err)
// 		}
// 	})

// 	tracer := otel.Tracer("bench.tracer")
// 	b.ReportAllocs()
// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		ctx, span := tracer.Start(context.Background(), "benchmark-operation")
// 		span.AddEvent("cache-check")
// 		span.AddEvent("db-fetch")
// 		span.End()
// 		_ = ctx
// 	}
// }

// func benchmarkLogging(b *testing.B, cfg observability.LogConfig) {
// 	cleanup, err := observability.SetupLogging(cfg)
// 	if err != nil {
// 		b.Fatalf("failed to init logging: %v", err)
// 	}
// 	b.Cleanup(func() {
// 		if err := cleanup(); err != nil {
// 			b.Fatalf("failed to close log writer: %v", err)
// 		}
// 	})

// 	b.ReportAllocs()
// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		log.Printf("bench-event-%d: user_id=%d action=reserve seat=%d", i%3, i%10, i%5)
// 	}
// }

func baseConfig() observability.Config {
	// Benchmarks are self-contained and do not rely on env so results are stable
	// even when CI/user env differs from defaults used by the application.
	return observability.Config{
		ServiceName:    "cinema-api",
		Environment:    "bench",
		Exporter:       "stdout",
		Endpoint:       "localhost:4318",
		InsecureOTLP:   true,
		TracesEnabled:  false, // toggled in benchmarks
		MetricsEnabled: false,
		SampleRatio:    1.0,
		OutputPath:     filepath.Join("artifacts", "observability", "traces.jsonl"),
	}
}
