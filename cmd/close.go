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

func newCloseCmd() *cobra.Command {
	var (
		actor  string
		force  bool
		asJSON bool
	)
	c := &cobra.Command{
		Use:               "close TASK_ID",
		Short:             "Close a completed task",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeTaskIDs,
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := store.NormalizeID(args[0])
			resolved := ResolveActor(actor)
			if resolved == "" || resolved == "unknown" {
				return fmt.Errorf("actor identity is required: use --actor or set TL_ACTOR, ACTOR_NAME, or BEADS_ACTOR")
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

			if err := ensureTaskCanClose(t, resolved, force, time.Now().UTC()); err != nil {
				return err
			}
			now := time.Now().UTC().Truncate(time.Second)
			t.Status = "done"
			t.UpdatedAt = now
			t.Claim = task.Claim{}

			if err := store.Write(ledger, t); err != nil {
				return err
			}
			if err := events.Append(ledger, events.Event{
				Event:  "closed",
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
			fmt.Fprintf(cmd.OutOrStdout(), "Closed task %s\n", t.ID)
			return nil
		},
	}
	c.Flags().StringVar(&actor, "actor", "", "Actor closing the task (overrides TL_ACTOR / ACTOR_NAME / BEADS_ACTOR)")
	c.Flags().BoolVar(&force, "force", false, "Close even when another actor holds an active claim")
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	return c
}

func ensureTaskCanClose(t *task.Task, actor string, force bool, now time.Time) error {
	switch t.Status {
	case "done":
		return NewExitError(4, "task %s is already closed", t.ID)
	case "blocked":
		return NewExitError(4, "task %s is blocked", t.ID)
	case "open", "in_progress":
		// closable
	default:
		return NewExitError(4, "task %s cannot be closed while status is %s", t.ID, t.Status)
	}

	if t.Claim.Actor != nil && *t.Claim.Actor != actor && claimIsActive(t, now) && !force {
		return NewExitError(5, "task %s is claimed by a different actor (%s); use --force to close", t.ID, *t.Claim.Actor)
	}
	return nil
}

func claimIsActive(t *task.Task, now time.Time) bool {
	if t.Claim.Actor == nil {
		return false
	}
	return t.Claim.ExpiresAt == nil || t.Claim.ExpiresAt.After(now)
}
