package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/aholbreich/taskledger/internal/events"
	"github.com/aholbreich/taskledger/internal/store"
)

func newNoteCmd() *cobra.Command {
	var (
		actor   string
		message string
	)
	c := &cobra.Command{
		Use:   "note TASK_ID",
		Short: "Append a note to a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if message == "" {
				return fmt.Errorf("--message is required")
			}

			taskID := args[0]
			actor = ResolveActor(actor)

			ledger, err := requireLedger()
			if err != nil {
				return err
			}

			release, err := acquireLock(ledger)
			if err != nil {
				return err
			}
			defer release()

			t, err := store.Read(ledger, taskID)
			if err != nil {
				return err
			}

			now := time.Now().UTC().Truncate(time.Second)
			ts := now.Format(time.RFC3339)

			// Append note to body.
			body := strings.TrimRight(t.Body, "\n")
			note := fmt.Sprintf("\n\n## Notes\n\n### %s - %s\n\n%s\n", ts, actor, message)
			if strings.Contains(body, "## Notes") {
				// Already has a Notes section — append entry under it.
				note = fmt.Sprintf("\n### %s - %s\n\n%s\n", ts, actor, message)
			}
			t.Body = body + note
			t.UpdatedAt = now

			if err := store.Write(ledger, t); err != nil {
				return err
			}

			return events.Append(ledger, events.Event{
				Event:  "note_added",
				TaskID: t.ID,
				Actor:  actor,
			})
		},
	}
	c.Flags().StringVar(&actor, "actor", "", "Actor writing the note (resolved from env or auto-detected if unset)")
	c.Flags().StringVarP(&message, "message", "m", "", "Note message (required)")
	return c
}
