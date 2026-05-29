package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/aholbreich/tl/internal/events"
	"github.com/aholbreich/tl/internal/store"
	"github.com/aholbreich/tl/internal/task"
)

func newNoteCmd() *cobra.Command {
	var (
		actor   string
		message string
	)
	c := &cobra.Command{
		Use:               "note TASK_ID",
		Short:             "Append a note to a task",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeTaskIDs,
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

			t.Body = task.AppendNote(t.Body, now, actor, "note", message)
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
