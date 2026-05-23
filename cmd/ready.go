package cmd

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/aholbreich/tl/internal/store"
	"github.com/aholbreich/tl/internal/task"
)

func newReadyCmd() *cobra.Command {
	var asJSON bool
	var tag string
	c := &cobra.Command{
		Use:   "ready",
		Short: "List tasks that are ready to be claimed",
		RunE: func(cmd *cobra.Command, args []string) error {
			ledger, err := requireLedger()
			if err != nil {
				return err
			}
			all, err := store.List(ledger)
			if err != nil {
				return err
			}

			now := time.Now().UTC()
			ready := make([]*task.Task, 0, len(all))
			for _, t := range all {
				if !isReady(t, ledger, now) {
					continue
				}
				if tag != "" && !taskHasTag(t, tag) {
					continue
				}
				ready = append(ready, t)
			}

			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(ready)
			}

			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tStatus\tPriority\tTitle")
			for _, t := range ready {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", t.ID, t.Status, t.Priority, t.Title)
			}
			return tw.Flush()
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	c.Flags().StringVar(&tag, "tag", "", "Only show tasks carrying this tag")
	return c
}

func isReady(t *task.Task, ledger string, now time.Time) bool {
	// Must be open or in_progress (in_progress is only ready if claim expired).
	switch t.Status {
	case "open", "in_progress":
	default:
		return false // blocked, pending_human, done, cancelled
	}
	// Must not have an active claim (actor set and not expired).
	if t.Claim.Actor != nil && t.Claim.ExpiresAt != nil && t.Claim.ExpiresAt.After(now) {
		return false
	}
	// All dependencies must be done.
	for _, depID := range t.DependsOn {
		dep, err := store.Read(ledger, depID)
		if err != nil || dep.Status != "done" {
			return false
		}
	}
	return true
}
