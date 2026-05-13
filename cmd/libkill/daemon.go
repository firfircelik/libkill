package main

import (
	"fmt"
	"os"

	"github.com/firfircelik/libkill/internal/daemon"
	"github.com/spf13/cobra"
)

func newDaemonCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "daemon",
		Short: "Run LibKill as a background daemon",
		Long:  "Runs LibKill in daemon mode, periodically scanning and sending notifications.",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := daemon.New(a.store, a.feed, a.cfg.ScanInterval)
			return d.Run()
		},
	}
}

func newInstallCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install LibKill as an OS service",
		Long:  "Installs LibKill as a launchd (macOS) or systemd (Linux) user service.",
		RunE: func(cmd *cobra.Command, args []string) error {
			binaryPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("finding binary: %w", err)
			}
			return daemon.InstallService(binaryPath)
		},
	}
}

func newUninstallCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall LibKill OS service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return daemon.UninstallService()
		},
	}
}
