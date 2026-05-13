package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCreatesDefaultConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.DBPath == "" {
		t.Error("DBPath should not be empty")
	}
	if cfg.ScanInterval == 0 {
		t.Error("ScanInterval should not be zero")
	}
	if len(cfg.Ecosystems) == 0 {
		t.Error("Ecosystems should not be empty")
	}

	home, _ := HomeDir()
	cfgPath := filepath.Join(home, "config.json")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Error("config.json should have been created")
	}
}

func TestLoadReadsExistingConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	home, _ := HomeDir()
	os.MkdirAll(home, 0o700)
	cfgPath := filepath.Join(home, "config.json")

	existing := Config{
		DBPath:         "/custom/path/db",
		ScanInterval:   30,
		AutoClean:      true,
		Feeds:          []string{"socket"},
		Ecosystems:     []string{"npm", "pip", "cargo"},
		NotificationOn: false,
	}

	data, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(cfgPath, data, 0o600)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.DBPath != "/custom/path/db" {
		t.Errorf("DBPath = %q, want /custom/path/db", cfg.DBPath)
	}
	if !cfg.AutoClean {
		t.Error("AutoClean should be true")
	}
	if len(cfg.Ecosystems) != 3 {
		t.Errorf("len(Ecosystems) = %d, want 3", len(cfg.Ecosystems))
	}
}

func TestHomeDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir, err := HomeDir()
	if err != nil {
		t.Fatalf("HomeDir() error: %v", err)
	}

	expected := filepath.Join(tmp, ".libkill")
	if dir != expected {
		t.Errorf("HomeDir() = %q, want %q", dir, expected)
	}
}
