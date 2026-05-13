//go:build darwin
// +build darwin

package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.libkill.daemon</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>daemon</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>%s/Library/Logs/libkill.log</string>
	<key>StandardErrorPath</key>
	<string>%s/Library/Logs/libkill.err</string>
</dict>
</plist>`

func InstallService(binaryPath string) error {
	home, _ := os.UserHomeDir()
	plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.libkill.daemon.plist")

	if err := os.MkdirAll(filepath.Dir(plistPath), 0700); err != nil {
		return fmt.Errorf("creating LaunchAgents dir: %w", err)
	}

	plist := fmt.Sprintf(plistTemplate, binaryPath, home, home)
	if err := os.WriteFile(plistPath, []byte(plist), 0644); err != nil {
		return fmt.Errorf("writing plist: %w", err)
	}

	cmd := exec.Command("launchctl", "load", plistPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl load: %w: %s", err, strings.TrimSpace(string(out)))
	}

	fmt.Printf("Service installed: %s\n", plistPath)
	return nil
}

func UninstallService() error {
	home, _ := os.UserHomeDir()
	plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.libkill.daemon.plist")

	cmd := exec.Command("launchctl", "unload", plistPath)
	cmd.Run()

	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing plist: %w", err)
	}

	fmt.Println("Service uninstalled")
	return nil
}
