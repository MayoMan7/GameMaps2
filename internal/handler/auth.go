package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) AuthRoutes(r *mux.Router) {
	r.HandleFunc("/auth/register", s.handleRegister).Methods("POST")
	r.HandleFunc("/auth/login", s.handleLogin).Methods("POST")
	r.HandleFunc("/auth/logout", s.handleLogout).Methods("POST")
	r.HandleFunc("/auth/me", s.handleMe).Methods("GET")
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	writeNotImplemented(w, r)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	writeNotImplemented(w, r)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	writeNotImplemented(w, r)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	writeNotImplemented(w, r)
}
