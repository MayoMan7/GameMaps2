package handler

// Server is intentionally empty in the barebones build.
type Server struct{}

// Payload is the standard JSON response wrapper.
type Payload struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}
