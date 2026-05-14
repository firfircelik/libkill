package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newUpdateCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update the local threat database from all feeds",
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := a.feed.Update(context.Background())
			if err != nil {
				return fmt.Errorf("update: %w", err)
			}
			if n == 0 {
				fmt.Println("Threat database is up to date.")
			} else {
				fmt.Printf("Added %d new threat entries.\n", n)
			}

			m, err := a.feed.UpdateMalware(context.Background())
			if err != nil {
				return fmt.Errorf("update malware: %w", err)
			}
			if m > 0 {
				fmt.Printf("Added %d new malware family entries.\n", m)
			}

			return nil
		},
	}
}
