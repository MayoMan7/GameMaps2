package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) HomeHandler(r *mux.Router) {
	r.HandleFunc("/", s.handleHome).Methods("GET")
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to the Game Database API"))
}
