package test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"task5/cmd"
	"task5/internal/config"
)

var testDb *sqlx.DB

const testServerURL = "http://localhost:7073"

func TestMain(m *testing.M) {
	go cmd.NewApi(&config.Config{
		Mode: "api",
		Server: config.AuthServer{
			Host: "localhost",
			Port: 7073,
		},
		Log: config.Log{
			Title:  "api",
			Format: "text",
			Level:  "debug",
		},
		TestUser: config.TestUser{
			Allowed:  true,
			Name:     "tst",
			Password: "tst",
		},
		Psql: config.Psql{
			User:    "postgres",
			Pass:    "postgres",
			Host:    "localhost",
			Port:    5433,
			DBName:  "task_5",
			SSLMode: "disable",
		},
	})

	time.Sleep(2 * time.Second)

	databaseURL := "postgres://postgres:postgres@localhost:5433/task_5?sslmode=disable"
	var err error
	testDb, err = sqlx.Connect("postgres", databaseURL)
	if err != nil {
		panic(fmt.Sprintf("cannot connect to DB, are u started \"task dk-start\"? conn error: %v", err))
	}

	code := m.Run()

	_ = testDb.Close()
	os.Exit(code)
}