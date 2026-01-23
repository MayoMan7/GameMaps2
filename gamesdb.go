package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
)

// Game matches the fields you used in the Python TF-IDF build.
// Notes on types:
//   - Tags is a dict of weights: {"Action": 123, ...} (but we coerce other shapes too)
//   - Developers/Publishers/Categories/Genres can be any Steam-ish JSON shape,
//     so we store them as []string after normalization-ish extraction.
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

	// Structured metadata (normalized into list of strings)
	Developers []string `json:"developers"`
	Publishers []string `json:"publishers"`
	Categories []string `json:"categories"`
	Genres     []string `json:"genres"`

	// Weighted tags (best-effort coercion)
	Tags map[string]int `json:"tags"`

	Embeddings []float64 `json:"embeddings,omitempty"`
}

// -----------------------------
// JSON helpers
// -----------------------------

func trimJSONText(s string) string {
	if s == "" || s == "null" {
		return ""
	}
	return s
}

func filterNonEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

// parseStringListFromSteamJSON tries to handle common Steam shapes:
// - ["Action","RPG"]
// - [{"description":"Single-player"}, ...]
// - {"Action": 123, "RPG": 55} (keys become values)
// - "Action" (single string)
// - null / empty
func parseStringListFromSteamJSON(raw string) ([]string, error) {
	raw = trimJSONText(raw)
	if raw == "" {
		return []string{}, nil
	}

	// Try as []string
	var arrStr []string
	if err := json.Unmarshal([]byte(raw), &arrStr); err == nil {
		return filterNonEmpty(arrStr), nil
	}

	// Try as map[string]any (keys become items)
	var obj map[string]any
	if err := json.Unmarshal([]byte(raw), &obj); err == nil {
		out := make([]string, 0, len(obj))
		for k := range obj {
			if k != "" {
				out = append(out, k)
			}
		}
		return filterNonEmpty(out), nil
	}

	// Try as []map[string]any (grab name/description/label/value if present)
	var arrObj []map[string]any
	if err := json.Unmarshal([]byte(raw), &arrObj); err == nil {
		out := make([]string, 0, len(arrObj))
		for _, m := range arrObj {
			for _, key := range []string{"name", "description", "label", "value"} {
				if v, ok := m[key]; ok {
					if s, ok := v.(string); ok && s != "" {
						out = append(out, s)
						break
					}
				}
			}
		}
		return filterNonEmpty(out), nil
	}

	// Try as string
	var s string
	if err := json.Unmarshal([]byte(raw), &s); err == nil {
		if s == "" {
			return []string{}, nil
		}
		return []string{s}, nil
	}

	return []string{}, fmt.Errorf("unrecognized JSON shape: %s", raw)
}

// coerceTags best-effort parses tags into map[string]int.
// Handles shapes:
// - {"Action": 123, "RPG": 55}
// - ["Action","RPG"] -> weight 1
// - [{"description":"Action","count":123}, ...] -> weight if present, else 1
// - {"Action": 123.0} (json numbers -> float64) -> int
// - "Action" -> {"Action": 1}
// - null/empty -> {}
func coerceTags(raw string) map[string]int {
	raw = trimJSONText(raw)
	if raw == "" {
		return map[string]int{}
	}

	// 1) map[string]int
	{
		var m map[string]int
		if err := json.Unmarshal([]byte(raw), &m); err == nil && m != nil {
			return m
		}
	}

	// 2) map[string]any -> coerce numeric values
	{
		var obj map[string]any
		if err := json.Unmarshal([]byte(raw), &obj); err == nil && obj != nil {
			out := make(map[string]int, len(obj))
			for k, v := range obj {
				if k == "" {
					continue
				}
				switch x := v.(type) {
				case float64:
					out[k] = int(x)
				case int:
					out[k] = x
				default:
					out[k] = 1
				}
			}
			return out
		}
	}

	// 3) []string -> weight 1
	{
		var arr []string
		if err := json.Unmarshal([]byte(raw), &arr); err == nil {
			out := make(map[string]int, len(arr))
			for _, t := range arr {
				if t != "" {
					out[t] = 1
				}
			}
			return out
		}
	}

	// 4) []map[string]any
	{
		var arr []map[string]any
		if err := json.Unmarshal([]byte(raw), &arr); err == nil {
			out := make(map[string]int, len(arr))
			for _, obj := range arr {
				// find name
				name := ""
				for _, k := range []string{"tag", "name", "description", "label", "value"} {
					if v, ok := obj[k]; ok {
						if s, ok := v.(string); ok && s != "" {
							name = s
							break
						}
					}
				}
				if name == "" {
					continue
				}

				// find weight
				w := 1
				for _, k := range []string{"count", "votes", "weight", "score"} {
					if v, ok := obj[k]; ok {
						switch x := v.(type) {
						case float64:
							w = int(x)
						case int:
							w = x
						}
						break
					}
				}
				out[name] = w
			}
			return out
		}
	}

	// 5) single string
	{
		var s string
		if err := json.Unmarshal([]byte(raw), &s); err == nil {
			if s == "" {
				return map[string]int{}
			}
			return map[string]int{s: 1}
		}
	}

	// Unknown shape -> empty (don't fail the row)
	return map[string]int{}
}

// -----------------------------
// DB functions
// -----------------------------

// GetAllGamesSkipBad returns all games, skipping only rows that fail Scan() or list-field parsing.
// Tags are parsed best-effort and will NOT cause skipping.
func GetAllGamesSkipBad(ctx context.Context, db *sql.DB) ([]Game, int, error) {
	const q = `
		SELECT
			app_id,
			name,
			release_date,
			short_description,
			detailed_description,
			about_the_game,
			header_image,
			metacritic_score,
			achievements,
			recommendations,
			positive,
			negative,

			COALESCE(tags, '{}'::jsonb)::text AS tags_json,
			COALESCE(developers, '[]'::jsonb)::text AS developers_json,
			COALESCE(publishers, '[]'::jsonb)::text AS publishers_json,
			COALESCE(categories, '[]'::jsonb)::text AS categories_json,
			COALESCE(genres, '[]'::jsonb)::text AS genres_json
		FROM public.steam_games
		ORDER BY app_id;
	`

	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("query all games: %w", err)
	}
	defer rows.Close()

	games := make([]Game, 0, 1024)
	skipped := 0

	for rows.Next() {
		var g Game
		var tagsJSON, devJSON, pubJSON, catJSON, genreJSON string

		// 1) Scan
		if err := rows.Scan(
			&g.AppID,
			&g.Name,
			&g.ReleaseDate,
			&g.ShortDescription,
			&g.DetailedDescription,
			&g.AboutTheGame,
			&g.HeaderImage,
			&g.MetacriticScore,
			&g.Achievements,
			&g.Recommendations,
			&g.Positive,
			&g.Negative,
			&tagsJSON,
			&devJSON,
			&pubJSON,
			&catJSON,
			&genreJSON,
		); err != nil {
			skipped++
			log.Printf("[SKIP] scan error: %v", err)
			continue
		}

		// 2) Tags: best-effort coerce (never skip row for tags)
		g.Tags = coerceTags(tagsJSON)

		// 3) Parse list-ish JSON fields (skip if broken)
		if g.Developers, err = parseStringListFromSteamJSON(devJSON); err != nil {
			skipped++
			log.Printf("[SKIP] app_id=%d developers parse error: %v | raw=%q", g.AppID, err, devJSON)
			continue
		}
		if g.Publishers, err = parseStringListFromSteamJSON(pubJSON); err != nil {
			skipped++
			log.Printf("[SKIP] app_id=%d publishers parse error: %v | raw=%q", g.AppID, err, pubJSON)
			continue
		}
		if g.Categories, err = parseStringListFromSteamJSON(catJSON); err != nil {
			skipped++
			log.Printf("[SKIP] app_id=%d categories parse error: %v | raw=%q", g.AppID, err, catJSON)
			continue
		}
		if g.Genres, err = parseStringListFromSteamJSON(genreJSON); err != nil {
			skipped++
			log.Printf("[SKIP] app_id=%d genres parse error: %v | raw=%q", g.AppID, err, genreJSON)
			continue
		}

		g.Embeddings = nil
		games = append(games, g)
	}

	if err := rows.Err(); err != nil {
		return games, skipped, fmt.Errorf("iterate games rows: %w", err)
	}

	return games, skipped, nil
}

// GetGameByAppID returns the full game row as a Game struct.
// Tags are parsed best-effort (same as GetAllGamesSkipBad) to avoid failures.
func GetGameByAppID(ctx context.Context, db *sql.DB, appID int64) (*Game, error) {
	const q = `
		SELECT
			app_id,
			name,
			release_date,
			short_description,
			detailed_description,
			about_the_game,
			header_image,
			metacritic_score,
			achievements,
			recommendations,
			positive,
			negative,

			COALESCE(tags, '{}'::jsonb)::text AS tags_json,
			COALESCE(developers, '[]'::jsonb)::text AS developers_json,
			COALESCE(publishers, '[]'::jsonb)::text AS publishers_json,
			COALESCE(categories, '[]'::jsonb)::text AS categories_json,
			COALESCE(genres, '[]'::jsonb)::text AS genres_json
		FROM public.steam_games
		WHERE app_id = $1
		LIMIT 1;
	`

	var g Game
	var tagsJSON, devJSON, pubJSON, catJSON, genreJSON string

	err := db.QueryRowContext(ctx, q, appID).Scan(
		&g.AppID,
		&g.Name,
		&g.ReleaseDate,
		&g.ShortDescription,
		&g.DetailedDescription,
		&g.AboutTheGame,
		&g.HeaderImage,
		&g.MetacriticScore,
		&g.Achievements,
		&g.Recommendations,
		&g.Positive,
		&g.Negative,
		&tagsJSON,
		&devJSON,
		&pubJSON,
		&catJSON,
		&genreJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query game by app_id=%d: %w", appID, err)
	}

	// Tags: best-effort coerce
	g.Tags = coerceTags(tagsJSON)

	// developers/publishers/categories/genres -> []string
	if g.Developers, err = parseStringListFromSteamJSON(devJSON); err != nil {
		return nil, fmt.Errorf("parse developers for app_id=%d: %w", appID, err)
	}
	if g.Publishers, err = parseStringListFromSteamJSON(pubJSON); err != nil {
		return nil, fmt.Errorf("parse publishers for app_id=%d: %w", appID, err)
	}
	if g.Categories, err = parseStringListFromSteamJSON(catJSON); err != nil {
		return nil, fmt.Errorf("parse categories for app_id=%d: %w", appID, err)
	}
	if g.Genres, err = parseStringListFromSteamJSON(genreJSON); err != nil {
		return nil, fmt.Errorf("parse genres for app_id=%d: %w", appID, err)
	}

	g.Embeddings = nil
	return &g, nil
}

func SaveGameEmbedding(ctx context.Context, db *sql.DB, appID int64, emb map[string]float64) error {
	b, err := json.Marshal(emb)
	if err != nil {
		return fmt.Errorf("marshal embedding app_id=%d: %w", appID, err)
	}

	const q = `
		UPDATE public.steam_games
		SET tfidf_embedding = $1::jsonb
		WHERE app_id = $2;
	`

	_, err = db.ExecContext(ctx, q, string(b), appID)
	if err != nil {
		return fmt.Errorf("update embedding app_id=%d: %w", appID, err)
	}
	return nil
}
