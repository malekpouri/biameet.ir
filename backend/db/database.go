package db

import (
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

//go:embed migrations/*.sql
var migrationFS embed.FS

// Connection pragmas, applied to every pooled connection:
//   - foreign_keys: SQLite ignores FOREIGN KEY / ON DELETE CASCADE without it.
//   - WAL + synchronous=NORMAL: readers never block the writer, and far fewer fsyncs.
//   - busy_timeout: wait for the write lock instead of failing with SQLITE_BUSY.
//   - _txlock=immediate: transactions take the write lock at BEGIN, so a
//     read-then-write transaction can't fail mid-way on lock upgrade.
const dsnParams = "?_pragma=foreign_keys(1)" +
	"&_pragma=journal_mode(WAL)" +
	"&_pragma=synchronous(NORMAL)" +
	"&_pragma=busy_timeout(5000)" +
	"&_pragma=temp_store(memory)" +
	"&_txlock=immediate"

func InitDB(dbPath string) error {
	if DB != nil {
		DB.Close()
	}

	var err error
	DB, err = sql.Open("sqlite", "file:"+dbPath+dsnParams)
	if err != nil {
		return err
	}
	// SQLite has a single writer; a small pool is enough and keeps
	// per-connection page caches from piling up. Idle connections are released.
	DB.SetMaxOpenConns(4)
	DB.SetMaxIdleConns(1)
	DB.SetConnMaxIdleTime(2 * time.Minute)

	if err = DB.Ping(); err != nil {
		return err
	}

	return runMigrations()
}

func Close() error {
	if DB == nil {
		return nil
	}
	// Fold the WAL back into the main file so backups of biameet.db alone are complete.
	DB.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	return DB.Close()
}

// legacyProbes detect migrations that were applied before schema_migrations
// existed (the old runner re-ran every file on each start and ignored errors).
// Each probe checks the last change its migration makes.
var legacyProbes = map[string]string{
	"001_init_schema.sql":        `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='votes'`,
	"002_add_dynamic_fields.sql": `SELECT COUNT(*) FROM pragma_table_info('sessions') WHERE name='dynamic_config'`,
	"003_add_participants.sql":   `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='participants'`,
	"004_secure_timeslots.sql":   `SELECT COUNT(*) FROM pragma_table_info('timeslots') WHERE name='password_hash'`,
}

func runMigrations() error {
	if _, err := DB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at_utc TEXT NOT NULL
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied := map[string]bool{}
	rows, err := DB.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return err
		}
		applied[v] = true
	}
	rows.Close()

	files, err := fs.Glob(migrationFS, "migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)

	firstTracked := len(applied) == 0
	now := time.Now().UTC().Format(time.RFC3339)

	for _, path := range files {
		name := strings.TrimPrefix(path, "migrations/")
		if applied[name] {
			continue
		}

		if probe, ok := legacyProbes[name]; ok && firstTracked {
			var n int
			if err := DB.QueryRow(probe).Scan(&n); err != nil {
				return fmt.Errorf("probe %s: %w", name, err)
			}
			if n > 0 {
				if _, err := DB.Exec(`INSERT INTO schema_migrations (version, applied_at_utc) VALUES (?, ?)`, name, now); err != nil {
					return err
				}
				log.Printf("migration %s: already present, recorded as applied", name)
				continue
			}
		}

		content, err := migrationFS.ReadFile(path)
		if err != nil {
			return err
		}
		tx, err := DB.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version, applied_at_utc) VALUES (?, ?)`, name, now); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		log.Printf("migration %s: applied", name)
	}
	return nil
}

// Setting returns a value from app_settings, creating it with create() on first use.
func Setting(key string, create func() string) (string, error) {
	var v string
	err := DB.QueryRow(`SELECT value FROM app_settings WHERE key = ?`, key).Scan(&v)
	if err == nil {
		return v, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}
	v = create()
	if _, err := DB.Exec(`INSERT OR IGNORE INTO app_settings (key, value) VALUES (?, ?)`, key, v); err != nil {
		return "", err
	}
	// Re-read in case another process won the insert race.
	err = DB.QueryRow(`SELECT value FROM app_settings WHERE key = ?`, key).Scan(&v)
	return v, err
}

// RandomHex returns n random bytes, hex-encoded.
func RandomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
