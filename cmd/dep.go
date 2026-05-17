package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aholbreich/taskledger/internal/events"
	"github.com/aholbreich/taskledger/internal/store"
)

func newDepCmd() *cobra.Command {
	dep := &cobra.Command{
		Use:   "dep",
		Short: "Manage task dependencies",
	}
	dep.AddCommand(newDepAddCmd())
	dep.AddCommand(newDepRemoveCmd())
	return dep
}

func newDepAddCmd() *cobra.Command {
	var on string
	c := &cobra.Command{
		Use:   "add TASK_ID --on TASK_ID",
		Short: "Add a dependency link between tasks",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return depAdd(cmd, args, on)
		},
	}
	c.Flags().StringVar(&on, "on", "", "Target task this task depends on (required)")
	return c
}

func newDepRemoveCmd() *cobra.Command {
	var on string
	c := &cobra.Command{
		Use:   "remove TASK_ID --on TASK_ID",
		Short: "Remove a dependency link between tasks",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return depRemove(cmd, args, on)
		},
	}
	c.Flags().StringVar(&on, "on", "", "Target task to drop as a dependency (required)")
	return c
}

func depAdd(_ *cobra.Command, args []string, on string) error {
	if on == "" {
		return fmt.Errorf("--on is required")
	}
	sourceID := store.NormalizeID(args[0])
	targetID := store.NormalizeID(on)

	ledger, err := requireLedger()
	if err != nil {
		return err
	}

	release, err := acquireLock(ledger)
	if err != nil {
		return err
	}
	defer release()

	src, err := store.Read(ledger, sourceID)
	if errors.Is(err, store.ErrTaskNotFound) {
		return NewExitError(3, "task %s not found", sourceID)
	}
	if err != nil {
		return err
	}

	_, err = store.Read(ledger, targetID)
	if errors.Is(err, store.ErrTaskNotFound) {
		return NewExitError(3, "task %s not found", targetID)
	}
	if err != nil {
		return err
	}

	for _, d := range src.DependsOn {
		if d == targetID {
			return nil
		}
	}

	src.DependsOn = append(src.DependsOn, targetID)
	if err := store.Write(ledger, src); err != nil {
		return err
	}

	return events.Append(ledger, events.Event{
		Event:  "dependency_added",
		TaskID: sourceID,
	})
}

func depRemove(_ *cobra.Command, args []string, on string) error {
	if on == "" {
		return fmt.Errorf("--on is required")
	}
	sourceID := store.NormalizeID(args[0])
	targetID := store.NormalizeID(on)

	ledger, err := requireLedger()
	if err != nil {
		return err
	}

	release, err := acquireLock(ledger)
	if err != nil {
		return err
	}
	defer release()

	src, err := store.Read(ledger, sourceID)
	if errors.Is(err, store.ErrTaskNotFound) {
		return NewExitError(3, "task %s not found", sourceID)
	}
	if err != nil {
		return err
	}

	_, err = store.Read(ledger, targetID)
	if errors.Is(err, store.ErrTaskNotFound) {
		return NewExitError(3, "task %s not found", targetID)
	}
	if err != nil {
		return err
	}

	found := false
	filtered := src.DependsOn[:0]
	for _, d := range src.DependsOn {
		if d == targetID {
			found = true
			continue
		}
		filtered = append(filtered, d)
	}
	if !found {
		return nil
	}
	src.DependsOn = filtered

	if err := store.Write(ledger, src); err != nil {
		return err
	}

	return events.Append(ledger, events.Event{
		Event:  "dependency_removed",
		TaskID: sourceID,
	})
}
