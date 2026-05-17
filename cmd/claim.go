package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/aholbreich/taskledger/internal/events"
	"github.com/aholbreich/taskledger/internal/repo"
	"github.com/aholbreich/taskledger/internal/store"
)

func newClaimCmd() *cobra.Command {
	var (
		actor  string
		ttl    string
		asJSON bool
	)
	c := &cobra.Command{
		Use:   "claim TASK_ID",
		Short: "Claim a task with a time-limited lease",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]
			actor = ResolveActor(actor)

			ledger, err := requireLedger()
			if err != nil {
				return err
			}

			cfg, err := repo.LoadConfig(ledger)
			if err != nil {
				return err
			}

			// Parse TTL: --ttl flag wins, else config default.
			var ttlDuration time.Duration
			src := ttl
			if src == "" {
				src = cfg.DefaultClaimTTL
			}
			ttlDuration, err = time.ParseDuration(src)
			if err != nil {
				return NewExitError(2, "invalid claim TTL %q", src)
			}

			t, err := store.Read(ledger, taskID)
			if errors.Is(err, store.ErrTaskNotFound) {
				return NewExitError(3, "task %s not found", taskID)
			}
			if err != nil {
				return err
			}

			// Reject if another actor holds an active (non-expired) claim.
			if t.Claim.Actor != nil && *t.Claim.Actor != actor {
				if t.Claim.ExpiresAt != nil && t.Claim.ExpiresAt.After(time.Now().UTC()) {
					return NewExitError(5, "task %s is already claimed by %s", taskID, *t.Claim.Actor)
				}
			}

			// Must be open.
			if t.Status != "open" {
				return NewExitError(4, "task %s is not ready (status %s)", taskID, t.Status)
			}

			// All dependencies must be done.
			for _, depID := range t.DependsOn {
				dep, err := store.Read(ledger, depID)
				if errors.Is(err, store.ErrTaskNotFound) {
					return fmt.Errorf("task %s depends on %s which does not exist", taskID, depID)
				}
				if err != nil {
					return err
				}
				if dep.Status != "done" {
					return NewExitError(4, "task %s is not ready: dependency %s is not done (status %s)", taskID, depID, dep.Status)
				}
			}

			now := time.Now().UTC().Truncate(time.Second)
			expires := now.Add(ttlDuration)
			t.Claim.Actor = &actor
			t.Claim.ClaimedAt = &now
			t.Claim.ExpiresAt = &expires
			t.Claim.HeartbeatAt = &now
			t.Status = "in_progress"
			t.UpdatedAt = now

			if err := store.Write(ledger, t); err != nil {
				return err
			}
			if err := events.Append(ledger, events.Event{
				Event:  "claimed",
				TaskID: t.ID,
				Actor:  actor,
			}); err != nil {
				return err
			}

			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(t)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Claimed task %s (%s, expires %s)\n", t.ID, actor, expires.Format(time.RFC3339))
			return nil
		},
	}
	c.Flags().StringVar(&actor, "actor", "", "Claiming actor (resolved from env or auto-detected if unset)")
	c.Flags().StringVar(&ttl, "ttl", "", "Lease duration, e.g. 60m or 2h (default from config)")
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	return c
}
