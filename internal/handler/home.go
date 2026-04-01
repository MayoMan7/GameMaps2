package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) HomeHandler(r *mux.Router) {
	r.HandleFunc("/", s.handleHome).Methods("GET")
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Payload{
		Status: "success",
		Data:   map[string]string{"message": "Barebones API"},
	})
}
