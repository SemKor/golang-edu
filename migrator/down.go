package migrator

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func Down(conn *pgx.Conn, dir string) {
	current := GetCurrentVersion(conn)
	if current == 0 {
		println("No migrations to down")
		return
	}

	migs := LoadMigrations(dir)
	var migToDown *Migration
	for _, m := range migs {
		if m.Version == current {
			migToDown = &m
			break
		}
	}

	if migToDown == nil || migToDown.DownFile == "" {
		println("No down migration found for version", current)
		return
	}

	ApplyMigration(conn, migToDown.DownFile)

	_, err := conn.Exec(context.Background(), "DELETE FROM go_migrations WHERE version=$1", current)
	if err != nil {
		log.Fatal("Failed to update version after down:", err)
	}
}