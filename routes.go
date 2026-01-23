package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) homeHandler(r *mux.Router) {
	r.HandleFunc("/", s.handleHome).Methods("GET")
}
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to the Game Database API"))
}

func (s *Server) createUserRoute(r *mux.Router) {
	r.HandleFunc("/users", s.handleCreateUser).Methods("POST")
}

type Payload struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	// parse request body -> user
	// then call your db function:
	// err := AddUser(r.Context(), s.DB, user)
	// handle err / respond
	username := r.FormValue("name")
	var usertoadd User = User{Name: username}
	err := AddUser(r.Context(), s.DB, usertoadd)
	if err != nil {
		http.Error(w, "Failed to add user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) GetGameByAppIDRoute(r *mux.Router) {
	r.HandleFunc("/getgame/{id}", s.handleGetGameByAppID).Methods("GET")
}

func (s *Server) handleGetGameByAppID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// parse id from vars and call GetGameByAppID
	// respond with game data or error
	appID := vars["id"]
	// Convert appID to int64
	var id int64
	_, err := fmt.Sscanf(appID, "%d", &id)
	if err != nil {
		http.Error(w, "Invalid app ID", http.StatusBadRequest)
		return
	}
	var game *Game
	game, err = GetGameByAppID(r.Context(), s.DB, id)
	if err != nil {
		http.Error(w, "Failed to retrieve game", http.StatusInternalServerError)
		return
	}
	if game == nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}
	fmt.Println(game.Name)
	p := Payload{Status: "success", Data: game, Error: ""}
	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
