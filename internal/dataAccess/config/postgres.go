package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDatabase() (*pgxpool.Pool, error) {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASS", "postgres")
	dbName := getEnv("DB_NAME", "cinema")
	dbSSL := getEnv("DB_SSL", "disable")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbHost, dbUser, dbPass, dbName, dbPort, dbSSL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database connected successfully to %s:%s/%s", dbHost, dbPort, dbName)
	return db, nil
}

func NewTestDatabase() (*pgxpool.Pool, error) {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASS", "postgres")
	dbName := getEnv("TEST_DB_NAME", "cinema_test")
	dbSSL := getEnv("DB_SSL", "disable")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbHost, dbUser, dbPass, dbName, dbPort, dbSSL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create test connection pool: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping test database: %w", err)
	}

	log.Printf("Test database connected successfully to %s:%s/%s", dbHost, dbPort, dbName)
	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func CleanTestDatabase(db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	script, err := os.ReadFile("../sql/007_clean_database.sql")
	if err != nil {
		return fmt.Errorf("failed to read clean script: %w", err)
	}

	_, err = db.Exec(ctx, string(script))
	if err != nil {
		return fmt.Errorf("failed to clean test database: %w", err)
	}

	log.Println("Test database cleaned successfully")
	return nil
}

func SeedTestDatabase(db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	script, err := os.ReadFile("../sql/008_seed_test_data.sql")
	if err != nil {
		return fmt.Errorf("failed to read seed script: %w", err)
	}

	_, err = db.Exec(ctx, string(script))
	if err != nil {
		return fmt.Errorf("failed to seed test database: %w", err)
	}

	log.Println("Test database seeded successfully")
	return nil
}

func CleanAndSeedTestDatabase(db *pgxpool.Pool) error {
	if err := CleanTestDatabase(db); err != nil {
		return err
	}
	return SeedTestDatabase(db)
}
