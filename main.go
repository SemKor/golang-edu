package main

// CLI-утилита для миграций БД, которая:
// подключается к PostgreSQL
// читает SQL-файлы из папки migrations
// применяет их по порядку (up)
// откатывает (down)
// хранит текущую версию в таблице go_migrations

import (
	"context"
	"fmt"
	"os"

	"golang-edu/config"
	"golang-edu/db"
	"golang-edu/migrator"
)

func main() {
	// Проверяем, передана ли команда
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [up|down|reset|version]")
		return
	}

	cmd := os.Args[1]

	// Загружаем конфиг
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		panic(err)
	}

	// Подключаемся к БД
	conn := db.Connect(cfg)
	defer conn.Close(context.Background())

	// Выполняем команду
	switch cmd {
	case "up":
		migrator.Up(conn, "migrations")
	case "down":
		migrator.Down(conn, "migrations")
	case "reset":
		migrator.Reset(conn, "migrations")
	case "version":
		migrator.Version(conn)
	default:
		fmt.Println("Unknown command:", cmd)
	}
}
