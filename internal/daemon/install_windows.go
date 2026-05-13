//go:build windows
// +build windows

package daemon

import "fmt"

func InstallService(binaryPath string) error {
	return fmt.Errorf("windows service install not yet implemented")
}

func UninstallService() error {
	return fmt.Errorf("windows service uninstall not yet implemented")
}
