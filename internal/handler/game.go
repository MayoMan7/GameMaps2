package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) GameRoutes(r *mux.Router) {
	r.HandleFunc("/getgame/{id}", s.handleGetGameByAppID).Methods("GET")
	r.HandleFunc("/recommend/{id}", s.handleGetRecommendedGames).Methods("GET")
	r.HandleFunc("/search", s.handleSearchGames).Methods("GET")
}

func (s *Server) handleGetGameByAppID(w http.ResponseWriter, r *http.Request) {
	writeNotImplemented(w, r)
}

func (s *Server) handleGetRecommendedGames(w http.ResponseWriter, r *http.Request) {
	writeNotImplemented(w, r)
}

func (s *Server) handleSearchGames(w http.ResponseWriter, r *http.Request) {
	writeNotImplemented(w, r)
}
