package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type Config struct {
	DBPath         string        `json:"db_path"`
	LogDir         string        `json:"log_dir"`
	ScanInterval   time.Duration `json:"scan_interval"`
	AutoClean      bool          `json:"auto_clean"`
	Feeds          []string      `json:"feeds"`
	Ecosystems     []string      `json:"ecosystems"`
	NotificationOn bool          `json:"notification_on"`
}

func Load() (*Config, error) {
	cfg := defaultConfig()
	cfgPath, err := configPath()
	if err != nil {
		return cfg, err
	}

	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o700); err != nil {
		return cfg, fmt.Errorf("creating config dir: %w", err)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := saveConfig(cfgPath, cfg); err != nil {
				return cfg, err
			}
			return cfg, nil
		}
		return cfg, fmt.Errorf("reading config: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

func HomeDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home directory: %w", err)
	}
	return filepath.Join(dir, ".libkill"), nil
}

func defaultConfig() *Config {
	home, _ := HomeDir()
	return &Config{
		DBPath:         filepath.Join(home, "libkill.db"),
		LogDir:         filepath.Join(home, "logs"),
		ScanInterval:   1 * time.Hour,
		AutoClean:      false,
		Feeds:          []string{"socket", "osv", "github"},
		Ecosystems:     []string{"npm", "pip"},
		NotificationOn: true,
	}
}

func configPath() (string, error) {
	home, err := HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config.json"), nil
}

func saveConfig(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

func dataDir() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "LibKill")
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "LibKill")
	default:
		return filepath.Join(os.Getenv("HOME"), ".local", "share", "libkill")
	}
}
