package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Yandex-Practicum/go_final_project/pkg/db"
	"github.com/Yandex-Practicum/go_final_project/pkg/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func run() (err error) {
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}
	if err := db.Init(dbFile); err != nil {
		return fmt.Errorf("initialize SQLite: %w", err)
	}
	defer func() {
		err = errors.Join(err, db.Close())
	}()
	return server.Run()
}
