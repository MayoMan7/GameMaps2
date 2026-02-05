package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"gogamemaps/internal/models"

	"github.com/lib/pq"
)

type AuthUser struct {
	User         models.User
	PasswordHash string
}

// CreateUserWithAuth inserts a new user with email + password hash.
func CreateUserWithAuth(ctx context.Context, db *sql.DB, name, email, passwordHash string) (int64, error) {
	const q = `
		INSERT INTO public.users (name, email, password_hash, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id;
	`
	var id int64
	if err := db.QueryRowContext(ctx, q, name, email, passwordHash).Scan(&id); err != nil {
		return 0, fmt.Errorf("create user with auth: %w", err)
	}
	return id, nil
}

// GetUserAuthByEmail loads a user plus password hash for login.
func GetUserAuthByEmail(ctx context.Context, db *sql.DB, email string) (*AuthUser, error) {
	const q = `
		SELECT id, name, COALESCE(email, ''), COALESCE(password_hash, ''),
			games_liked, COALESCE(taste_embedding, '{}'::jsonb)::text
		FROM public.users
		WHERE LOWER(email) = LOWER($1)
		LIMIT 1;
	`
	var u models.User
	var embJSON string
	var passwordHash string
	if err := db.QueryRowContext(ctx, q, email).Scan(
		&u.ID, &u.Name, &u.Email, &passwordHash, pq.Array(&u.GamesLiked), &embJSON,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	u.TasteEmbedding = make(map[string]float64)
	if embJSON != "" && embJSON != "null" && embJSON != "{}" {
		_ = json.Unmarshal([]byte(embJSON), &u.TasteEmbedding)
	}
	return &AuthUser{User: u, PasswordHash: passwordHash}, nil
}

// CreateSession stores a new session token.
func CreateSession(ctx context.Context, db *sql.DB, userID int64, token string, expiresAt time.Time) error {
	const q = `
		INSERT INTO public.user_sessions (token, user_id, created_at, expires_at)
		VALUES ($1, $2, NOW(), $3);
	`
	_, err := db.ExecContext(ctx, q, token, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// DeleteSession removes a session token.
func DeleteSession(ctx context.Context, db *sql.DB, token string) error {
	const q = `DELETE FROM public.user_sessions WHERE token = $1;`
	_, err := db.ExecContext(ctx, q, token)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// GetUserBySession loads a user from a session token.
func GetUserBySession(ctx context.Context, db *sql.DB, token string) (*models.User, error) {
	const q = `
		SELECT u.id, u.name, COALESCE(u.email, ''), u.games_liked, COALESCE(u.taste_embedding, '{}'::jsonb)::text
		FROM public.user_sessions s
		JOIN public.users u ON u.id = s.user_id
		WHERE s.token = $1 AND s.expires_at > NOW()
		LIMIT 1;
	`
	var u models.User
	var embJSON string
	if err := db.QueryRowContext(ctx, q, token).Scan(
		&u.ID, &u.Name, &u.Email, pq.Array(&u.GamesLiked), &embJSON,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by session: %w", err)
	}
	u.TasteEmbedding = make(map[string]float64)
	if embJSON != "" && embJSON != "null" && embJSON != "{}" {
		_ = json.Unmarshal([]byte(embJSON), &u.TasteEmbedding)
	}
	return &u, nil
}
