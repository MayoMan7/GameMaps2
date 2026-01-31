// user_profiles.go
// Minimal user system:
// 1) Create user
// 2) Add liked game (app_id) to users.games_liked
// 3) Recompute + store users.taste_embedding as the average of liked game embeddings
//
// Assumptions:
// - public.users exists with columns: id, name, games_liked (BIGINT[]), taste_embedding (JSONB)
// - public.steam_games has: app_id (BIGINT), tfidf_embedding (JSONB)
// - Your game embeddings are stored as JSONB maps: map[string]float64
//
// Drop this into your project. It only depends on stdlib + database/sql.

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

// -------------------------
// Types
// -------------------------

type User struct {
	ID             int64
	Name           string
	GamesLiked     []int64
	TasteEmbedding map[string]float64
}

// -------------------------
// Create user
// -------------------------

// CreateUser inserts a new user and returns their ID.
func CreateUser(ctx context.Context, db *sql.DB, name string) (int64, error) {
	if name == "" {
		return 0, errors.New("name cannot be empty")
	}

	const q = `
		INSERT INTO public.users (name)
		VALUES ($1)
		RETURNING id;
	`
	var id int64
	if err := db.QueryRowContext(ctx, q, name).Scan(&id); err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}
	return id, nil
}

// GetUserByID loads a user row.
func GetUserByID(ctx context.Context, db *sql.DB, userID int64) (*User, error) {
	const q = `
		SELECT id, name, games_liked, COALESCE(taste_embedding, '{}'::jsonb)::text
		FROM public.users
		WHERE id = $1
		LIMIT 1;
	`

	var u User
	var embJSON string
	if err := db.QueryRowContext(ctx, q, userID).Scan(&u.ID, &u.Name, pq.Array(&u.GamesLiked), &embJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user id=%d: %w", userID, err)
	}

	u.TasteEmbedding = make(map[string]float64)
	if embJSON != "" && embJSON != "null" && embJSON != "{}" {
		_ = json.Unmarshal([]byte(embJSON), &u.TasteEmbedding) // if it fails, keep empty
	}

	return &u, nil
}

// -------------------------
// Add game to liked list
// -------------------------

// AddLikedGame adds appID to games_liked (no duplicates).
func AddLikedGame(ctx context.Context, db *sql.DB, userID int64, appID int64) error {
	const q = `
		UPDATE public.users
		SET games_liked =
			CASE
				WHEN NOT ($2 = ANY(games_liked)) THEN array_append(games_liked, $2)
				ELSE games_liked
			END
		WHERE id = $1;
	`
	res, err := db.ExecContext(ctx, q, userID, appID)
	if err != nil {
		return fmt.Errorf("add liked game user_id=%d app_id=%d: %w", userID, appID, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user id=%d not found", userID)
	}
	return nil
}

// -------------------------
// Taste embedding computation
// -------------------------

// loadEmbeddingOnly loads steam_games.tfidf_embedding for one app_id.
func loadGameEmbeddingOnly(ctx context.Context, db *sql.DB, appID int64) (map[string]float64, error) {
	const q = `
		SELECT COALESCE(tfidf_embedding, '{}'::jsonb)::text
		FROM public.steam_games
		WHERE app_id = $1
		LIMIT 1;
	`

	var embJSON string
	if err := db.QueryRowContext(ctx, q, appID).Scan(&embJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("load game embedding app_id=%d: %w", appID, err)
	}

	emb := make(map[string]float64)
	if embJSON != "" && embJSON != "null" && embJSON != "{}" {
		if err := json.Unmarshal([]byte(embJSON), &emb); err != nil {
			return nil, fmt.Errorf("unmarshal game embedding app_id=%d: %w", appID, err)
		}
	}
	return emb, nil
}

// BuildUserEmbedding averages sparse embeddings.
func BuildUserEmbedding(embs []map[string]float64) map[string]float64 {
	user := make(map[string]float64)
	if len(embs) == 0 {
		return user
	}

	// sum
	for _, e := range embs {
		for term, w := range e {
			user[term] += w
		}
	}

	// average
	inv := 1.0 / float64(len(embs))
	for term, w := range user {
		user[term] = w * inv
	}

	return user
}

// RecomputeAndSaveTasteEmbedding:
// - loads user.games_liked
// - loads each liked game's tfidf_embedding
// - averages them into a taste embedding
// - saves into users.taste_embedding
//
// Returns the computed embedding (and the number of liked games used).
func RecomputeAndSaveTasteEmbedding(ctx context.Context, db *sql.DB, userID int64) (map[string]float64, int, error) {
	u, err := GetUserByID(ctx, db, userID)
	if err != nil {
		return nil, 0, err
	}
	if u == nil {
		return nil, 0, fmt.Errorf("user id=%d not found", userID)
	}
	if len(u.GamesLiked) == 0 {
		// save empty embedding
		if err := saveUserTasteEmbedding(ctx, db, userID, map[string]float64{}); err != nil {
			return nil, 0, err
		}
		return map[string]float64{}, 0, nil
	}

	embs := make([]map[string]float64, 0, len(u.GamesLiked))
	used := 0

	for _, appID := range u.GamesLiked {
		emb, err := loadGameEmbeddingOnly(ctx, db, appID)
		if err != nil {
			// skip bad embeddings, don't fail user
			continue
		}
		if len(emb) == 0 {
			// embedding missing/empty -> skip
			continue
		}
		embs = append(embs, emb)
		used++
	}

	taste := BuildUserEmbedding(embs)

	if err := saveUserTasteEmbedding(ctx, db, userID, taste); err != nil {
		return nil, used, err
	}

	return taste, used, nil
}

func saveUserTasteEmbedding(ctx context.Context, db *sql.DB, userID int64, emb map[string]float64) error {
	b, err := json.Marshal(emb)
	if err != nil {
		return fmt.Errorf("marshal taste embedding: %w", err)
	}

	const q = `
		UPDATE public.users
		SET taste_embedding = $2::jsonb
		WHERE id = $1;
	`
	res, err := db.ExecContext(ctx, q, userID, string(b))
	if err != nil {
		return fmt.Errorf("save taste embedding user_id=%d: %w", userID, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user id=%d not found", userID)
	}
	return nil
}
