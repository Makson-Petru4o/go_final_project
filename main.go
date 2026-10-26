package main

import (
	"log"

	"github.com/Yandex-Practicum/go_final_project/pkg/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
