//go:build darwin
// +build darwin

package notify

import (
	"fmt"
	"os/exec"
)

func sendNotification(title, message string) error {
	script := fmt.Sprintf(
		`display notification "%s" with title "%s" sound name "Glass"`,
		escapeAppleScript(message),
		escapeAppleScript(title),
	)
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

func escapeAppleScript(s string) string {
	result := ""
	for _, c := range s {
		switch c {
		case '\\', '"':
			result += "\\" + string(c)
		default:
			result += string(c)
		}
	}
	return result
}
