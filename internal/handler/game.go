package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gogamemaps/internal/db"
	"gogamemaps/internal/models"
	"gogamemaps/internal/similar"

	"github.com/gorilla/mux"
)

func (s *Server) GameRoutes(r *mux.Router) {
	r.HandleFunc("/getgame/{id}", s.handleGetGameByAppID).Methods("GET")
	r.HandleFunc("/recommend/{id}", s.handleGetRecommendedGames).Methods("GET")
	r.HandleFunc("/search", s.handleSearchGames).Methods("GET")
}

func (s *Server) handleGetGameByAppID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id int64
	_, err := fmt.Sscanf(vars["id"], "%d", &id)
	if err != nil {
		http.Error(w, "Invalid app ID", http.StatusBadRequest)
		return
	}
	game, err := db.GetGameByAppID(r.Context(), s.DB, id)
	if err != nil {
		http.Error(w, "Failed to retrieve game", http.StatusInternalServerError)
		return
	}
	if game == nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: game})
}

func (s *Server) handleGetRecommendedGames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	var id int64
	_, err := fmt.Sscanf(vars["id"], "%d", &id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Invalid app ID"})
		return
	}
	results, _, err := similar.FindSimilarGamesFromDB(r.Context(), s.DB, id, 10, 10000)
	if err != nil {
		writeServerError(w, "Failed to load recommendations.", err)
		return
	}
	recommendations := make([]models.SimilarResult, len(results))
	for i, res := range results {
		recommendations[i] = models.SimilarResult{AppID: res.AppID, Name: res.Name, Score: res.Score}
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: recommendations})
}

func (s *Server) handleSearchGames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query().Get("q")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Missing query parameter 'q'"})
		return
	}
	limit := 8
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 && v <= 25 {
			limit = v
		}
	}
	results, err := db.SearchGameNames(r.Context(), s.DB, query, limit)
	if err != nil {
		writeServerError(w, "Search failed.", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: results})
}
