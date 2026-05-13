package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/firfircelik/libkill/internal/scanner"
	"github.com/spf13/cobra"
)

func newScanCmd(a *app) *cobra.Command {
	var autoClean bool
	var ecosystems []string

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan system for compromised packages",
		Long:  "Scans npm global, pip, and bun caches for known compromised packages.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.forceUpdate(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: feed update failed: %v\n", err)
			}

			scanners := []scanner.Scanner{
				scanner.NewNPMScanner(a.store),
				scanner.NewPIPScanner(a.store),
			}

			ctx := context.Background()
			var allResults []scanner.Result

			fmt.Fprintln(os.Stderr, "LibKill — Supply-Chain Compromise Scanner")
			fmt.Fprintln(os.Stderr, strings.Repeat("─", 50))

			for _, s := range scanners {
				if len(ecosystems) > 0 && !contains(ecosystems, s.Ecosystem()) {
					continue
				}
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
				return nil
			}

			printResultsTable(allResults)

			if autoClean {
				fmt.Println("\nAuto-clean enabled. Removing all...")
				return cleanPackages(allResults)
			}

			return interactiveClean(allResults)
		},
	}

	cmd.Flags().BoolVar(&autoClean, "auto", false, "Automatically remove compromised packages without confirmation")
	cmd.Flags().StringSliceVar(&ecosystems, "eco", nil, "Ecosystems to scan (npm,pip)")
	return cmd
}

func printResultsTable(results []scanner.Result) {
	fmt.Printf("\n%d compromised package(s) found:\n\n", len(results))
	for i, r := range results {
		fmt.Printf("  [%d] %s@%s\n", i+1, r.Package, r.Version)
		fmt.Printf("      ecosystem: %s\n", r.Ecosystem)
		fmt.Printf("      location:  %s\n", r.Location)
		fmt.Printf("      reason:    %s\n", truncate(r.Threat.Reason, 100))
		if i < len(results)-1 {
			fmt.Println()
		}
	}
}

func interactiveClean(results []scanner.Result) error {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("\nRemove? [1-%d / all / none / quit] (default: all): ", len(results))
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "q", "quit", "n", "no", "none":
			fmt.Println("Cancelled. No packages removed.")
			return nil

		case "a", "all", "y", "yes", "":
			return cleanPackages(results)

		default:
			num, err := strconv.Atoi(input)
			if err != nil || num < 1 || num > len(results) {
				fmt.Printf("Invalid choice. Enter 1-%d, all, none, or quit.\n", len(results))
				continue
			}
			return cleanPackages([]scanner.Result{results[num-1]})
		}
	}
}

func cleanPackages(results []scanner.Result) error {
	cleaned := 0
	for _, r := range results {
		fmt.Printf("\nRemoving %s@%s...\n", r.Package, r.Version)
		switch r.Ecosystem {
		case "npm":
			if err := runCleanNPM(r); err != nil {
				fmt.Fprintf(os.Stderr, "  failed: %v\n", err)
			} else {
				fmt.Println("  removed")
				cleaned++
			}
		case "pip":
			if err := runCleanPIP(r); err != nil {
				fmt.Fprintf(os.Stderr, "  failed: %v\n", err)
			} else {
				fmt.Println("  removed")
				cleaned++
			}
		}
	}
	fmt.Printf("\n✓ Cleaned %d/%d compromised package(s)\n", cleaned, len(results))
	return nil
}

func runCleanNPM(r scanner.Result) error {
	var c *exec.Cmd
	if strings.Contains(r.Location, "global") {
		c = exec.CommandContext(context.Background(), "npm", "uninstall", "-g", r.Package)
	} else if strings.Contains(r.Location, "bun cache") {
		return cleanBunCache(r)
	} else {
		c = exec.CommandContext(context.Background(), "npm", "uninstall", r.Package)
	}
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func cleanBunCache(r scanner.Result) error {
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".bun", "install", "cache")
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return fmt.Errorf("reading bun cache: %w", err)
	}
	pkgName := strings.TrimPrefix(r.Package, "@")
	pkgName = strings.ReplaceAll(pkgName, "/", "")

	var found string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), pkgName+"@") || e.Name() == pkgName {
			found = filepath.Join(cacheDir, e.Name())
			break
		}
	}
	if found == "" {
		return fmt.Errorf("package %s not found in bun cache", r.Package)
	}

	fmt.Printf("  removing from bun cache: %s\n", found)
	if err := os.RemoveAll(found); err != nil {
		c := exec.CommandContext(context.Background(), "rm", "-rf", found)
		c.Stdout = os.Stderr
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("cannot remove %s: %w", found, err)
		}
	}

	// Also run bun pm cache rm to clean metadata
	exec.CommandContext(context.Background(), "bun", "pm", "cache", "rm").Run()

	return nil
}

func runCleanPIP(r scanner.Result) error {
	c := exec.CommandContext(context.Background(), "pip3", "uninstall", "-y", r.Package)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
