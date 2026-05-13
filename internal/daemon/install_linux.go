//go:build linux
// +build linux

package daemon

import (
	"fmt"
	"os"
	"path/filepath"
)

const unitTemplate = `[Unit]
Description=LibKill Supply-Chain Security Daemon
After=network.target

[Service]
Type=simple
ExecStart=%s daemon
Restart=always
RestartSec=10

[Install]
WantedBy=default.target
`

func InstallService(binaryPath string) error {
	home, _ := os.UserHomeDir()
	unitDir := filepath.Join(home, ".config", "systemd", "user")

	if err := os.MkdirAll(unitDir, 0700); err != nil {
		return fmt.Errorf("creating systemd user dir: %w", err)
	}

	unitPath := filepath.Join(unitDir, "libkill.service")
	unit := fmt.Sprintf(unitTemplate, binaryPath)
	return os.WriteFile(unitPath, []byte(unit), 0644)
}

func UninstallService() error {
	home, _ := os.UserHomeDir()
	unitPath := filepath.Join(home, ".config", "systemd", "user", "libkill.service")
	if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing unit: %w", err)
	}
	return nil
}
