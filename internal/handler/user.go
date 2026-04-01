package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) UserRoutes(r *mux.Router) {
	r.HandleFunc("/users", s.handleCreateUser).Methods("POST")
	r.HandleFunc("/users/{id}", s.handleGetUserByID).Methods("GET")
	r.HandleFunc("/users/{id}", s.handleUpdateUser).Methods("PATCH")
	r.HandleFunc("/users/{userId}/like/{appId}", s.handleAddLikedGame).Methods("POST")
	r.HandleFunc("/users/{id}/recompute-taste", s.handleRecomputeTaste).Methods("POST")
	r.HandleFunc("/users/{id}/recommendations", s.handleGetUserRecommendations).Methods("GET")
	r.HandleFunc("/users/{id}/map", s.handleGetUserMap).Methods("GET")
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	writeNotImplemented(w, r)
}

func (s *Server) handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	writeNotImplemented(w, r)
}

func (s *Server) handleAddLikedGame(w http.ResponseWriter, r *http.Request) {
	writeNotImplemented(w, r)
}

func (s *Server) handleRecomputeTaste(w http.ResponseWriter, r *http.Request) {
	writeNotImplemented(w, r)
}

func (s *Server) handleGetUserRecommendations(w http.ResponseWriter, r *http.Request) {
	writeNotImplemented(w, r)
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	writeNotImplemented(w, r)
}

func (s *Server) handleGetUserMap(w http.ResponseWriter, r *http.Request) {
	writeNotImplemented(w, r)
}
