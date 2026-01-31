package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// Run the whole TF-IDF -> save embeddings -> query similar from DB pipeline
func main1() {
	// ---- DB connect ----
	db, err := sql.Open("postgres", "postgresql://postgres:5274@localhost:5433/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Optional pool tuning (good for concurrent updates)
	db.SetMaxOpenConns(12)
	db.SetMaxIdleConns(12)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Database connected successfully")

	// Single context used everywhere
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	// ---- Load games ----
	games, skipped, err := GetAllGamesSkipBad(ctx, db)
	if err != nil {
		log.Fatalf("Failed to get games: %v", err)
	}
	fmt.Printf("Retrieved %d games (skipped %d)\n", len(games), skipped)

	// ---- Choose cutoff ----
	CUTOFF := 200000
	if CUTOFF > len(games) {
		CUTOFF = len(games)
	}
	fmt.Printf("Using first %d games\n", CUTOFF)

	// ---- Tokenize only the cutoff set ----
	corpus := make([][]string, CUTOFF)
	for i := 0; i < CUTOFF; i++ {
		corpus[i] = tokenizeGame(&games[i])
	}
	fmt.Println("Tokenization complete")

	// ---- Build DF + IDF from cutoff corpus ----
	// df[token] = #docs containing token
	df := precomputeDocumentsContainingTerm(corpus)
	fmt.Println("Document frequency computed. Vocab size:", len(df))

	// idfmap[token] = IDF score
	idfmap := PrecmputeIDF(corpus, df)
	fmt.Println("IDF precomputation complete")

	// ---- Parallel compute embeddings + save to DB ----
	// CPU bound compute + IO bound DB update:
	// cap workers to something reasonable for DB, e.g. min(NumCPU, 8)
	numWorkers := min(runtime.NumCPU(), 12)
	fmt.Println("Workers:", numWorkers)

	jobs := make(chan int, 256)

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for w := 0; w < numWorkers; w++ {
		workerID := w // avoid closure capture bug
		go func() {
			defer wg.Done()
			for i := range jobs {
				emb := TFIDFEmbedding(corpus[i], idfmap)

				if err := SaveGameEmbedding(ctx, db, games[i].AppID, emb); err != nil {
					log.Printf("[worker %d] save failed app_id=%d: %v", workerID, games[i].AppID, err)
					continue
				}
				// Optional progress log:
				// log.Printf("[worker %d] saved app_id=%d", workerID, games[i].AppID)
			}
		}()
	}

	// Feed jobs
	for i := 0; i < CUTOFF; i++ {
		jobs <- i
	}
	close(jobs)

	// Wait for workers to finish
	wg.Wait()
	fmt.Println("Saved embeddings to DB")
}

// helper
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetGameNameByAppID returns the game's name (or empty string if not found).
func GetGameNameByAppID(ctx context.Context, db *sql.DB, appID int64) (string, error) {
	const q = `
		SELECT name
		FROM public.steam_games
		WHERE app_id = $1
		LIMIT 1;
	`
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

func main() {
	// ---- DB connect ----
	db, err := sql.Open("postgres", "postgresql://postgres:5274@localhost:5433/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✅ Connected to Postgres")

	// ---- CLI loop ----
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\nGameMaps CLI")
	fmt.Println("Enter an app_id to get top 15 similar games.")
	fmt.Println("Commands: 'q' to quit, 'help' for info.\n")

	for {
		fmt.Print("app_id> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("\nGoodbye.")
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		switch strings.ToLower(line) {
		case "q", "quit", "exit":
			fmt.Println("Goodbye.")
			return
		case "help":
			fmt.Println("Type a Steam app_id (e.g. 730).")
			fmt.Println("I will look up its stored TF-IDF embedding in the DB and print the top 15 similar games.")
			fmt.Println("Commands: q/quit/exit")
			continue
		}

		appID, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			fmt.Println("❌ Not a valid integer app_id. Try again.")
			continue
		}

		// short per-request context
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

		// fetch target name
		name, err := GetGameNameByAppID(ctx, db, appID)
		if err != nil {
			cancel()
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}
		if name == "" {
			cancel()
			fmt.Printf("❌ app_id %d not found in DB.\n\n", appID)
			continue
		}

		// get similar
		top, _, err := FindSimilarGamesFromDB(ctx, db, appID, 15, 50000)
		cancel()

		if err != nil {
			fmt.Printf("❌ Error getting recommendations for %d (%s): %v\n", appID, name, err)
			fmt.Println("Tip: make sure this app_id has a non-empty tfidf_embedding in the DB.")
			continue
		}

		fmt.Printf("\n🎮 Recommendations for %d — %s\n", appID, name)
		fmt.Println(strings.Repeat("-", 60))

		if len(top) == 0 {
			fmt.Println("No similar games found (maybe embedding missing or similarities were 0).")
			fmt.Println()
			continue
		}

		for i, r := range top {
			fmt.Printf("%2d) %.4f — %s (%d)\n", i+1, r.Score, r.Name, r.AppID)
		}
		fmt.Println()
	}
}
