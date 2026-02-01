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
