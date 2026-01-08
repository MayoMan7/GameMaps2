package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Server struct {
	DB *sql.DB
}

func main() {
	db, err := sql.Open("postgres", "postgresql://postgres:5274@localhost:5433/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Optional but good to verify connection early
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	s := &Server{DB: db}

	r := mux.NewRouter()
	s.createUserRoute(r)
	s.GetGameByAppIDRoute(r)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe("localhost:8080", r))
}
