package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	// adjust the import path accordingly
)

func main() {
	db, err := sql.Open("postgres", "postgresql://postgres:5274@localhost:5433/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Optional but good to verify connection early
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database connected successfully")

	games, skipped, err := GetAllGamesSkipBad(context.Background(), db)
	if err != nil {
		log.Fatalf("Failed to get games: %v", err)
	}

	fmt.Printf("Retrieved %d games from the database (skipped %d)\n", len(games), skipped)

	CUTOFF := 10
	fmt.Printf("Using first %d games for computation\n", CUTOFF)

	corpus := make([][]string, 0, len(games))
	for i := range games {
		tokens := tokenizeGame(&games[i])
		corpus = append(corpus, tokens)
	}
	fmt.Println("Tokenization complete")

	embeddings := make([]map[string]float64, 0, len(games))
	termCount := precomputeDocumentsContainingTerm(corpus[:CUTOFF])
	fmt.Println("Document count per term computed")
	idfmap := PrecmputeIDF(corpus[:CUTOFF], termCount)
	fmt.Println("IDF precomputation complete")
	for i := range games[:CUTOFF] {
		embedding := TFIDFEmbedding(corpus[i], idfmap)
		// fmt.Println("Computed embedding for game:", games[i].AppID, embedding)
		if err := SaveGameEmbedding(context.Background(), db, games[i].AppID, embedding); err != nil {
			log.Printf("Failed to save embedding for app_id=%d: %v", games[i].AppID, err)
		}
		embeddings = append(embeddings, embedding)
	}

	targetIdx := 0
	top := FindSimilarGames(games, embeddings, targetIdx, 10)

	fmt.Printf("Top similar games to %s:\n", games[targetIdx].Name)
	for _, r := range top {
		fmt.Printf("Score %.4f — %s (%d)\n", r.Score, r.Name, r.AppID)
	}

	// explain the top match
	if len(top) > 0 {
		best := top[0].Index
		terms := SharedTopTerms(embeddings[targetIdx], embeddings[best], 10)
		fmt.Println("Shared terms:", terms)
	}
}
