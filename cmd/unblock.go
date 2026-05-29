package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/aholbreich/tl/internal/events"
	"github.com/aholbreich/tl/internal/store"
)

func newUnblockCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:               "unblock TASK_ID",
		Short:             "Remove a block and return a task to open",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeTaskIDs,
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := store.NormalizeID(args[0])

			ledger, err := requireLedger()
			if err != nil {
				return err
			}
			t, err := store.Read(ledger, taskID)
			if err != nil {
				return err
			}
			if t.Status != "blocked" {
				return fmt.Errorf("task %s is not blocked (status: %s)", t.ID, t.Status)
			}

			now := time.Now().UTC().Truncate(time.Second)
			t.Status = "open"
			t.UpdatedAt = now

			if err := store.Write(ledger, t); err != nil {
				return err
			}
			if err := events.Append(ledger, events.Event{
				Event:  "unblocked",
				TaskID: t.ID,
				Actor:  ResolveActor(""),
			}); err != nil {
				return err
			}

			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(t)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Unblocked %s\n", t.ID)
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	return c
}
