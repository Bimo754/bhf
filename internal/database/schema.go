package database

import (
	"database/sql"
	"fmt"
)

// MigrateSchema creates all foundational tables and indexes if they do not exist.
func MigrateSchema(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS targets (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			domain TEXT NOT NULL UNIQUE,
			notes TEXT DEFAULT '',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS scans (
			id TEXT PRIMARY KEY,
			target_id TEXT NOT NULL,
			tool_name TEXT NOT NULL,
			status TEXT NOT NULL,
			command TEXT NOT NULL,
			arguments TEXT DEFAULT '[]',
			started_at DATETIME NOT NULL,
			completed_at DATETIME,
			error_msg TEXT DEFAULT '',
			FOREIGN KEY(target_id) REFERENCES targets(id) ON DELETE CASCADE
		);`,

		`CREATE TABLE IF NOT EXISTS assets (
			id TEXT PRIMARY KEY,
			target_id TEXT NOT NULL,
			scan_id TEXT NOT NULL,
			host TEXT NOT NULL,
			url TEXT DEFAULT '',
			ip TEXT DEFAULT '',
			port INTEGER DEFAULT 0,
			status_code INTEGER DEFAULT 0,
			title TEXT DEFAULT '',
			technologies TEXT DEFAULT '[]',
			priority_score INTEGER DEFAULT 0,
			created_at DATETIME NOT NULL,
			FOREIGN KEY(target_id) REFERENCES targets(id) ON DELETE CASCADE,
			FOREIGN KEY(scan_id) REFERENCES scans(id) ON DELETE CASCADE,
			UNIQUE(target_id, host, port)
		);`,

		`CREATE TABLE IF NOT EXISTS logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			scan_id TEXT NOT NULL,
			stream TEXT NOT NULL,
			line TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			FOREIGN KEY(scan_id) REFERENCES scans(id) ON DELETE CASCADE
		);`,

		`CREATE TABLE IF NOT EXISTS findings (
			id TEXT PRIMARY KEY,
			target_id TEXT NOT NULL,
			scan_id TEXT NOT NULL,
			tool_name TEXT NOT NULL,
			severity TEXT NOT NULL,
			name TEXT NOT NULL,
			matched_at TEXT NOT NULL,
			description TEXT DEFAULT '',
			extracted_results TEXT DEFAULT '[]',
			curl_command TEXT DEFAULT '',
			created_at DATETIME NOT NULL,
			FOREIGN KEY(target_id) REFERENCES targets(id) ON DELETE CASCADE,
			FOREIGN KEY(scan_id) REFERENCES scans(id) ON DELETE CASCADE
		);`,

		`CREATE INDEX IF NOT EXISTS idx_scans_target_id ON scans(target_id);`,
		`CREATE INDEX IF NOT EXISTS idx_assets_target_id ON assets(target_id);`,
		`CREATE INDEX IF NOT EXISTS idx_assets_host ON assets(host);`,
		`CREATE INDEX IF NOT EXISTS idx_logs_scan_id ON logs(scan_id);`,
		`CREATE INDEX IF NOT EXISTS idx_findings_target_id ON findings(target_id);`,
		`CREATE INDEX IF NOT EXISTS idx_findings_severity ON findings(severity);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("migration query failed: %w (query: %s)", err, query)
		}
	}

	// Ensure priority_score column exists for existing database instances
	_, _ = db.Exec(`ALTER TABLE assets ADD COLUMN priority_score INTEGER DEFAULT 0;`)

	return nil
}
