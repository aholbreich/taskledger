package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/aholbreich/taskledger/internal/store"
	"github.com/aholbreich/taskledger/internal/task"
)

func newStaleCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "stale",
		Short: "List tasks with expired claims",
		RunE: func(cmd *cobra.Command, args []string) error {
			ledger, err := requireLedger()
			if err != nil {
				return err
			}
			tasks, err := store.List(ledger)
			if err != nil {
				return err
			}

			stale := filterStaleClaims(tasks, time.Now().UTC())

			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(stale)
			}

			for _, t := range stale {
				actor := "-"
				if t.Claim.Actor != nil {
					actor = *t.Claim.Actor
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", t.ID, actor)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	return c
}

func filterStaleClaims(tasks []*task.Task, now time.Time) []*task.Task {
	var stale []*task.Task
	for _, t := range tasks {
		if t.Claim.Actor == nil {
			continue
		}
		if t.Claim.ExpiresAt != nil && t.Claim.ExpiresAt.Before(now) {
			stale = append(stale, t)
		}
	}
	return stale
}
