package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type Game struct {
	app_id               int64
	name                 string
	release_date         string
	detailed_description string
	about_the_game       string
	header_image         string
	metacritic_score     int
	achievements         int
	recommendations      int
	positve              int
	negative             int
	tags                 map[string]int
	embedings            []float64 // (optional) only if you add this column later (pgvector or array)
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
		&g.app_id,
		&g.name,
		&g.release_date,
		&g.detailed_description,
		&g.about_the_game,
		&g.header_image,
		&g.metacritic_score,
		&g.achievements,
		&g.recommendations,
		&g.positve,
		&g.negative,
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
		g.tags = map[string]int{}
	} else {
		var tags map[string]int
		if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
			return nil, fmt.Errorf("unmarshal tags for app_id=%d: %w", appID, err)
		}
		g.tags = tags
	}

	// You don't have embeddings in the table yet, so leave nil/empty.
	g.embedings = nil
	fmt.Println(g)

	return &g, nil
}
