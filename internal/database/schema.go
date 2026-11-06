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

		`CREATE TABLE IF NOT EXISTS endpoints (
			id TEXT PRIMARY KEY,
			target_id TEXT NOT NULL,
			scan_id TEXT DEFAULT '',
			host TEXT NOT NULL,
			url TEXT NOT NULL,
			path TEXT NOT NULL,
			method TEXT DEFAULT 'GET',
			status_code INTEGER DEFAULT 0,
			content_length INTEGER DEFAULT 0,
			title TEXT DEFAULT '',
			source_tool TEXT DEFAULT '',
			created_at DATETIME NOT NULL,
			FOREIGN KEY(target_id) REFERENCES targets(id) ON DELETE CASCADE,
			UNIQUE(target_id, url)
		);`,

		`CREATE TABLE IF NOT EXISTS traffic_records (
			id TEXT PRIMARY KEY,
			target_id TEXT NOT NULL,
			scan_id TEXT DEFAULT '',
			method TEXT NOT NULL,
			url TEXT NOT NULL,
			path TEXT NOT NULL,
			host TEXT NOT NULL,
			request_headers TEXT DEFAULT '{}',
			request_body TEXT DEFAULT '',
			response_status_code INTEGER DEFAULT 0,
			response_headers TEXT DEFAULT '{}',
			response_body TEXT DEFAULT '',
			content_type TEXT DEFAULT '',
			content_length INTEGER DEFAULT 0,
			duration_ms INTEGER DEFAULT 0,
			category TEXT DEFAULT 'general',
			auth_context TEXT DEFAULT '',
			parameters TEXT DEFAULT '[]',
			created_at DATETIME NOT NULL,
			FOREIGN KEY(target_id) REFERENCES targets(id) ON DELETE CASCADE
		);`,

		`CREATE TABLE IF NOT EXISTS hosts (
			id TEXT PRIMARY KEY,
			target_id TEXT NOT NULL,
			ip TEXT NOT NULL,
			hostname TEXT DEFAULT '',
			os TEXT DEFAULT '',
			mac TEXT DEFAULT '',
			status TEXT DEFAULT 'up',
			created_at DATETIME NOT NULL,
			FOREIGN KEY(target_id) REFERENCES targets(id) ON DELETE CASCADE,
			UNIQUE(target_id, ip)
		);`,

		`CREATE TABLE IF NOT EXISTS services (
			id TEXT PRIMARY KEY,
			host_id TEXT NOT NULL,
			target_id TEXT NOT NULL,
			port INTEGER NOT NULL,
			protocol TEXT DEFAULT 'tcp',
			name TEXT DEFAULT '',
			product TEXT DEFAULT '',
			version TEXT DEFAULT '',
			banner TEXT DEFAULT '',
			status TEXT DEFAULT 'open',
			created_at DATETIME NOT NULL,
			FOREIGN KEY(target_id) REFERENCES targets(id) ON DELETE CASCADE,
			FOREIGN KEY(host_id) REFERENCES hosts(id) ON DELETE CASCADE,
			UNIQUE(host_id, port, protocol)
		);`,

		`CREATE INDEX IF NOT EXISTS idx_scans_target_id ON scans(target_id);`,
		`CREATE INDEX IF NOT EXISTS idx_assets_target_id ON assets(target_id);`,
		`CREATE INDEX IF NOT EXISTS idx_assets_host ON assets(host);`,
		`CREATE INDEX IF NOT EXISTS idx_logs_scan_id ON logs(scan_id);`,
		`CREATE INDEX IF NOT EXISTS idx_findings_target_id ON findings(target_id);`,
		`CREATE INDEX IF NOT EXISTS idx_findings_severity ON findings(severity);`,
		`CREATE INDEX IF NOT EXISTS idx_endpoints_target_id ON endpoints(target_id);`,
		`CREATE INDEX IF NOT EXISTS idx_endpoints_host ON endpoints(host);`,
		`CREATE INDEX IF NOT EXISTS idx_endpoints_url ON endpoints(url);`,
		`CREATE INDEX IF NOT EXISTS idx_traffic_target_id ON traffic_records(target_id);`,
		`CREATE INDEX IF NOT EXISTS idx_traffic_host ON traffic_records(host);`,
		`CREATE INDEX IF NOT EXISTS idx_traffic_category ON traffic_records(category);`,
		`CREATE INDEX IF NOT EXISTS idx_traffic_method ON traffic_records(method);`,
		`CREATE INDEX IF NOT EXISTS idx_traffic_status ON traffic_records(response_status_code);`,
		`CREATE INDEX IF NOT EXISTS idx_hosts_target_id ON hosts(target_id);`,
		`CREATE INDEX IF NOT EXISTS idx_hosts_ip ON hosts(ip);`,
		`CREATE INDEX IF NOT EXISTS idx_services_target_id ON services(target_id);`,
		`CREATE INDEX IF NOT EXISTS idx_services_host_id ON services(host_id);`,
		`CREATE INDEX IF NOT EXISTS idx_services_port ON services(port);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("migration query failed: %w (query: %s)", err, query)
		}
	}

	// Ensure priority_score column exists for existing database instances
	_, _ = db.Exec(`ALTER TABLE assets ADD COLUMN priority_score INTEGER DEFAULT 0;`)

	// Data migration: Repair any assets where host was mistakenly set to an internal host UUID
	_, _ = db.Exec(`
		DELETE FROM assets 
		WHERE LENGTH(host) = 36 AND host LIKE '%-%-%-%-%'
		  AND EXISTS (
		      SELECT 1 FROM hosts h 
		      JOIN assets existing ON existing.target_id = assets.target_id 
		                          AND existing.port = assets.port 
		                          AND existing.host = COALESCE(NULLIF(h.hostname, ''), h.ip)
		      WHERE h.id = assets.host
		  );
	`)

	_, _ = db.Exec(`
		UPDATE assets
		SET 
			ip = (SELECT hosts.ip FROM hosts WHERE hosts.id = assets.host),
			url = REPLACE(assets.url, assets.host, (SELECT COALESCE(NULLIF(hosts.hostname, ''), hosts.ip) FROM hosts WHERE hosts.id = assets.host)),
			host = (SELECT COALESCE(NULLIF(hosts.hostname, ''), hosts.ip) FROM hosts WHERE hosts.id = assets.host)
		WHERE LENGTH(assets.host) = 36 AND assets.host LIKE '%-%-%-%-%'
		  AND EXISTS (SELECT 1 FROM hosts WHERE hosts.id = assets.host);
	`)

	_, _ = db.Exec(`DELETE FROM assets WHERE LENGTH(host) = 36 AND host LIKE '%-%-%-%-%';`)

	return nil
}
