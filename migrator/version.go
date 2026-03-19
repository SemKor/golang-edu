package migrator

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func GetCurrentVersion(conn *pgx.Conn) int {
	_, err := conn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS go_migrations (
			id SERIAL PRIMARY KEY,
			version INT NOT NULL,
			created_at TIMESTAMP DEFAULT now() NOT NULL
		)
	`)
	if err != nil {
		log.Fatal("Failed to create go_migrations table:", err)
	}

	var version int
	err = conn.QueryRow(context.Background(), "SELECT COALESCE(MAX(version),0) FROM go_migrations").Scan(&version)
	if err != nil {
		log.Fatal("Failed to get current version:", err)
	}

	return version
}

func Version(conn *pgx.Conn) {
	current := GetCurrentVersion(conn)
	println("Current migration version:", current)
}