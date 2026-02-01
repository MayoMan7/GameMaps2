package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gogamemaps/internal/db"
	"gogamemaps/internal/similar"

	_ "github.com/lib/pq"
)

func main() {
	database, err := sql.Open("postgres", "postgresql://postgres:5274@localhost:5433/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	database.SetMaxOpenConns(8)
	database.SetMaxIdleConns(8)
	database.SetConnMaxLifetime(30 * time.Minute)

	if err := database.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✅ Connected to Postgres")

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

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

		name, err := db.GetGameNameByAppID(ctx, database, appID)
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

		top, _, err := similar.FindSimilarGamesFromDB(ctx, database, appID, 15, 50000)
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
