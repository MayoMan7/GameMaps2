// Main entry point - runs the web server.
// For other commands: go run ./cmd/embed (TF-IDF pipeline), go run ./cmd/cli (interactive).
package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"gogamemaps/internal/handler"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	s := &handler.Server{DB: db}
	r := mux.NewRouter()

	s.GameRoutes(r)
	s.UserRoutes(r)
	s.AuthRoutes(r)
	s.HomeHandler(r)

	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
