package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type APIConfig struct {
	Addr          string
	JWTSecret     string
	TokenDuration time.Duration
}

func GetEnvDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func GetEnvRequired(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("required environment variable %s is not set", key)
	}

	return value, nil
}

func LoadAPIConfigFromEnv() APIConfig {
	cfg := APIConfig{
		Addr:      GetEnvDefault("ADDR", ""),
		JWTSecret: GetEnvDefault("JWT_SECRET", ""),
	}

	tokenDurationStr := GetEnvDefault("TOKEN_DURATION", "")
	if tokenDurationStr == "" {
		cfg.TokenDuration = 24 * time.Hour
		return cfg
	}

	tokenDuration, err := time.ParseDuration(tokenDurationStr)
	if err != nil {
		cfg.TokenDuration = 24 * time.Hour
		return cfg
	}

	cfg.TokenDuration = tokenDuration
	return cfg
}
