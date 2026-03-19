package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"golang-edu/config"
)


func Connect(cfg *config.Config) *pgx.Conn {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	fmt.Println("✅ Successfully connected to PostgreSQL!")
	return conn
}