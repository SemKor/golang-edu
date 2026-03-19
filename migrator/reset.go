package migrator

import "github.com/jackc/pgx/v5"

// Reset откатывает все миграции до версии 0
func Reset(conn *pgx.Conn, dir string) {
	for {
		current := GetCurrentVersion(conn)
		if current == 0 {
			break
		}
		Down(conn, dir)
	}
}