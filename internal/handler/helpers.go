package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"gogamemaps/internal/db"
	"gogamemaps/internal/models"
)

func decodeJSONBody(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, payload Payload) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeServerError(w http.ResponseWriter, publicMessage string, err error) {
	if err != nil {
		log.Printf("server error: %v", err)
	}
	writeJSON(w, http.StatusInternalServerError, Payload{Status: "error", Error: publicMessage})
}

func readPathInt64(vars map[string]string, key string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(vars[key]), 10, 64)
}

func (s *Server) requireSessionUser(w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	token := readSessionToken(r)
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, Payload{Status: "error", Error: "Not authenticated"})
		return nil, false
	}

	sessionUser, err := db.GetUserBySession(r.Context(), s.DB, token)
	if err != nil {
		writeServerError(w, "Failed to verify session.", err)
		return nil, false
	}
	if sessionUser == nil {
		writeJSON(w, http.StatusUnauthorized, Payload{Status: "error", Error: "Session expired"})
		return nil, false
	}
	return sessionUser, true
}

func (s *Server) requireUserAccess(w http.ResponseWriter, r *http.Request, targetUserID int64) bool {
	sessionUser, ok := s.requireSessionUser(w, r)
	if !ok {
		return false
	}
	if sessionUser.ID != targetUserID {
		writeJSON(w, http.StatusForbidden, Payload{Status: "error", Error: "Forbidden"})
		return false
	}
	return true
}
