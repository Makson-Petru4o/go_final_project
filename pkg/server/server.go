// Package server configures and starts the application's HTTP server.
package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func Run() error {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           http.FileServer(http.Dir("web")),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Starting HTTP server on %s, serving web", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen and serve on %s: %w", srv.Addr, err)
	}
	return nil
}
