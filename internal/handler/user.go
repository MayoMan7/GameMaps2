package handler

import (
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
	var req struct {
		Name string `json:"name"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Payload{Status: "error", Error: "Invalid request body"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, Payload{Status: "error", Error: "Name is required"})
		return
	}
	id, err := db.CreateUser(r.Context(), s.DB, req.Name)
	if err != nil {
		writeServerError(w, "Failed to create user.", err)
		return
	}
	writeJSON(w, http.StatusCreated, Payload{Status: "success", Data: map[string]interface{}{"id": id, "name": req.Name}})
}

func (s *Server) handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := readPathInt64(vars, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Payload{Status: "error", Error: "Invalid user ID"})
		return
	}
	if !s.requireUserAccess(w, r, userID) {
		return
	}
	user, err := db.GetUserByID(r.Context(), s.DB, userID)
	if err != nil {
		writeServerError(w, "Failed to load user.", err)
		return
	}
	if user == nil {
		writeJSON(w, http.StatusNotFound, Payload{Status: "error", Error: "User not found"})
		return
	}
	writeJSON(w, http.StatusOK, Payload{Status: "success", Data: map[string]interface{}{
		"id": user.ID, "name": user.Name, "email": user.Email, "games_liked": user.GamesLiked,
	}})
}

func (s *Server) handleAddLikedGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := readPathInt64(vars, "userId")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Payload{Status: "error", Error: "Invalid user ID"})
		return
	}
	appID, err := readPathInt64(vars, "appId")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Payload{Status: "error", Error: "Invalid app ID"})
		return
	}
	if !s.requireUserAccess(w, r, userID) {
		return
	}
	err = db.AddLikedGame(r.Context(), s.DB, userID, appID)
	if err != nil {
		writeServerError(w, "Failed to like game.", err)
		return
	}
	writeJSON(w, http.StatusOK, Payload{Status: "success", Data: map[string]interface{}{
		"user_id": userID, "app_id": appID, "message": "Game added to liked list",
	}})
}

func (s *Server) handleRecomputeTaste(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := readPathInt64(vars, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Payload{Status: "error", Error: "Invalid user ID"})
		return
	}
	if !s.requireUserAccess(w, r, userID) {
		return
	}
	embedding, gamesUsed, err := db.RecomputeAndSaveTasteEmbedding(r.Context(), s.DB, userID)
	if err != nil {
		writeServerError(w, "Failed to recompute taste.", err)
		return
	}
	writeJSON(w, http.StatusOK, Payload{Status: "success", Data: map[string]interface{}{
		"user_id": userID, "games_used": gamesUsed, "embedding": embedding,
	}})
}

func (s *Server) handleGetUserRecommendations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := readPathInt64(vars, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Payload{Status: "error", Error: "Invalid user ID"})
		return
	}
	if !s.requireUserAccess(w, r, userID) {
		return
	}
	results, err := similar.FindGamesForUserTaste(r.Context(), s.DB, userID, 10)
	if err != nil {
		writeServerError(w, "Failed to load recommendations.", err)
		return
	}
	recommendations := make([]models.SimilarResult, len(results))
	for i, res := range results {
		recommendations[i] = models.SimilarResult{AppID: res.AppID, Name: res.Name, Score: res.Score}
	}
	writeJSON(w, http.StatusOK, Payload{Status: "success", Data: recommendations})
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := readPathInt64(vars, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Payload{Status: "error", Error: "Invalid user ID"})
		return
	}
	if !s.requireUserAccess(w, r, userID) {
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Payload{Status: "error", Error: "Invalid request body"})
		return
	}
	if err := db.UpdateUserName(r.Context(), s.DB, userID, req.Name); err != nil {
		writeJSON(w, http.StatusBadRequest, Payload{Status: "error", Error: "Invalid display name."})
		return
	}
	writeJSON(w, http.StatusOK, Payload{Status: "success", Data: map[string]interface{}{
		"id": userID, "name": req.Name,
	}})
}
