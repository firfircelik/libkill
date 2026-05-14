package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

type ThreatEntry struct {
	ID        int64     `json:"id"`
	Package   string    `json:"package"`
	Version   string    `json:"version"`
	Ecosystem string    `json:"ecosystem"`
	Feed      string    `json:"feed"`
	Severity  string    `json:"severity"`
	Reason    string    `json:"reason"`
	Detected  time.Time `json:"detected"`
	Created   time.Time `json:"created"`
	Platform  string    `json:"platform,omitempty"`
	CVE       string    `json:"cve,omitempty"`
	SourceURL string    `json:"source_url,omitempty"`
}

type MalwareFamily struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Platform    string `json:"platform"`
	Discovered  string `json:"discovered"`
	Attribution string `json:"attribution"`
	Description string `json:"description"`
	CVE         string `json:"cve,omitempty"`
	SourceURL   string `json:"source_url,omitempty"`
}

type ScanResult struct {
	Package   string `json:"package"`
	Version   string `json:"version"`
	Ecosystem string `json:"ecosystem"`
	Location  string `json:"location"`
	Threat    ThreatEntry
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) UpsertThreats(entries []ThreatEntry) (int, error) {
	inserted := 0
	query := `INSERT OR IGNORE INTO threats (package, version, ecosystem, feed, severity, reason, detected, platform, cve, source_url)
	           VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, e := range entries {
		res, err := stmt.Exec(e.Package, e.Version, e.Ecosystem, e.Feed, e.Severity, e.Reason, e.Detected.UTC().Format("2006-01-02 15:04:05"), e.Platform, e.CVE, e.SourceURL)
		if err != nil {
			return inserted, fmt.Errorf("inserting %s@%s: %w", e.Package, e.Version, err)
		}
		n, _ := res.RowsAffected()
		inserted += int(n)
	}

	if err := tx.Commit(); err != nil {
		return inserted, fmt.Errorf("commit: %w", err)
	}

	return inserted, nil
}

func (s *Store) Search(pkg, ecosystem string) ([]ThreatEntry, error) {
	query := `SELECT id, package, version, ecosystem, feed, severity, reason, detected, created, platform, cve, source_url
	           FROM threats WHERE package = ? AND ecosystem = ?`

	rows, err := s.db.Query(query, pkg, ecosystem)
	if err != nil {
		return nil, fmt.Errorf("query threats: %w", err)
	}
	defer rows.Close()

	var results []ThreatEntry
	for rows.Next() {
		var t ThreatEntry
		var detectedStr, createdStr string
		if err := rows.Scan(&t.ID, &t.Package, &t.Version, &t.Ecosystem, &t.Feed, &t.Severity, &t.Reason, &detectedStr, &createdStr, &t.Platform, &t.CVE, &t.SourceURL); err != nil {
			return results, fmt.Errorf("scan: %w", err)
		}
		t.Detected, _ = time.Parse("2006-01-02 15:04:05", detectedStr)
		t.Created, _ = time.Parse("2006-01-02 15:04:05", createdStr)
		results = append(results, t)
	}

	return results, rows.Err()
}

func (s *Store) AllThreats(ecosystem string) ([]ThreatEntry, error) {
	query := `SELECT id, package, version, ecosystem, feed, severity, reason, detected, created, platform, cve, source_url
	           FROM threats`
	args := []interface{}{}
	if ecosystem != "" {
		query += " WHERE ecosystem = ?"
		args = append(args, ecosystem)
	}
	query += " ORDER BY detected DESC LIMIT 5000"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query threats: %w", err)
	}
	defer rows.Close()

	var results []ThreatEntry
	for rows.Next() {
		var t ThreatEntry
		var detectedStr, createdStr string
		if err := rows.Scan(&t.ID, &t.Package, &t.Version, &t.Ecosystem, &t.Feed, &t.Severity, &t.Reason, &detectedStr, &createdStr, &t.Platform, &t.CVE, &t.SourceURL); err != nil {
			return results, fmt.Errorf("scan: %w", err)
		}
		t.Detected, _ = time.Parse("2006-01-02 15:04:05", detectedStr)
		t.Created, _ = time.Parse("2006-01-02 15:04:05", createdStr)
		results = append(results, t)
	}

	return results, rows.Err()
}

func (s *Store) CountThreats() (int, error) {
	var n int
	err := s.db.QueryRow("SELECT COUNT(*) FROM threats").Scan(&n)
	return n, err
}

func (s *Store) LastSync(feed string) (time.Time, error) {
	var raw sql.NullString
	err := s.db.QueryRow("SELECT MAX(detected) FROM threats WHERE feed = ?", feed).Scan(&raw)
	if err != nil || !raw.Valid {
		return time.Time{}, nil
	}
	t, err := time.Parse("2006-01-02 15:04:05", raw.String)
	if err != nil {
		return time.Time{}, nil
	}
	return t, nil
}

func (s *Store) UpsertMalwareFamilies(families []MalwareFamily) (int, error) {
	inserted := 0
	query := `INSERT OR IGNORE INTO malware_families (name, type, platform, discovered, attribution, description, cve, source_url)
	           VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, f := range families {
		res, err := stmt.Exec(f.Name, f.Type, f.Platform, f.Discovered, f.Attribution, f.Description, f.CVE, f.SourceURL)
		if err != nil {
			return inserted, fmt.Errorf("inserting %s: %w", f.Name, err)
		}
		n, _ := res.RowsAffected()
		inserted += int(n)
	}

	if err := tx.Commit(); err != nil {
		return inserted, fmt.Errorf("commit: %w", err)
	}

	return inserted, nil
}

func (s *Store) AllMalwareFamilies() ([]MalwareFamily, error) {
	query := `SELECT id, name, type, platform, discovered, attribution, description, cve, source_url
	           FROM malware_families ORDER BY name`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query malware: %w", err)
	}
	defer rows.Close()

	var results []MalwareFamily
	for rows.Next() {
		var m MalwareFamily
		if err := rows.Scan(&m.ID, &m.Name, &m.Type, &m.Platform, &m.Discovered, &m.Attribution, &m.Description, &m.CVE, &m.SourceURL); err != nil {
			return results, fmt.Errorf("scan malware: %w", err)
		}
		results = append(results, m)
	}

	return results, rows.Err()
}

func (s *Store) SearchMalware(name string) (*MalwareFamily, error) {
	query := `SELECT id, name, type, platform, discovered, attribution, description, cve, source_url
	           FROM malware_families WHERE name = ?`

	var m MalwareFamily
	err := s.db.QueryRow(query, name).Scan(&m.ID, &m.Name, &m.Type, &m.Platform, &m.Discovered, &m.Attribution, &m.Description, &m.CVE, &m.SourceURL)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query malware: %w", err)
	}

	return &m, nil
}

func (s *Store) CountMalwareFamilies() (int, error) {
	var n int
	err := s.db.QueryRow("SELECT COUNT(*) FROM malware_families").Scan(&n)
	return n, err
}
