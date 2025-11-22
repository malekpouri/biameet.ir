package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	return runMigrations()
}

func runMigrations() error {
	// Read the migration directory
	migrationDir := "db/migrations"
	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		// Try looking one level up if we are in cmd/ (local dev)
		migrationDir = "../db/migrations"
	}

	files, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("could not read migration directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		// Simple check for .sql extension
		if len(file.Name()) < 4 || file.Name()[len(file.Name())-4:] != ".sql" {
			continue
		}

		migrationPath := fmt.Sprintf("%s/%s", migrationDir, file.Name())
		content, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("could not read migration file from %s: %v", migrationPath, err)
		}

		_, err = DB.Exec(string(content))
		if err != nil {
			// Ignore errors for now as we might be re-running migrations
			// In a real app, we would track applied migrations
			fmt.Printf("Migration %s warning: %v\n", file.Name(), err)
		}
	}

	return nil
}
