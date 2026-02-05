// Main entry point - runs the web server.
// For other commands: go run ./cmd/embed (TF-IDF pipeline), go run ./cmd/cli (interactive).
package main

import (
	"database/sql"
	"log"
	"net/http"

	"gogamemaps/internal/handler"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgresql://postgres:5274@localhost:5433/postgres?sslmode=disable")
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

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe("localhost:8080", r))
}
