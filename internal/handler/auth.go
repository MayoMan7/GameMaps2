package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
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
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Payload{Status: "error", Error: "Invalid request body"})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	if req.Name == "" || req.Email == "" || len(req.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, Payload{Status: "error", Error: "Name, email, and password are required"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeServerError(w, "Failed to secure password.", err)
		return
	}
	id, err := db.CreateUserWithAuth(r.Context(), s.DB, req.Name, req.Email, string(hash))
	if err != nil {
		writeServerError(w, "Failed to register account.", err)
		return
	}
	token, expiresAt, err := createSession(r.Context(), s.DB, id)
	if err != nil {
		writeServerError(w, "Failed to create session.", err)
		return
	}
	setSessionCookie(w, r, token, expiresAt)
	writeJSON(w, http.StatusCreated, Payload{Status: "success", Data: map[string]interface{}{
		"id": id, "name": req.Name, "email": req.Email,
	}})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, Payload{Status: "error", Error: "Invalid request body"})
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, Payload{Status: "error", Error: "Email and password are required"})
		return
	}
	authUser, err := db.GetUserAuthByEmail(r.Context(), s.DB, req.Email)
	if err != nil {
		writeServerError(w, "Login failed.", err)
		return
	}
	if authUser == nil {
		writeJSON(w, http.StatusUnauthorized, Payload{Status: "error", Error: "Invalid credentials"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(authUser.PasswordHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, Payload{Status: "error", Error: "Invalid credentials"})
		return
	}
	token, expiresAt, err := createSession(r.Context(), s.DB, authUser.User.ID)
	if err != nil {
		writeServerError(w, "Failed to create session.", err)
		return
	}
	setSessionCookie(w, r, token, expiresAt)
	writeJSON(w, http.StatusOK, Payload{Status: "success", Data: map[string]interface{}{
		"id": authUser.User.ID, "name": authUser.User.Name, "email": authUser.User.Email, "games_liked": authUser.User.GamesLiked,
	}})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	token := readSessionToken(r)
	if token != "" {
		_ = db.DeleteSession(r.Context(), s.DB, token)
	}
	clearSessionCookie(w, r)
	writeJSON(w, http.StatusOK, Payload{Status: "success"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireSessionUser(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, Payload{Status: "success", Data: map[string]interface{}{
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

func setSessionCookie(w http.ResponseWriter, r *http.Request, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   shouldUseSecureCookie(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   shouldUseSecureCookie(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func shouldUseSecureCookie(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func readSessionToken(r *http.Request) string {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
