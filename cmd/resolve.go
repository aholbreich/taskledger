package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/aholbreich/tl/internal/events"
	"github.com/aholbreich/tl/internal/store"
)

func newResolveCmd() *cobra.Command {
	var (
		answer string
		asJSON bool
	)
	c := &cobra.Command{
		Use:   "resolve TASK_ID --answer ANSWER",
		Short: "Answer a pending question and return the task to open",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := store.NormalizeID(args[0])
			if answer == "" {
				return NewExitError(2, "an answer is required: use --answer")
			}

			ledger, err := requireLedger()
			if err != nil {
				return err
			}
			t, err := store.Read(ledger, taskID)
			if err != nil {
				return err
			}
			if t.Status != "pending_human" {
				return fmt.Errorf("task %s is not pending_human (status: %s)", t.ID, t.Status)
			}

			now := time.Now().UTC().Truncate(time.Second)
			t.Status = "open"
			t.UpdatedAt = now
			t.Pending = nil

			if t.Body == "" {
				t.Body = "## Notes\n"
			}
			ts := now.Format(time.RFC3339)
			t.Body += fmt.Sprintf("\n### %s — resolved\n%s\n", ts, answer)

			if err := store.Write(ledger, t); err != nil {
				return err
			}
			if err := events.Append(ledger, events.Event{
				Event:  "pending_resolved",
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
			fmt.Fprintf(cmd.OutOrStdout(), "Resolved %s\n", t.ID)
			return nil
		},
	}
	c.Flags().StringVarP(&answer, "answer", "a", "", "Answer to the pending question (required)")
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	return c
}
