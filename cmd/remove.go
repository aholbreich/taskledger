package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/aholbreich/tl/internal/events"
	"github.com/aholbreich/tl/internal/store"
	"github.com/aholbreich/tl/internal/task"
)

func newRemoveCmd() *cobra.Command {
	var (
		actor  string
		force  bool
		reason string
	)
	c := &cobra.Command{
		Use:               "remove TASK_ID -m REASON",
		Short:             "Remove a mistaken task from the active ledger",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeTaskIDs,
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := store.NormalizeID(args[0])
			resolved := ResolveActor(actor)
			if resolved == "" || resolved == "unknown" {
				return fmt.Errorf("actor identity is required: use --actor or set TL_ACTOR, ACTOR_NAME, or BEADS_ACTOR")
			}
			if reason == "" {
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

			if err := ensureTaskCanRemove(ledger, taskID, t, resolved, force, time.Now().UTC()); err != nil {
				return err
			}

			if err := os.Remove(store.TaskPath(ledger, taskID)); err != nil {
				return err
			}
			if err := events.Append(ledger, events.Event{Event: "removed", TaskID: taskID, Actor: resolved, Value: reason}); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Removed %s\n", taskID)
			return nil
		},
	}
	c.Flags().StringVar(&actor, "actor", "", "Actor removing the task (overrides TL_ACTOR / ACTOR_NAME / BEADS_ACTOR)")
	c.Flags().BoolVar(&force, "force", false, "Remove even when the task is not cancelled, is claimed, or has dependents")
	c.Flags().StringVarP(&reason, "message", "m", "", "Reason for removing the task (required)")
	return c
}

func ensureTaskCanRemove(ledger, taskID string, t *task.Task, actor string, force bool, now time.Time) error {
	if t.Claim.Actor != nil && *t.Claim.Actor != actor && claimIsActive(t, now) && !force {
		return NewExitError(5, "task %s is claimed by a different actor (%s); use --force to remove", taskID, *t.Claim.Actor)
	}
	if t.Status != "cancelled" && !force {
		return NewExitError(4, "task %s is %s; use --force to remove non-cancelled tasks", taskID, t.Status)
	}
	dependents, err := taskDependents(ledger, taskID)
	if err != nil {
		return err
	}
	if len(dependents) > 0 && !force {
		return NewExitError(4, "task %s has dependents (%s); use --force to remove", taskID, strings.Join(dependents, ", "))
	}
	return nil
}

func taskDependents(ledger, targetID string) ([]string, error) {
	tasks, err := store.List(ledger)
	if err != nil {
		return nil, err
	}
	var dependents []string
	for _, t := range tasks {
		if t.ID == targetID {
			continue
		}
		for _, dep := range t.DependsOn {
			if dep == targetID {
				dependents = append(dependents, t.ID)
				break
			}
		}
	}
	return dependents, nil
}
