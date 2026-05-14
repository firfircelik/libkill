package db

import (
	"database/sql"
	"fmt"
	"strings"
)

func migrate(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS threats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			package TEXT NOT NULL,
			version TEXT NOT NULL DEFAULT '',
			ecosystem TEXT NOT NULL,
			feed TEXT NOT NULL,
			severity TEXT NOT NULL DEFAULT 'unknown',
			reason TEXT NOT NULL DEFAULT '',
			detected DATETIME NOT NULL,
			created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_threats_pkg
		 ON threats(package, version, ecosystem, feed)`,
		`CREATE TABLE IF NOT EXISTS scan_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ecosystem TEXT NOT NULL,
			found INTEGER NOT NULL DEFAULT 0,
			scanned_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`ALTER TABLE threats ADD COLUMN platform TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE threats ADD COLUMN cve TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE threats ADD COLUMN source_url TEXT NOT NULL DEFAULT ''`,
		`CREATE TABLE IF NOT EXISTS malware_families (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			type TEXT NOT NULL DEFAULT '',
			platform TEXT NOT NULL DEFAULT '',
			discovered TEXT NOT NULL DEFAULT '',
			attribution TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			cve TEXT NOT NULL DEFAULT '',
			source_url TEXT NOT NULL DEFAULT '',
			created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_malware_name
		 ON malware_families(name)`,
	}

	for i, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			if strings.Contains(err.Error(), "duplicate column name") {
				continue
			}
			return fmt.Errorf("migration %d: %w", i, err)
		}
	}

	return nil
}
