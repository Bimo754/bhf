package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Open initializes and configures a SQLite database handle with high-performance PRAGMAs.
func Open(dbPath string) (*sql.DB, error) {
	// Enable WAL mode, busy timeout for concurrency, and foreign keys
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database at %s: %w", dbPath, err)
	}

	// Single-writer friendly connection pool for SQLite
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
