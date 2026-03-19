package main

// CLI-утилита для миграций БД, которая:
// подключается к PostgreSQL
// читает SQL-файлы из папки migrations
// применяет их по порядку (up)
// откатывает (down)
// хранит текущую версию в таблице go_migrations

import (
	"context"
	"golang-edu/config"
	"golang-edu/db"
)

func main() {
	// Загружаем конфиг
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		panic(err)
	}

	// Подключаемся к БД
	conn := db.Connect(cfg)
	defer conn.Close(context.Background())

	// Пример запроса
	var version string
	err = conn.QueryRow(context.Background(), "SELECT version()").Scan(&version)
	if err != nil {
		panic(err)
	}

	println("PostgreSQL version:", version)
}
