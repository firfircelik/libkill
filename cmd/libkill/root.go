package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/firfircelik/libkill/internal/scanner"
	"github.com/firfircelik/libkill/internal/tui"
	"github.com/spf13/cobra"
)

func newRootCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "libkill",
		Short: "LibKill — supply-chain compromise scanner and cleaner",
		Long: `LibKill scans your system for npm, pip, and other packages
known to be compromised in supply-chain attacks.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !cmd.Flags().Changed("help") {
				runInteractiveMenu(a)
				return nil
			}
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newScanCmd(a),
		newUpdateCmd(a),
		newListCmd(a),
		newDaemonCmd(a),
		newTUICmd(a),
		newInstallCmd(a),
		newUninstallCmd(a),
	)

	return cmd
}

func runInteractiveMenu(a *app) {
	fmt.Println()
	fmt.Println("  LibKill — Supply-Chain Compromise Scanner")
	fmt.Println("  ─────────────────────────────────────────")
	fmt.Println()
	fmt.Println("  [1] Scan system for compromised packages")
	fmt.Println("  [2] Update threat database")
	fmt.Println("  [3] List known threats")
	fmt.Println("  [4] Launch TUI (interactive)")
	fmt.Println("  [q] Quit")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("  Choice [1-4/q]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "1":
			runTerminalScan(a)
			return
		case "2":
			fmt.Println()
			n, err := a.feed.Update(context.Background())
			if err != nil {
				fmt.Fprintf(os.Stderr, "  error: %v\n", err)
			} else if n == 0 {
				fmt.Println("  Threat database is up to date.")
			} else {
				fmt.Printf("  Added %d new threat entries.\n", n)
			}
			return
		case "3":
			runTerminalList(a)
			return
		case "4":
			m := tui.New(a.store)
			p := tea.NewProgram(m, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
			}
			return
		case "q", "":
			return
		default:
			fmt.Println("  Invalid choice. Enter 1-4 or q.")
		}
	}
}

func runTerminalScan(a *app) {
	fmt.Fprintln(os.Stderr, "\nLibKill — Supply-Chain Compromise Scanner")
	fmt.Fprintln(os.Stderr, strings.Repeat("─", 50))

	if err := a.forceUpdate(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: feed update failed: %v\n", err)
	}

	scanners := []scanner.Scanner{
		scanner.NewNPMScanner(a.store),
		scanner.NewPIPScanner(a.store),
	}

	ctx := context.Background()
	var allResults []scanner.Result

	for _, s := range scanners {
		fmt.Fprintf(os.Stderr, "Scanning %s packages...\n", s.Name())
		results, total, err := s.Scan(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s error: %v\n", s.Name(), err)
			continue
		}
		allResults = append(allResults, results...)
		if len(results) > 0 {
			fmt.Fprintf(os.Stderr, "  %d scanned → %d threats found\n", total, len(results))
		} else {
			fmt.Fprintf(os.Stderr, "  %d scanned → clean\n", total)
		}
	}

	fmt.Fprintln(os.Stderr, strings.Repeat("─", 50))

	if len(allResults) == 0 {
		fmt.Println("\n✓ No compromised packages found. Your system is clean.")
		return
	}

	printResultsTable(allResults)
	interactiveClean(allResults)
}

func runTerminalList(a *app) {
	total, _ := a.store.CountThreats()
	threats, err := a.store.AllThreats("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	fmt.Printf("\n  %d known compromised package artifacts\n\n", total)
	for i, t := range threats {
		if i >= 10 && len(threats) > 10 {
			continue
		}
		fmt.Printf("  %s@%s (%s) — %s\n", t.Package, t.Version, t.Ecosystem, truncate(t.Reason, 60))
	}
	if len(threats) > 10 {
		fmt.Printf("\n  ... and %d more\n", total-10)
	}
}
