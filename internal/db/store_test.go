package db

import (
	"testing"
	"time"
)

func TestNewAndClose(t *testing.T) {
	store, err := New(":memory:")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer store.Close()
}

func TestUpsertAndSearch(t *testing.T) {
	store, _ := New(":memory:")
	defer store.Close()

	entries := []ThreatEntry{
		{
			Package:   "@tanstack/react-router",
			Version:   "1.120.0",
			Ecosystem: "npm",
			Feed:      "socket",
			Severity:  "critical",
			Reason:    "Mini Shai-Hulud campaign",
			Detected:  time.Now(),
		},
		{
			Package:   "mistralai",
			Version:   "2.4.6",
			Ecosystem: "pip",
			Feed:      "socket",
			Severity:  "critical",
			Reason:    "Mini Shai-Hulud campaign",
			Detected:  time.Now(),
		},
	}

	n, err := store.UpsertThreats(entries)
	if err != nil {
		t.Fatalf("UpsertThreats() error: %v", err)
	}
	if n != 2 {
		t.Errorf("UpsertThreats() inserted %d, want 2", n)
	}

	results, err := store.Search("@tanstack/react-router", "npm")
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Search() returned %d results, want 1", len(results))
	}
	if results[0].Package != "@tanstack/react-router" {
		t.Errorf("Package = %q", results[0].Package)
	}

	n, _ = store.UpsertThreats(entries)
	if n != 0 {
		t.Errorf("UpsertThreats() duplicate inserted %d, want 0", n)
	}
}

func TestAllThreats(t *testing.T) {
	store, _ := New(":memory:")
	defer store.Close()

	entries := []ThreatEntry{
		{Package: "pkg-a", Version: "1.0", Ecosystem: "npm", Feed: "socket", Detected: time.Now()},
		{Package: "pkg-b", Version: "2.0", Ecosystem: "pip", Feed: "osv", Detected: time.Now()},
	}
	store.UpsertThreats(entries)

	all, err := store.AllThreats("")
	if err != nil {
		t.Fatalf("AllThreats() error: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("AllThreats() = %d, want 2", len(all))
	}

	pip, err := store.AllThreats("pip")
	if err != nil {
		t.Fatalf("AllThreats('pip') error: %v", err)
	}
	if len(pip) != 1 {
		t.Errorf("AllThreats('pip') = %d, want 1", len(pip))
	}
}

func TestLastSync(t *testing.T) {
	store, _ := New(":memory:")
	defer store.Close()

	now := time.Now().UTC().Truncate(time.Second)
	store.UpsertThreats([]ThreatEntry{
		{Package: "test", Version: "1.0", Ecosystem: "npm", Feed: "socket", Detected: now},
	})

	ts, err := store.LastSync("socket")
	if err != nil {
		t.Fatalf("LastSync() error: %v", err)
	}
	if !ts.Equal(now) {
		t.Errorf("LastSync = %v, want %v", ts, now)
	}
}
