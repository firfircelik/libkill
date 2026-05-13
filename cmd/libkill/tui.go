package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/firfircelik/libkill/internal/tui"
	"github.com/spf13/cobra"
)

func newTUICmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive terminal UI",
		Long: `Opens an interactive terminal interface for scanning,
reviewing, and removing compromised packages.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.forceUpdate(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: feed update failed: %v\n", err)
			}

			m := tui.New(a.store)
			p := tea.NewProgram(m, tea.WithAltScreen())

			if _, err := p.Run(); err != nil {
				return fmt.Errorf("tui: %w", err)
			}
			return nil
		},
	}
}
