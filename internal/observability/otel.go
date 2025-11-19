package observability

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.24.0"
)

type Providers struct {
	tracer  *tracesdk.TracerProvider
	meter   *metric.MeterProvider
	closers []io.Closer
}

var activeConfig Config

func ActiveConfig() Config {
	return activeConfig
}

func SetupProviders(ctx context.Context, cfg Config) (*Providers, error) {
	activeConfig = cfg

	if !cfg.TracesEnabled && !cfg.MetricsEnabled {
		otel.SetTracerProvider(tracesdk.NewTracerProvider(tracesdk.WithSampler(tracesdk.NeverSample())))
		otel.SetMeterProvider(noop.NewMeterProvider())
		otel.SetTextMapPropagator(propagation.TraceContext{})
		return &Providers{}, nil
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	p := &Providers{}

	if cfg.TracesEnabled {
		tracer, closers, err := buildTracerProvider(ctx, cfg, res)
		if err != nil {
			return nil, err
		}
		p.tracer = tracer
		p.closers = append(p.closers, closers...)
		otel.SetTracerProvider(tracer)
	} else {
		otel.SetTracerProvider(tracesdk.NewTracerProvider(tracesdk.WithSampler(tracesdk.NeverSample())))
	}

	if cfg.MetricsEnabled {
		meterProvider, closers, err := buildMeterProvider(ctx, cfg, res)
		if err != nil {
			return nil, err
		}
		p.meter = meterProvider
		p.closers = append(p.closers, closers...)
		otel.SetMeterProvider(meterProvider)
	} else {
		otel.SetMeterProvider(noop.NewMeterProvider())
	}

	return p, nil
}

func buildTracerProvider(ctx context.Context, cfg Config, res *resource.Resource) (*tracesdk.TracerProvider, []io.Closer, error) {
	spanExporter, closer, err := buildTraceExporter(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	sampler := tracesdk.TraceIDRatioBased(cfg.SampleRatio)
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(sampler),
		tracesdk.WithBatcher(spanExporter),
		tracesdk.WithResource(res),
	)

	if closer != nil {
		return tp, []io.Closer{closer}, nil
	}
	return tp, nil, nil
}

func buildTraceExporter(ctx context.Context, cfg Config) (tracesdk.SpanExporter, io.Closer, error) {
	switch cfg.Exporter {
	case "otlp", "otlphttp":
		opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(cfg.Endpoint)}
		if cfg.InsecureOTLP {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		exporter, err := otlptracehttp.New(ctx, opts...)
		return exporter, nil, err
	case "otlp-grpc", "otlpgrpc":
		opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.Endpoint)}
		if cfg.InsecureOTLP {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		exporter, err := otlptracegrpc.New(ctx, opts...)
		return exporter, nil, err
	default:
		writer, closer, err := ensureWriter(cfg.OutputPath)
		if err != nil {
			return nil, nil, fmt.Errorf("prepare stdout trace writer: %w", err)
		}
		exporter, err := stdouttrace.New(
			stdouttrace.WithWriter(writer),
			stdouttrace.WithPrettyPrint(),
		)
		return exporter, closer, err
	}
}

func buildMeterProvider(ctx context.Context, cfg Config, res *resource.Resource) (*metric.MeterProvider, []io.Closer, error) {
	readerInterval := time.Second * 15

	switch cfg.Exporter {
	case "otlp", "otlphttp":
		opts := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpoint(cfg.Endpoint)}
		if cfg.InsecureOTLP {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}
		exporter, err := otlpmetrichttp.New(ctx, opts...)
		if err != nil {
			return nil, nil, fmt.Errorf("create otlp metric exporter: %w", err)
		}
		return metric.NewMeterProvider(
			metric.WithResource(res),
			metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(readerInterval))),
		), nil, nil
	case "otlp-grpc", "otlpgrpc":
		opts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(cfg.Endpoint)}
		if cfg.InsecureOTLP {
			opts = append(opts, otlpmetricgrpc.WithInsecure())
		}
		exporter, err := otlpmetricgrpc.New(ctx, opts...)
		if err != nil {
			return nil, nil, fmt.Errorf("create otlp metric exporter: %w", err)
		}
		return metric.NewMeterProvider(
			metric.WithResource(res),
			metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(readerInterval))),
		), nil, nil
	default:
		writer, closer, err := ensureWriter(cfg.OutputPath)
		if err != nil {
			return nil, nil, fmt.Errorf("prepare stdout metric writer: %w", err)
		}
		exporter, err := stdoutmetric.New(stdoutmetric.WithWriter(writer))
		if err != nil {
			return nil, nil, fmt.Errorf("create stdout metric exporter: %w", err)
		}
		return metric.NewMeterProvider(
			metric.WithResource(res),
			metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(readerInterval))),
		), []io.Closer{closer}, nil
	}
}

func ensureWriter(path string) (io.Writer, io.Closer, error) {
	if path == "" {
		return os.Stdout, nil, nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, nil, fmt.Errorf("create directory for %s: %w", path, err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("open writer %s: %w", path, err)
	}

	return file, file, nil
}

func (p *Providers) Shutdown(ctx context.Context) error {
	var shutdownErr error
	if p.tracer != nil {
		if err := p.tracer.Shutdown(ctx); err != nil {
			shutdownErr = errors.Join(shutdownErr, err)
		}
	}
	if p.meter != nil {
		if err := p.meter.Shutdown(ctx); err != nil {
			shutdownErr = errors.Join(shutdownErr, err)
		}
	}
	for _, closer := range p.closers {
		if closer != nil {
			if err := closer.Close(); err != nil {
				shutdownErr = errors.Join(shutdownErr, err)
			}
		}
	}
	return shutdownErr
}
