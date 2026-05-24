package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/aholbreich/tl/internal/events"
	"github.com/aholbreich/tl/internal/store"
	"github.com/aholbreich/tl/internal/task"
)

func newCancelCmd() *cobra.Command {
	var (
		actor   string
		force   bool
		message string
		asJSON  bool
	)
	c := &cobra.Command{
		Use:   "cancel TASK_ID -m REASON",
		Short: "Cancel a task with a reason",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := store.NormalizeID(args[0])
			resolved := ResolveActor(actor)
			if resolved == "" || resolved == "unknown" {
				return fmt.Errorf("actor identity is required: use --actor or set TL_ACTOR, ACTOR_NAME, or BEADS_ACTOR")
			}
			if message == "" {
				return NewExitError(2, "a reason is required: use -m / --message")
			}

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
			if errors.Is(err, store.ErrTaskNotFound) {
				return NewExitError(3, "task %s not found", taskID)
			}
			if err != nil {
				return err
			}

			if err := ensureTaskCanCancel(t, resolved, force, time.Now().UTC()); err != nil {
				return err
			}
			now := time.Now().UTC().Truncate(time.Second)
			t.Status = "cancelled"
			t.UpdatedAt = now
			t.Claim = task.Claim{}

			t.Body = task.AppendNote(t.Body, now, resolved, "cancelled", message)

			if err := store.Write(ledger, t); err != nil {
				return err
			}
			if err := events.Append(ledger, events.Event{
				Event:  "cancelled",
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
			fmt.Fprintf(cmd.OutOrStdout(), "Cancelled %s\n", t.ID)
			return nil
		},
	}
	c.Flags().StringVar(&actor, "actor", "", "Actor cancelling the task (overrides TL_ACTOR / ACTOR_NAME / BEADS_ACTOR)")
	c.Flags().BoolVar(&force, "force", false, "Cancel even when another actor holds an active claim")
	c.Flags().StringVarP(&message, "message", "m", "", "Reason for cancelling the task (required)")
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	return c
}

func ensureTaskCanCancel(t *task.Task, actor string, force bool, now time.Time) error {
	switch t.Status {
	case "done":
		return NewExitError(4, "task %s is already closed", t.ID)
	case "cancelled":
		return fmt.Errorf("task %s is already cancelled", t.ID)
	case "blocked":
		return NewExitError(4, "task %s is blocked", t.ID)
	case "open", "in_progress", "pending_human":
		// cancellable
	default:
		return NewExitError(4, "task %s cannot be cancelled while status is %s", t.ID, t.Status)
	}

	if t.Claim.Actor != nil && *t.Claim.Actor != actor && claimIsActive(t, now) && !force {
		return fmt.Errorf("task %s is claimed by a different actor (%s); use --force to cancel", t.ID, *t.Claim.Actor)
	}
	return nil
}
