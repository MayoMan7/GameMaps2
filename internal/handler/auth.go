package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gogamemaps/internal/db"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

const sessionCookieName = "gm_session"

func (s *Server) AuthRoutes(r *mux.Router) {
	r.HandleFunc("/auth/register", s.handleRegister).Methods("POST")
	r.HandleFunc("/auth/login", s.handleLogin).Methods("POST")
	r.HandleFunc("/auth/logout", s.handleLogout).Methods("POST")
	r.HandleFunc("/auth/me", s.handleMe).Methods("GET")
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Invalid request body"})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	if req.Name == "" || req.Email == "" || len(req.Password) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Name, email, and password are required"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Failed to secure password"})
		return
	}
	id, err := db.CreateUserWithAuth(r.Context(), s.DB, req.Name, req.Email, string(hash))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: err.Error()})
		return
	}
	token, expiresAt, err := createSession(r.Context(), s.DB, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: err.Error()})
		return
	}
	setSessionCookie(w, token, expiresAt)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: map[string]interface{}{
		"id": id, "name": req.Name, "email": req.Email,
	}})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Invalid request body"})
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Email and password are required"})
		return
	}
	authUser, err := db.GetUserAuthByEmail(r.Context(), s.DB, req.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: err.Error()})
		return
	}
	if authUser == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Invalid credentials"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(authUser.PasswordHash), []byte(req.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Invalid credentials"})
		return
	}
	token, expiresAt, err := createSession(r.Context(), s.DB, authUser.User.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: err.Error()})
		return
	}
	setSessionCookie(w, token, expiresAt)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: map[string]interface{}{
		"id": authUser.User.ID, "name": authUser.User.Name, "email": authUser.User.Email, "games_liked": authUser.User.GamesLiked,
	}})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	token := readSessionToken(r)
	if token != "" {
		_ = db.DeleteSession(r.Context(), s.DB, token)
	}
	clearSessionCookie(w)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Payload{Status: "success"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	token := readSessionToken(r)
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Not authenticated"})
		return
	}
	user, err := db.GetUserBySession(r.Context(), s.DB, token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: err.Error()})
		return
	}
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Payload{Status: "error", Error: "Session expired"})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Payload{Status: "success", Data: map[string]interface{}{
		"id": user.ID, "name": user.Name, "email": user.Email, "games_liked": user.GamesLiked,
	}})
}

func createSession(ctx context.Context, database *sql.DB, userID int64) (string, time.Time, error) {
	token, err := generateToken(32)
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(14 * 24 * time.Hour)
	if err := db.CreateSession(ctx, database, userID, token, expiresAt); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

func generateToken(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func readSessionToken(r *http.Request) string {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
