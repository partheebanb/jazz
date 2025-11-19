package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	conn, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer conn.Close(context.Background())

	migrationsDir := "./database/migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatal("Failed to read migrations:", err)
	}

	var sqlFiles []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles)

	for _, file := range sqlFiles {
		log.Printf("Running migration: %s", file)

		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			log.Fatal("Failed to read file:", err)
		}

		_, err = conn.Exec(context.Background(), string(content))
		if err != nil {
			log.Fatalf("Failed to execute %s: %v", file, err)
		}

		log.Printf("✓ %s", file)
	}

	fmt.Println("\nAll migrations completed!")
}