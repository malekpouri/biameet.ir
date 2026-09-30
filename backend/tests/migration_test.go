package tests

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"biameet.ir/db"
	_ "modernc.org/sqlite"
)

// A database created by the old runner (all four migrations applied, no
// schema_migrations table) must be adopted without re-running old ALTERs,
// keep its data, and receive the new migrations.
func TestLegacyDatabaseIsAdopted(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	legacy, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"001_init_schema.sql", "002_add_dynamic_fields.sql", "003_add_participants.sql", "004_secure_timeslots.sql"} {
		content, err := os.ReadFile(filepath.Join("..", "db", "migrations", f))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := legacy.Exec(string(content)); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
	}
	if _, err := legacy.Exec(`INSERT INTO sessions (id, title, creator_name, created_at_utc) VALUES ('abcde', 'old', 'x', '2024-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	legacy.Close()

	if err := db.InitDB(path); err != nil {
		t.Fatalf("InitDB on legacy database: %v", err)
	}
	defer db.Close()

	var applied, sessions, idx int
	db.DB.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&applied)
	db.DB.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&sessions)
	db.DB.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE name = 'idx_timeslots_session_start'`).Scan(&idx)
	if applied != 5 || sessions != 1 || idx != 1 {
		t.Fatalf("applied=%d sessions=%d index=%d", applied, sessions, idx)
	}

	// Restarting is a no-op.
	if err := db.InitDB(path); err != nil {
		t.Fatalf("second InitDB: %v", err)
	}

	var fk, wal string
	db.DB.QueryRow(`PRAGMA foreign_keys`).Scan(&fk)
	db.DB.QueryRow(`PRAGMA journal_mode`).Scan(&wal)
	if fk != "1" || wal != "wal" {
		t.Fatalf("pragmas not applied: foreign_keys=%s journal_mode=%s", fk, wal)
	}
}
