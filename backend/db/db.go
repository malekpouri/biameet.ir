package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	// Ensure the directory exists
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatal(err)
	}

	DB, err = sql.Open("sqlite3", "./data/biameet.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Database connection established")
	runMigrations()
}

func runMigrations() {
	migrationFile := "db/migrations/001_init.sql"
	content, err := os.ReadFile(migrationFile)
	if err != nil {
		// Try looking relative to where the binary might be running if not root
		migrationFile = "../db/migrations/001_init.sql"
		content, err = os.ReadFile(migrationFile)
		if err != nil {
			log.Fatalf("Failed to read migration file: %v", err)
		}
	}

	_, err = DB.Exec(string(content))
	if err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	log.Println("Migrations executed successfully")
}
