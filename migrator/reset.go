package migrator

import "github.com/jackc/pgx/v5"

func Reset(conn *pgx.Conn, dir string) {
	for {
		current := GetCurrentVersion(conn)
		if current == 0 {
			break
		}
		Down(conn, dir)
	}
}