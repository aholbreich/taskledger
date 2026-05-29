package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/aholbreich/tl/internal/events"
	"github.com/aholbreich/tl/internal/repo"
	"github.com/aholbreich/tl/internal/store"
	"github.com/aholbreich/tl/internal/task"
)

func newClaimCmd() *cobra.Command {
	var (
		flagActor string
		ttl       string
		asJSON    bool
		force     bool
	)
	c := &cobra.Command{
		Use:               "claim TASK_ID",
		Short:             "Claim a task with a time-limited lease",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeTaskIDs,
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := store.NormalizeID(args[0])

			resolved := ResolveActor(flagActor)
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

			// Same-actor renewal: extend the lease without requiring open status.
			if t.Claim.Actor != nil && *t.Claim.Actor == resolved {
				if t.Claim.ExpiresAt != nil && t.Claim.ExpiresAt.After(time.Now().UTC()) {
					return renewClaim(ledger, t, ttlDuration, resolved, asJSON, cmd)
				}
			}

			// Reject if another actor holds an active claim (unless --force).
			if !force && t.Claim.Actor != nil && *t.Claim.Actor != resolved {
				if t.Claim.ExpiresAt != nil && t.Claim.ExpiresAt.After(time.Now().UTC()) {
					return NewExitError(5, "task %s is already claimed by %s", taskID, *t.Claim.Actor)
				}
			}

			// Must be open (unless --force).
			if t.Status != "open" && !force {
				return NewExitError(4, "task %s is not ready (status %s)", taskID, t.Status)
			}

			// All dependencies must be done (unless --force).
			if !force {
				if err := checkDeps(ledger, t); err != nil {
					return err
				}
			}

			return claimTask(ledger, t, ttlDuration, resolved, asJSON, cmd)
		},
	}
	c.Flags().StringVar(&flagActor, "actor", "", "Claiming actor (overrides TL_ACTOR / ACTOR_NAME / BEADS_ACTOR)")
	c.Flags().StringVar(&ttl, "ttl", "", "Lease duration, e.g. 60m or 2h (default from config)")
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	c.Flags().BoolVar(&force, "force", false, "Force claim even when another actor holds an active claim")
	return c
}

// checkDeps verifies every dependency of t has status "done".
func checkDeps(ledger string, t *task.Task) error {
	for _, depID := range t.DependsOn {
		dep, err := store.Read(ledger, depID)
		if errors.Is(err, store.ErrTaskNotFound) {
			return fmt.Errorf("task %s depends on %s which does not exist", t.ID, depID)
		}
		if err != nil {
			return err
		}
		if dep.Status != "done" {
			return NewExitError(4, "task %s is not ready: dependency %s is not done (status %s)", t.ID, depID, dep.Status)
		}
	}
	return nil
}

// claimTask sets the claim on t and writes it.
func claimTask(ledger string, t *task.Task, ttlDuration time.Duration, resolved string, asJSON bool, cmd *cobra.Command) error {
	now := time.Now().UTC().Truncate(time.Second)
	expires := now.Add(ttlDuration)
	t.Claim.Actor = &resolved
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
		Actor:  resolved,
	}); err != nil {
		return err
	}

	return emitClaimResult(t, resolved, expires, asJSON, cmd)
}

// renewClaim extends the existing claim expiry without changing status.
func renewClaim(ledger string, t *task.Task, ttlDuration time.Duration, resolved string, asJSON bool, cmd *cobra.Command) error {
	now := time.Now().UTC().Truncate(time.Second)
	expires := now.Add(ttlDuration)
	t.Claim.ExpiresAt = &expires
	t.Claim.HeartbeatAt = &now
	t.UpdatedAt = now

	if err := store.Write(ledger, t); err != nil {
		return err
	}
	if err := events.Append(ledger, events.Event{
		Event:  "claim_renewed",
		TaskID: t.ID,
		Actor:  resolved,
	}); err != nil {
		return err
	}

	return emitClaimResult(t, resolved, expires, asJSON, cmd)
}

func emitClaimResult(t *task.Task, actor string, expires time.Time, asJSON bool, cmd *cobra.Command) error {
	if asJSON {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(t)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Claimed task %s (%s, expires %s)\n", t.ID, actor, expires.Format(time.RFC3339))
	return nil
}
