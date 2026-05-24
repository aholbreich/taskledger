package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/aholbreich/tl/internal/events"
	"github.com/aholbreich/tl/internal/store"
	"github.com/aholbreich/tl/internal/task"
)

func newBlockCmd() *cobra.Command {
	var (
		actor   string
		message string
		asJSON  bool
	)
	c := &cobra.Command{
		Use:   "block TASK_ID -m REASON",
		Short: "Mark a task blocked with a reason",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := store.NormalizeID(args[0])
			resolved := ResolveActor(actor)
			if message == "" {
				return NewExitError(2, "a reason is required: use -m / --message")
			}

			ledger, err := requireLedger()
			if err != nil {
				return err
			}
			t, err := store.Read(ledger, taskID)
			if err != nil {
				return err
			}

			now := time.Now().UTC().Truncate(time.Second)
			t.Status = "blocked"
			t.UpdatedAt = now
			t.Claim = task.Claim{}

			t.Body = task.AppendNote(t.Body, now, resolved, "blocked", message)

			if err := store.Write(ledger, t); err != nil {
				return err
			}
			if err := events.Append(ledger, events.Event{
				Event:  "blocked",
				TaskID: t.ID,
				Actor:  resolved,
			}); err != nil {
				return err
			}

			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(t)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Blocked %s\n", t.ID)
			return nil
		},
	}
	c.Flags().StringVar(&actor, "actor", "", "Actor performing the block (overrides TL_ACTOR / ACTOR_NAME / BEADS_ACTOR)")
	c.Flags().StringVarP(&message, "message", "m", "", "Reason the task is blocked (required)")
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	return c
}
