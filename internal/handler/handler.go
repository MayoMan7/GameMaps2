package handler

import (
	"database/sql"
)

// Server holds dependencies for HTTP handlers.
type Server struct {
	DB *sql.DB
}

// Payload is the standard JSON response wrapper.
type Payload struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}
