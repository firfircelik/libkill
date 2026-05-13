//go:build linux
// +build linux

package notify

import "os/exec"

func sendNotification(title, message string) error {
	cmd := exec.Command("notify-send", title, message, "--urgency=critical")
	return cmd.Run()
}
