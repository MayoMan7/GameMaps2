package models

// Game matches the Steam game schema used in the TF-IDF build.
type Game struct {
	AppID               int64  `json:"app_id"`
	Name                string `json:"name"`
	ReleaseDate         string `json:"release_date"`
	ShortDescription    string `json:"short_description"`
	DetailedDescription string `json:"detailed_description"`
	AboutTheGame        string `json:"about_the_game"`
	HeaderImage         string `json:"header_image"`
	MetacriticScore     int    `json:"metacritic_score"`
	Achievements        int    `json:"achievements"`
	Recommendations     int    `json:"recommendations"`
	Positive            int    `json:"positive"`
	Negative            int    `json:"negative"`

	Developers []string       `json:"developers"`
	Publishers []string       `json:"publishers"`
	Categories []string       `json:"categories"`
	Genres     []string       `json:"genres"`
	Tags       map[string]int `json:"tags"`

	Embeddings []float64 `json:"embeddings,omitempty"`
}

// User represents a user with taste profile.
type User struct {
	ID             int64
	Name           string
	Email          string
	GamesLiked     []int64
	TasteEmbedding map[string]float64
}

// GameSearchResult is a minimal game row returned by name search.
type GameSearchResult struct {
	AppID int64  `json:"app_id"`
	Name  string `json:"name"`
}

// SimilarResult represents a game with similarity score.
type SimilarResult struct {
	AppID int64   `json:"app_id"`
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

// MapNode represents a node in the taste profile map.
type MapNode struct {
	ID     string  `json:"id"`
	Label  string  `json:"label"`
	Kind   string  `json:"kind"`
	AppID  int64   `json:"app_id,omitempty"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Score  float64 `json:"score,omitempty"`
	Anchor string  `json:"anchor,omitempty"`
}

// MapEdge represents a relationship between nodes.
type MapEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
}

// MapPayload is the full map data returned for rendering.
type MapPayload struct {
	UserID int64     `json:"user_id"`
	Nodes  []MapNode `json:"nodes"`
	Edges  []MapEdge `json:"edges"`
}
