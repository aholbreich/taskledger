package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aholbreich/tl/internal/store"
	"github.com/aholbreich/tl/internal/task"
)

func newShowCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:               "show TASK_ID",
		Short:             "Show a task in detail",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeTaskIDs,
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
			printTaskDetail(cmd.OutOrStdout(), t, commandColorEnabled(cmd))
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	return c
}

func printTaskDetail(out interface{ Write([]byte) (int, error) }, t *task.Task, useColor bool) {
	printTaskField(out, useColor, "ID", t.ID)
	printTaskField(out, useColor, "Title", t.Title)
	printTaskField(out, useColor, "Status", colorStatus(useColor, t.Status))
	printTaskField(out, useColor, "Priority", colorPriority(useColor, t.Priority))

	if len(t.DependsOn) == 0 {
		printTaskField(out, useColor, "Depends On", "none")
	} else {
		fmt.Fprintf(out, "%s:\n", colorFieldLabel(useColor, "Depends On"))
		for _, id := range t.DependsOn {
			fmt.Fprintf(out, "  - %s\n", colorFieldValue(useColor, id))
		}
	}

	if t.Claim.Actor == nil {
		printTaskField(out, useColor, "Claim", "none")
	} else {
		printTaskField(out, useColor, "Claim", *t.Claim.Actor)
	}

	if strings.TrimSpace(t.Body) != "" {
		fmt.Fprintln(out)
		fmt.Fprint(out, colorMarkdownHeadings(useColor, t.Body))
		if !strings.HasSuffix(t.Body, "\n") {
			fmt.Fprintln(out)
		}
	}
}

func printTaskField(out interface{ Write([]byte) (int, error) }, useColor bool, label, value string) {
	fmt.Fprintf(out, "%s: %s\n", colorFieldLabel(useColor, label), colorFieldValue(useColor, value))
}
