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

func newReleaseCmd() *cobra.Command {
	var (
		actor  string
		force  bool
		asJSON bool
	)
	c := &cobra.Command{
		Use:   "release TASK_ID",
		Short: "Voluntarily release a claim on a task",
		Args:  cobra.ExactArgs(1),
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
			if err != nil {
				return err
			}

			if t.Claim.Actor == nil {
				return NewExitError(4, "task %s is not claimed", t.ID)
			}

			now := time.Now().UTC()
			if *t.Claim.Actor != resolved && !force {
				return NewExitError(5, "task %s is claimed by a different actor (%s); use --force to release", t.ID, *t.Claim.Actor)
			}

			t.Status = "open"
			t.UpdatedAt = now.Truncate(time.Second)
			t.Claim = task.Claim{}

			if err := store.Write(ledger, t); err != nil {
				return err
			}
			if err := events.Append(ledger, events.Event{
				Event:  "released",
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
			fmt.Fprintf(cmd.OutOrStdout(), "Released claim on %s\n", t.ID)
			return nil
		},
	}
	c.Flags().StringVar(&actor, "actor", "", "Releasing actor (overrides TL_ACTOR / ACTOR_NAME / BEADS_ACTOR)")
	c.Flags().BoolVar(&force, "force", false, "Release even when another actor holds the claim")
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	return c
}
