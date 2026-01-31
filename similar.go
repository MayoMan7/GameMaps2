package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
)

type DBSimilarResult struct {
	AppID int64
	Name  string
	Score float64
}

func loadEmbeddingByAppID(ctx context.Context, db *sql.DB, appID int64) (string, map[string]float64, error) {
	const q = `
		SELECT name, COALESCE(tfidf_embedding, '{}'::jsonb)::text
		FROM public.steam_games
		WHERE app_id = $1
		LIMIT 1;
	`

	var name string
	var embJSON string
	if err := db.QueryRowContext(ctx, q, appID).Scan(&name, &embJSON); err != nil {
		if err == sql.ErrNoRows {
			return "", nil, nil
		}
		return "", nil, fmt.Errorf("load embedding app_id=%d: %w", appID, err)
	}

	emb := make(map[string]float64)
	if embJSON != "" && embJSON != "null" && embJSON != "{}" {
		if err := json.Unmarshal([]byte(embJSON), &emb); err != nil {
			return "", nil, fmt.Errorf("unmarshal embedding app_id=%d: %w", appID, err)
		}
	}
	return name, emb, nil
}

func FindSimilarGamesFromDB(ctx context.Context, db *sql.DB, targetAppID int64, topK int, candidateLimit int) ([]DBSimilarResult, map[string]float64, error) {
	if topK <= 0 {
		return nil, nil, nil
	}

	targetName, targetEmb, err := loadEmbeddingByAppID(ctx, db, targetAppID)
	if err != nil {
		return nil, nil, err
	}
	if targetEmb == nil {
		return nil, nil, fmt.Errorf("target app_id=%d not found", targetAppID)
	}
	if len(targetEmb) == 0 {
		return nil, targetEmb, fmt.Errorf("target app_id=%d (%s) has empty embedding in DB", targetAppID, targetName)
	}

	q := `
		SELECT app_id, name, COALESCE(tfidf_embedding, '{}'::jsonb)::text
		FROM public.steam_games
		WHERE tfidf_embedding IS NOT NULL
		  AND app_id <> $1
	`
	if candidateLimit > 0 {
		q += fmt.Sprintf("\nLIMIT %d", candidateLimit)
	}
	q += ";"

	rows, err := db.QueryContext(ctx, q, targetAppID)
	if err != nil {
		return nil, targetEmb, fmt.Errorf("query candidates: %w", err)
	}
	defer rows.Close()

	results := make([]DBSimilarResult, 0, topK*4)

	for rows.Next() {
		var appID int64
		var name, embJSON string
		if err := rows.Scan(&appID, &name, &embJSON); err != nil {
			continue
		}

		emb := make(map[string]float64)
		if embJSON != "" && embJSON != "null" && embJSON != "{}" {
			if err := json.Unmarshal([]byte(embJSON), &emb); err != nil {
				continue
			}
		}
		if len(emb) == 0 {
			continue
		}

		score := cosineSim(targetEmb, emb) // from your similar.go
		if score <= 0 {
			continue
		}

		results = append(results, DBSimilarResult{
			AppID: appID,
			Name:  name,
			Score: score,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, targetEmb, fmt.Errorf("iterate candidates: %w", err)
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	if len(results) > topK {
		results = results[:topK]
	}

	return results, targetEmb, nil
}

func loadEmbeddingOnly(ctx context.Context, db *sql.DB, appID int64) (map[string]float64, error) {
	_, emb, err := loadEmbeddingByAppID(ctx, db, appID)
	return emb, err
}

// Cosine similarity between two sparse vectors represented as map[token]weight.
func cosineSim(a, b map[string]float64) float64 {
	// iterate over smaller map for speed
	if len(a) > len(b) {
		a, b = b, a
	}

	var dot, na, nb float64

	// dot + norm(a)
	for k, av := range a {
		dot += av * b[k]
		na += av * av
	}
	// norm(b)
	for _, bv := range b {
		nb += bv * bv
	}

	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// SharedTopTerms returns up to topN overlapping tokens ranked by contribution.
// Great for explaining *why* two games are similar.
func SharedTopTerms(a, b map[string]float64, topN int) []struct {
	Term  string
	Score float64
} {
	if topN <= 0 {
		return nil
	}

	// iterate smaller map
	if len(a) > len(b) {
		a, b = b, a
	}

	type termScore struct {
		Term  string
		Score float64
	}

	list := make([]termScore, 0, topN+16)
	for term, av := range a {
		if bv, ok := b[term]; ok {
			// contribution proxy: min weight
			s := av
			if bv < s {
				s = bv
			}
			if s > 0 {
				list = append(list, termScore{Term: term, Score: s})
			}
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Score > list[j].Score
	})

	if len(list) > topN {
		list = list[:topN]
	}

	out := make([]struct {
		Term  string
		Score float64
	}, len(list))

	for i := range list {
		out[i] = struct {
			Term  string
			Score float64
		}{Term: list[i].Term, Score: list[i].Score}
	}

	return out
}
