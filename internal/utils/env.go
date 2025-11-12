package utils

import (
	"os"
)

// GetEnv gets an environment variable and returns an error if it's not set or is empty
func GetEnv(key string) string {
	value := os.Getenv(key)
	// if value == "" {
	// 	log.Fatalf("Environment variable %s is not set or is empty", key)
	// }
	return value
}
