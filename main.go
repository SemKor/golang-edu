package main

import (
	"context"
	"fmt"
	"os"

	"golang-edu/config"
	"golang-edu/db"
	"golang-edu/migrator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [up|down|reset|version]")
		return
	}

	cmd := os.Args[1]

	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		panic(err)
	}

	conn := db.Connect(cfg)
	defer conn.Close(context.Background())

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
