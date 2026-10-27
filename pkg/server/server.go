package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Yandex-Practicum/go_final_project/pkg/api"
)

func Run() error {
	api.Init()
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", http.DefaultServeMux)
	mux.Handle("/", http.FileServer(http.Dir("web")))

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Starting HTTP server on %s, serving web", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen and serve on %s: %w", srv.Addr, err)
	}
	return nil
}
