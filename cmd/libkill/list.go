package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newListCmd(a *app) *cobra.Command {
	var ecosystem string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List known compromised packages in local database",
		RunE: func(cmd *cobra.Command, args []string) error {
			threats, err := a.store.AllThreats(ecosystem)
			if err != nil {
				return fmt.Errorf("list: %w", err)
			}

			if len(threats) == 0 {
				fmt.Println("No threat entries in database. Run 'libkill update' first.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "PACKAGE\tVERSION\tECOSYSTEM\tFEED\tREASON")
			for _, t := range threats {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", t.Package, t.Version, t.Ecosystem, t.Feed, t.Reason)
			}
			w.Flush()
			fmt.Printf("\nTotal: %d known compromised artifacts\n", len(threats))
			return nil
		},
	}

	cmd.Flags().StringVar(&ecosystem, "eco", "", "Filter by ecosystem (npm, pip)")
	return cmd
}
