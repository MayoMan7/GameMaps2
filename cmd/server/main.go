package main

import (
	"log"
	"net/http"
	"os"

	"gogamemaps/internal/handler"

	"github.com/gorilla/mux"
)

func main() {
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	s := &handler.Server{}
	r := mux.NewRouter()

	s.GameRoutes(r)
	s.UserRoutes(r)
	s.AuthRoutes(r)
	s.HomeHandler(r)

	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
