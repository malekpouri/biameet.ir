package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	return runMigrations()
}

func runMigrations() error {
	// Read the migration file
	// In a real app, we might use a migration tool like golang-migrate
	// For this task, we'll read the file directly.
	// Assuming running from backend/ or root, we need to find the file.
	// Let's try to locate it relative to the binary or current working directory.

	migrationPath := "db/migrations/001_init_schema.sql"
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		// Try looking one level up if we are in cmd/
		migrationPath = "../db/migrations/001_init_schema.sql"
	}

	content, err := os.ReadFile(migrationPath)
	if err != nil {
		// Fallback for docker container where path might be absolute or different
		migrationPath = "/app/backend/db/migrations/001_init_schema.sql"
		content, err = os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("could not read migration file: %v", err)
		}
	}

	_, err = DB.Exec(string(content))
	if err != nil {
		// Ignore "table already exists" errors for simplicity in this basic setup
		// or better, check if tables exist.
		// For now, let's just print it and assume it might be okay if it's just re-running.
		// But sqlite doesn't have "IF NOT EXISTS" in the CREATE TABLE in my schema?
		// Wait, I should update the schema to use IF NOT EXISTS to be safe.
		return fmt.Errorf("migration failed: %v", err)
	}

	return nil
}
