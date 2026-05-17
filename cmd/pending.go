package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/aholbreich/taskledger/internal/events"
	"github.com/aholbreich/taskledger/internal/store"
	"github.com/aholbreich/taskledger/internal/task"
)

func newPendingCmd() *cobra.Command {
	var (
		actor    string
		question string
		asJSON   bool
	)
	c := &cobra.Command{
		Use:   "pending TASK_ID",
		Short: "Mark a task as pending human input",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := store.NormalizeID(args[0])
			resolved := ResolveActor(actor)
			if resolved == "" || resolved == "unknown" {
				return fmt.Errorf("actor identity is required: use --actor or set TL_ACTOR, ACTOR_NAME, or BEADS_ACTOR")
			}
			if question == "" {
				return fmt.Errorf("--question is required")
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
			t.Status = "pending_human"
			t.UpdatedAt = now
			t.Claim = task.Claim{}
			t.Pending = &task.Pending{
				Question:    question,
				Requester:   resolved,
				RequestedAt: now,
			}

			if err := store.Write(ledger, t); err != nil {
				return err
			}
			if err := events.Append(ledger, events.Event{
				Event:  "pending_requested",
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
			fmt.Fprintf(cmd.OutOrStdout(), "Marked %s pending_human\n", t.ID)
			return nil
		},
	}
	c.Flags().StringVar(&actor, "actor", "", "Actor requesting human input (overrides TL_ACTOR / ACTOR_NAME / BEADS_ACTOR)")
	c.Flags().StringVarP(&question, "question", "q", "", "Question for the human (required)")
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	return c
}
