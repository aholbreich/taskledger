package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aholbreich/taskledger/internal/store"
	"github.com/aholbreich/taskledger/internal/task"
)

func newShowCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "show TASK_ID",
		Short: "Show a task in detail",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ledger, err := requireLedger()
			if err != nil {
				return err
			}
			t, err := store.Read(ledger, args[0])
			if errors.Is(err, store.ErrTaskNotFound) {
				return NewExitError(3, "task %s not found", args[0])
			}
			if err != nil {
				return err
			}

			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(t)
			}
			printTaskDetail(cmd.OutOrStdout(), t)
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	return c
}

func printTaskDetail(out interface{ Write([]byte) (int, error) }, t *task.Task) {
	fmt.Fprintf(out, "ID: %s\n", t.ID)
	fmt.Fprintf(out, "Title: %s\n", t.Title)
	fmt.Fprintf(out, "Status: %s\n", t.Status)
	fmt.Fprintf(out, "Priority: %s\n", t.Priority)

	if len(t.DependsOn) == 0 {
		fmt.Fprintln(out, "Depends On: none")
	} else {
		fmt.Fprintln(out, "Depends On:")
		for _, id := range t.DependsOn {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}

	if t.Claim.Actor == nil {
		fmt.Fprintln(out, "Claim: none")
	} else {
		fmt.Fprintf(out, "Claim: %s\n", *t.Claim.Actor)
	}

	if strings.TrimSpace(t.Body) != "" {
		fmt.Fprintln(out)
		fmt.Fprint(out, t.Body)
		if !strings.HasSuffix(t.Body, "\n") {
			fmt.Fprintln(out)
		}
	}
}
