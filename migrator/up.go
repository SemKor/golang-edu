package migrator

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func ApplyMigration(conn *pgx.Conn, path string) {
	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read migration file %s: %v", path, err)
	}

	_, err = conn.Exec(context.Background(), string(sqlBytes))
	if err != nil {
		log.Fatalf("Failed to execute migration %s: %v", path, err)
	}

	println("Applied migration:", path)
}

func Up(conn *pgx.Conn, dir string) {
	current := GetCurrentVersion(conn)
	migs := LoadMigrations(dir)

	for _, m := range migs {
		if m.Version > current {
			ApplyMigration(conn, m.UpFile)

			
			_, err := conn.Exec(context.Background(), "INSERT INTO go_migrations (version) VALUES($1)", m.Version)
			if err != nil {
				log.Fatal("Failed to update version:", err)
			}
		}
	}
}