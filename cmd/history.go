package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/aholbreich/tl/internal/events"
	"github.com/aholbreich/tl/internal/store"
)

func newHistoryCmd() *cobra.Command {
	var asJSON bool
	var since string
	c := &cobra.Command{
		Use:   "history [TASK_ID]",
		Short: "Show event history for a task or recent ledger activity",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && since == "" {
				return NewExitError(2, "a task ID or --since duration is required")
			}

			ledger, err := requireLedger()
			if err != nil {
				return err
			}

			taskID := ""
			if len(args) == 1 {
				t, err := store.Read(ledger, args[0])
				if errors.Is(err, store.ErrTaskNotFound) {
					return NewExitError(3, "task %s not found", args[0])
				}
				if err != nil {
					return err
				}
				taskID = t.ID
			}

			var cutoff time.Time
			if since != "" {
				d, err := parseHistoryDuration(since)
				if err != nil {
					return NewExitError(2, "invalid --since duration %q", since)
				}
				cutoff = time.Now().UTC().Add(-d)
			}

			history, err := events.ReadAll(ledger)
			if err != nil {
				return err
			}
			history = filterHistoryEvents(history, taskID, cutoff)

			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(history)
			}

			for _, e := range history {
				actor := e.Actor
				if actor == "" {
					actor = "-"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", e.Time.Format(time.RFC3339), e.Event, e.TaskID, actor)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	c.Flags().StringVar(&since, "since", "", "Only show events within this duration (e.g. 24h, 7d)")
	return c
}

func filterHistoryEvents(in []events.Event, taskID string, cutoff time.Time) []events.Event {
	out := make([]events.Event, 0, len(in))
	for _, e := range in {
		if taskID != "" && e.TaskID != taskID {
			continue
		}
		if !cutoff.IsZero() && e.Time.Before(cutoff) {
			continue
		}
		out = append(out, e)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Time.Before(out[j].Time)
	})
	return out
}

func parseHistoryDuration(s string) (time.Duration, error) {
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}
	if strings.HasSuffix(s, "d") {
		days, err := strconv.ParseFloat(strings.TrimSuffix(s, "d"), 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(days * float64(24*time.Hour)), nil
	}
	return 0, fmt.Errorf("invalid duration")
}
