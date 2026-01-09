package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type Game struct {
	AppID               int64          `json:"app_id"`
	Name                string         `json:"name"`
	ReleaseDate         string         `json:"release_date"`
	DetailedDescription string         `json:"detailed_description"`
	AboutTheGame        string         `json:"about_the_game"`
	HeaderImage         string         `json:"header_image"`
	MetacriticScore     int            `json:"metacritic_score"`
	Achievements        int            `json:"achievements"`
	Recommendations     int            `json:"recommendations"`
	Positive            int            `json:"positive"`
	Negative            int            `json:"negative"`
	Tags                map[string]int `json:"tags"`
	Embeddings          []float64      `json:"embeddings,omitempty"`
}

// GetGameByAppID returns the full game row as a Game struct.
// Notes:
// - tags in your DB is json/jsonb shaped like {"tag": 123, ...}, so we cast to text and json.Unmarshal.
// - embedding is not in your table right now, so we return it as nil (or empty slice).
func GetGameByAppID(ctx context.Context, db *sql.DB, appID int64) (*Game, error) {
	const q = `
		SELECT
		app_id,
		name,
		release_date,
		detailed_description,
		about_the_game,
		header_image,
		metacritic_score,
		achievements,
		recommendations,
		positive,
		negative,
		COALESCE(tags, '{}'::jsonb)::text AS tags_json
		FROM public.steam_games
		WHERE app_id = $1
		LIMIT 1;
	`

	var g Game
	var tagsJSON string

	err := db.QueryRowContext(ctx, q, appID).Scan(
		&g.AppID,
		&g.Name,
		&g.ReleaseDate,
		&g.DetailedDescription,
		&g.AboutTheGame,
		&g.HeaderImage,
		&g.MetacriticScore,
		&g.Achievements,
		&g.Recommendations,
		&g.Positive,
		&g.Negative,
		&tagsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query game by app_id=%d: %w", appID, err)
	}

	// Parse tags JSON into map[string]int
	if tagsJSON == "" || tagsJSON == "null" {
		g.Tags = map[string]int{}
	} else {
		var tags map[string]int
		if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
			return nil, fmt.Errorf("unmarshal tags for app_id=%d: %w", appID, err)
		}
		g.Tags = tags
	}

	// You don't have embeddings in the table yet, so leave nil/empty.
	g.Embeddings = nil
	// fmt.Println(g)

	return &g, nil
}
