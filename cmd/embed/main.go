package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"gogamemaps/internal/db"
	"gogamemaps/internal/tfidf"

	_ "github.com/lib/pq"
)

func main() {
	database, err := sql.Open("postgres", "postgresql://postgres:5274@localhost:5433/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	database.SetMaxOpenConns(12)
	database.SetMaxIdleConns(12)
	database.SetConnMaxLifetime(30 * time.Minute)

	if err := database.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Database connected successfully")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	games, skipped, err := db.GetAllGamesSkipBad(ctx, database)
	if err != nil {
		log.Fatalf("Failed to get games: %v", err)
	}
	fmt.Printf("Retrieved %d games (skipped %d)\n", len(games), skipped)

	CUTOFF := 200000
	if CUTOFF > len(games) {
		CUTOFF = len(games)
	}
	fmt.Printf("Using first %d games\n", CUTOFF)

	corpus := make([]map[string]float64, CUTOFF)
	for i := 0; i < CUTOFF; i++ {
		corpus[i] = tfidf.TokenizeGameWeighted(&games[i])
	}
	fmt.Println("Tokenization complete")

	df := tfidf.PrecomputeDocumentsContainingTermWeighted(corpus)
	fmt.Println("Document frequency computed. Vocab size:", len(df))

	idfmap := tfidf.PrecomputeIDFWeighted(corpus, df)
	fmt.Println("IDF precomputation complete")

	numWorkers := min(runtime.NumCPU(), 12)
	fmt.Println("Workers:", numWorkers)

	jobs := make(chan int, 256)
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for w := 0; w < numWorkers; w++ {
		workerID := w
		go func() {
			defer wg.Done()
			for i := range jobs {
				emb := tfidf.TFIDFEmbeddingWeighted(corpus[i], idfmap)
				if err := db.SaveGameEmbedding(ctx, database, games[i].AppID, emb); err != nil {
					log.Printf("[worker %d] save failed app_id=%d: %v", workerID, games[i].AppID, err)
				}
			}
		}()
	}

	for i := 0; i < CUTOFF; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	fmt.Println("Saved embeddings to DB")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
