package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"task5/internal/config"
)

func NewDB(cfg *config.Config) (*sql.DB, error) {
	psql := cfg.Psql

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		psql.Host,
		psql.Port,
		psql.User,
		psql.Pass,
		psql.DBName,
		psql.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}