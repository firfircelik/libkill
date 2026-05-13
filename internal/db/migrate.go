package db

import (
	"database/sql"
	"fmt"
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
	}

	for i, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration %d: %w", i, err)
		}
	}

	return nil
}
