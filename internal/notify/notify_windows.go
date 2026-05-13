//go:build windows
// +build windows

package notify

import "fmt"

func sendNotification(title, message string) error {
	return fmt.Errorf("windows notifications not yet implemented")
}
