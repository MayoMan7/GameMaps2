package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"gogamemaps/internal/models"
)

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

func parseStringListFromSteamJSON(raw string) ([]string, error) {
	raw = trimJSONText(raw)
	if raw == "" {
		return []string{}, nil
	}
	var arrStr []string
	if err := json.Unmarshal([]byte(raw), &arrStr); err == nil {
		return filterNonEmpty(arrStr), nil
	}
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
	var s string
	if err := json.Unmarshal([]byte(raw), &s); err == nil {
		if s == "" {
			return []string{}, nil
		}
		return []string{s}, nil
	}
	return []string{}, fmt.Errorf("unrecognized JSON shape: %s", raw)
}

func coerceTags(raw string) map[string]int {
	raw = trimJSONText(raw)
	if raw == "" {
		return map[string]int{}
	}
	var m map[string]int
	if err := json.Unmarshal([]byte(raw), &m); err == nil && m != nil {
		return m
	}
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
	var arrObj []map[string]any
	if err := json.Unmarshal([]byte(raw), &arrObj); err == nil {
		out := make(map[string]int, len(arrObj))
		for _, obj := range arrObj {
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
	var s string
	if err := json.Unmarshal([]byte(raw), &s); err == nil {
		if s == "" {
			return map[string]int{}
		}
		return map[string]int{s: 1}
	}
	return map[string]int{}
}

// GetAllGamesSkipBad returns all games, skipping rows that fail Scan or list-field parsing.
func GetAllGamesSkipBad(ctx context.Context, db *sql.DB) ([]models.Game, int, error) {
	const q = `
		SELECT app_id, name, release_date, short_description, detailed_description,
			about_the_game, header_image, metacritic_score, achievements,
			recommendations, positive, negative,
			COALESCE(tags, '{}'::jsonb)::text, COALESCE(developers, '[]'::jsonb)::text,
			COALESCE(publishers, '[]'::jsonb)::text, COALESCE(categories, '[]'::jsonb)::text,
			COALESCE(genres, '[]'::jsonb)::text
		FROM public.steam_games
		ORDER BY app_id;
	`
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("query all games: %w", err)
	}
	defer rows.Close()

	games := make([]models.Game, 0, 1024)
	skipped := 0

	for rows.Next() {
		var g models.Game
		var tagsJSON, devJSON, pubJSON, catJSON, genreJSON string
		if err := rows.Scan(
			&g.AppID, &g.Name, &g.ReleaseDate, &g.ShortDescription, &g.DetailedDescription,
			&g.AboutTheGame, &g.HeaderImage, &g.MetacriticScore, &g.Achievements,
			&g.Recommendations, &g.Positive, &g.Negative,
			&tagsJSON, &devJSON, &pubJSON, &catJSON, &genreJSON,
		); err != nil {
			skipped++
			log.Printf("[SKIP] scan error: %v", err)
			continue
		}
		g.Tags = coerceTags(tagsJSON)
		if g.Developers, err = parseStringListFromSteamJSON(devJSON); err != nil {
			skipped++
			log.Printf("[SKIP] app_id=%d developers parse error: %v", g.AppID, err)
			continue
		}
		if g.Publishers, err = parseStringListFromSteamJSON(pubJSON); err != nil {
			skipped++
			log.Printf("[SKIP] app_id=%d publishers parse error: %v", g.AppID, err)
			continue
		}
		if g.Categories, err = parseStringListFromSteamJSON(catJSON); err != nil {
			skipped++
			log.Printf("[SKIP] app_id=%d categories parse error: %v", g.AppID, err)
			continue
		}
		if g.Genres, err = parseStringListFromSteamJSON(genreJSON); err != nil {
			skipped++
			log.Printf("[SKIP] app_id=%d genres parse error: %v", g.AppID, err)
			continue
		}
		games = append(games, g)
	}
	if err := rows.Err(); err != nil {
		return games, skipped, fmt.Errorf("iterate games rows: %w", err)
	}
	return games, skipped, nil
}

// GetGameByAppID returns the full game row.
func GetGameByAppID(ctx context.Context, db *sql.DB, appID int64) (*models.Game, error) {
	const q = `
		SELECT app_id, name, release_date, short_description, detailed_description,
			about_the_game, header_image, metacritic_score, achievements,
			recommendations, positive, negative,
			COALESCE(tags, '{}'::jsonb)::text, COALESCE(developers, '[]'::jsonb)::text,
			COALESCE(publishers, '[]'::jsonb)::text, COALESCE(categories, '[]'::jsonb)::text,
			COALESCE(genres, '[]'::jsonb)::text
		FROM public.steam_games
		WHERE app_id = $1 LIMIT 1;
	`
	var g models.Game
	var tagsJSON, devJSON, pubJSON, catJSON, genreJSON string
	err := db.QueryRowContext(ctx, q, appID).Scan(
		&g.AppID, &g.Name, &g.ReleaseDate, &g.ShortDescription, &g.DetailedDescription,
		&g.AboutTheGame, &g.HeaderImage, &g.MetacriticScore, &g.Achievements,
		&g.Recommendations, &g.Positive, &g.Negative,
		&tagsJSON, &devJSON, &pubJSON, &catJSON, &genreJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query game by app_id=%d: %w", appID, err)
	}
	g.Tags = coerceTags(tagsJSON)
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
	return &g, nil
}

// SaveGameEmbedding updates tfidf_embedding for a game.
func SaveGameEmbedding(ctx context.Context, db *sql.DB, appID int64, emb map[string]float64) error {
	b, err := json.Marshal(emb)
	if err != nil {
		return fmt.Errorf("marshal embedding app_id=%d: %w", appID, err)
	}
	const q = `UPDATE public.steam_games SET tfidf_embedding = $1::jsonb WHERE app_id = $2;`
	_, err = db.ExecContext(ctx, q, string(b), appID)
	if err != nil {
		return fmt.Errorf("update embedding app_id=%d: %w", appID, err)
	}
	return nil
}

// GetGameNameByAppID returns the game's name (or empty string if not found).
func GetGameNameByAppID(ctx context.Context, db *sql.DB, appID int64) (string, error) {
	const q = `SELECT name FROM public.steam_games WHERE app_id = $1 LIMIT 1;`
	var name string
	err := db.QueryRowContext(ctx, q, appID).Scan(&name)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get name app_id=%d: %w", appID, err)
	}
	return name, nil
}

func escapeLikeEscapes(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\', '%', '_':
			b = append(b, '\\')
		}
		b = append(b, s[i])
	}
	return string(b)
}

// SearchGameNames returns up to limit games whose names match the query.
func SearchGameNames(ctx context.Context, db *sql.DB, query string, limit int) ([]models.GameSearchResult, error) {
	if limit <= 0 {
		limit = 5
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return []models.GameSearchResult{}, nil
	}
	escaped := escapeLikeEscapes(query)
	pattern := "%" + escaped + "%"
	const q = `
		SELECT app_id, name FROM public.steam_games
		WHERE name ILIKE $1
		ORDER BY
			CASE WHEN LOWER(TRIM(name)) = LOWER(TRIM($2)) THEN 0
			     WHEN LOWER(name) LIKE LOWER(TRIM($2)) || '%%' THEN 1 ELSE 2 END,
			LENGTH(name), name
		LIMIT $3;
	`
	rows, err := db.QueryContext(ctx, q, pattern, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search game names: %w", err)
	}
	defer rows.Close()

	var results []models.GameSearchResult
	for rows.Next() {
		var r models.GameSearchResult
		if err := rows.Scan(&r.AppID, &r.Name); err != nil {
			continue
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search game names: %w", err)
	}
	return results, nil
}
