package observability

import (
	"os"
	"strconv"
	"strings"
)

// Config describes tracing/metrics exporters, sampling and destinations.
type Config struct {
	ServiceName  string
	Environment  string
	Exporter     string
	Endpoint     string
	InsecureOTLP bool

	TracesEnabled  bool
	MetricsEnabled bool

	SampleRatio float64
	OutputPath  string
}

// LogConfig controls the global stdlib logger output and verbosity.
type LogConfig struct {
	Level     string
	Mode      string
	Output    string
	Timestamp bool
}

func LoadConfigFromEnv() Config {
	return Config{
		ServiceName:   firstNonEmpty(os.Getenv("OTEL_SERVICE_NAME"), "cinema-api"),
		Environment:   firstNonEmpty(os.Getenv("OTEL_ENVIRONMENT"), "development"),
		Exporter:      strings.ToLower(firstNonEmpty(os.Getenv("OTEL_EXPORTER"), "stdout")),
		Endpoint:      firstNonEmpty(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"), "localhost:4318"),
		InsecureOTLP:  parseBoolEnv("OTEL_EXPORTER_OTLP_INSECURE", true),
		TracesEnabled: parseBoolEnv("OTEL_TRACES_ENABLED", false),
		MetricsEnabled: parseBoolEnv(
			"OTEL_METRICS_ENABLED",
			false,
		),
		SampleRatio: parseFloatEnv("OTEL_TRACES_SAMPLER_RATIO", 1.0),
		OutputPath:  firstNonEmpty(os.Getenv("OTEL_EXPORTER_STDOUT_PATH"), "artifacts/observability/traces.jsonl"),
	}
}

func LoadLogConfigFromEnv() LogConfig {
	return LogConfig{
		Level:     strings.ToLower(firstNonEmpty(os.Getenv("LOG_LEVEL"), "error")),
		Mode:      strings.ToLower(firstNonEmpty(os.Getenv("LOG_MODE"), "default")),
		Output:    os.Getenv("LOG_OUTPUT"),
		Timestamp: parseBoolEnv("LOG_TIMESTAMP", true),
	}
}

func firstNonEmpty(values ...string) string {
	for _, val := range values {
		if strings.TrimSpace(val) != "" {
			return val
		}
	}
	return ""
}

func parseBoolEnv(key string, def bool) bool {
	val := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if val == "" {
		return def
	}
	switch val {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

func parseFloatEnv(key string, def float64) float64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return def
	}
	return parsed
}
