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

type Payload struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
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

func (s *Server) GetRecommendedGamesRoute(r *mux.Router) {
	r.HandleFunc("/recommend/{id}", s.handleGetRecommendedGames).Methods("GET")
}

type RecommendedGame struct {
	AppID int64   `json:"app_id"`
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

func (s *Server) handleGetRecommendedGames(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appID := vars["id"]

	var id int64
	_, err := fmt.Sscanf(appID, "%d", &id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Invalid app ID"})
		return
	}

	// Find top 10 similar games, searching up to 10000 candidates
	results, _, err := FindSimilarGamesFromDB(r.Context(), s.DB, id, 10, 10000)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: err.Error()})
		return
	}

	// Convert to response format
	recommendations := make([]RecommendedGame, len(results))
	for i, res := range results {
		recommendations[i] = RecommendedGame{
			AppID: res.AppID,
			Name:  res.Name,
			Score: res.Score,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: recommendations})
}

// -------------------------
// User Routes
// -------------------------

func (s *Server) CreateUserRoute(r *mux.Router) {
	r.HandleFunc("/users", s.handleCreateUser).Methods("POST")
}

type CreateUserRequest struct {
	Name string `json:"name"`
}

type CreateUserResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req CreateUserRequest
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

	id, err := CreateUser(r.Context(), s.DB, req.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: CreateUserResponse{ID: id, Name: req.Name}})
}

func (s *Server) GetUserByIDRoute(r *mux.Router) {
	r.HandleFunc("/users/{id}", s.handleGetUserByID).Methods("GET")
}

type UserResponse struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	GamesLiked []int64 `json:"games_liked"`
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

	user, err := GetUserByID(r.Context(), s.DB, userID)
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
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: UserResponse{
		ID:         user.ID,
		Name:       user.Name,
		GamesLiked: user.GamesLiked,
	}})
}

func (s *Server) AddLikedGameRoute(r *mux.Router) {
	r.HandleFunc("/users/{userId}/like/{appId}", s.handleAddLikedGame).Methods("POST")
}

func (s *Server) handleAddLikedGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	var userID int64
	_, err := fmt.Sscanf(vars["userId"], "%d", &userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Invalid user ID"})
		return
	}

	var appID int64
	_, err = fmt.Sscanf(vars["appId"], "%d", &appID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Invalid app ID"})
		return
	}

	err = AddLikedGame(r.Context(), s.DB, userID, appID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: map[string]interface{}{
		"user_id": userID,
		"app_id":  appID,
		"message": "Game added to liked list",
	}})
}

func (s *Server) RecomputeTasteRoute(r *mux.Router) {
	r.HandleFunc("/users/{id}/recompute-taste", s.handleRecomputeTaste).Methods("POST")
}

type TasteResponse struct {
	UserID     int64              `json:"user_id"`
	GamesUsed  int                `json:"games_used"`
	Embedding  map[string]float64 `json:"embedding"`
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

	embedding, gamesUsed, err := RecomputeAndSaveTasteEmbedding(r.Context(), s.DB, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: TasteResponse{
		UserID:    userID,
		GamesUsed: gamesUsed,
		Embedding: embedding,
	}})
}

func (s *Server) GetUserRecommendationsRoute(r *mux.Router) {
	r.HandleFunc("/users/{id}/recommendations", s.handleGetUserRecommendations).Methods("GET")
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

	// Find top 10 games for the user, searching up to 10000 candidates
	results, err := FindGamesForUserTaste(r.Context(), s.DB, userID, 10, 10000)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: err.Error()})
		return
	}

	// Convert to response format
	recommendations := make([]RecommendedGame, len(results))
	for i, res := range results {
		recommendations[i] = RecommendedGame{
			AppID: res.AppID,
			Name:  res.Name,
			Score: res.Score,
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: recommendations})
}
