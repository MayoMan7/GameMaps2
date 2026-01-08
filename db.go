package main

import (
	"context"
	"database/sql"
	"fmt"
)

type User struct {
	Name     string
	Likes    []string
	Dislikes []string
}

func AddUser(ctx context.Context, db *sql.DB, u User) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO users (name) VALUES ($1)`,
		u.Name,
	)
	fmt.Println(err)
	return err
}
