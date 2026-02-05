package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gogamemaps/internal/db"
	"gogamemaps/internal/models"
	"gogamemaps/internal/similar"

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
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Invalid request body"})
		return
	}
	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Name is required"})
		return
	}
	id, err := db.CreateUser(r.Context(), s.DB, req.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: map[string]interface{}{"id": id, "name": req.Name}})
}

func (s *Server) handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	var userID int64
	_, err := fmt.Sscanf(vars["id"], "%d", &userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Invalid user ID"})
		return
	}
	user, err := db.GetUserByID(r.Context(), s.DB, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: err.Error()})
		return
	}
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "User not found"})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: map[string]interface{}{
		"id": user.ID, "name": user.Name, "email": user.Email, "games_liked": user.GamesLiked,
	}})
}

func (s *Server) handleAddLikedGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	var userID, appID int64
	_, err := fmt.Sscanf(vars["userId"], "%d", &userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Invalid user ID"})
		return
	}
	_, err = fmt.Sscanf(vars["appId"], "%d", &appID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Invalid app ID"})
		return
	}
	err = db.AddLikedGame(r.Context(), s.DB, userID, appID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: map[string]interface{}{
		"user_id": userID, "app_id": appID, "message": "Game added to liked list",
	}})
}

func (s *Server) handleRecomputeTaste(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	var userID int64
	_, err := fmt.Sscanf(vars["id"], "%d", &userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Invalid user ID"})
		return
	}
	embedding, gamesUsed, err := db.RecomputeAndSaveTasteEmbedding(r.Context(), s.DB, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: map[string]interface{}{
		"user_id": userID, "games_used": gamesUsed, "embedding": embedding,
	}})
}

func (s *Server) handleGetUserRecommendations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	var userID int64
	_, err := fmt.Sscanf(vars["id"], "%d", &userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Invalid user ID"})
		return
	}
	results, err := similar.FindGamesForUserTaste(r.Context(), s.DB, userID, 10)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: err.Error()})
		return
	}
	recommendations := make([]models.SimilarResult, len(results))
	for i, res := range results {
		recommendations[i] = models.SimilarResult{AppID: res.AppID, Name: res.Name, Score: res.Score}
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: recommendations})
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	var userID int64
	_, err := fmt.Sscanf(vars["id"], "%d", &userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Invalid user ID"})
		return
	}
	token := readSessionToken(r)
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Not authenticated"})
		return
	}
	sessionUser, err := db.GetUserBySession(r.Context(), s.DB, token)
	if err != nil || sessionUser == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Not authenticated"})
		return
	}
	if sessionUser.ID != userID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Forbidden"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Invalid request body"})
		return
	}
	if err := db.UpdateUserName(r.Context(), s.DB, userID, req.Name); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: map[string]interface{}{
		"id": userID, "name": req.Name,
	}})
}
