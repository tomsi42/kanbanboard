package store

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Connect opens a connection to PostgreSQL using environment variables.
// It retries up to 10 times with a 2-second delay to handle container startup ordering.
func Connect() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "kanban"),
		getEnv("DB_PASSWORD", "kanban"),
		getEnv("DB_NAME", "kanbanboard"),
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	for i := range 10 {
		if err := db.Ping(); err == nil {
			return db, nil
		}
		log.Printf("Waiting for database (attempt %d/10)...", i+1)
		time.Sleep(2 * time.Second)
	}

	// Final attempt
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database after retries: %w", err)
	}

	return db, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
