package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/aholbreich/tl/internal/events"
	"github.com/aholbreich/tl/internal/store"
	"github.com/aholbreich/tl/internal/task"
)

func newCreateCmd() *cobra.Command {
	var (
		flagTitle   string
		description string
		priority    string
		taskType    string
		tags        []string
		actor       string
		asJSON      bool
	)
	c := &cobra.Command{
		Use:     "create [title] [options]",
		Short:   "Create a new task in the ledger",
		Aliases: []string{"add"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := flagTitle
			if title == "" && len(args) > 0 {
				title = args[0]
			}
			if title == "" {
				return fmt.Errorf("a task title is required (positional argument or --title)")
			}

			var err error
			priority, err = normalizePriority(priority)
			if err != nil {
				return err
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

			id, err := store.NewID(ledger)
			if err != nil {
				return err
			}
			now := time.Now().UTC().Truncate(time.Second)
			t := &task.Task{
				ID:        id,
				Title:     title,
				Status:    "open",
				Priority:  priority,
				Type:      taskType,
				CreatedAt: now,
				UpdatedAt: now,
				CreatedBy: actor,
				DependsOn: []string{},
				Tags:      append([]string{}, tags...),
				Body:      descriptionBody(description),
			}

			if err := store.Write(ledger, t); err != nil {
				return err
			}
			if err := events.Append(ledger, events.Event{
				Event:  "created",
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
			fmt.Fprintf(cmd.OutOrStdout(), "Created task %s\n", colorID(commandColorEnabled(cmd), t.ID))
			return nil
		},
	}
	c.Flags().StringVar(&flagTitle, "title", "", "Task title (required, or positional argument)")
	c.Flags().StringVarP(&description, "description", "d", "", "Task description (stored under ## Description)")
	c.Flags().StringVarP(&priority, "priority", "p", "medium", "Task priority (low|medium|high)")
	c.Flags().StringVarP(&taskType, "type", "t", "", "Task type")
	c.Flags().StringArrayVar(&tags, "tag", nil, "Tag to apply (repeatable)")
	c.Flags().StringVar(&actor, "actor", "human", "Creator actor")
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	return c
}

// descriptionBody wraps a free-text description in a "## Description" Markdown
// section, or returns "" if the description is empty.
func descriptionBody(description string) string {
	if description == "" {
		return ""
	}
	return "## Description\n\n" + description + "\n"
}

func normalizePriority(priority string) (string, error) {
	switch priority {
	case "l", "low":
		return "low", nil
	case "m", "medium":
		return "medium", nil
	case "h", "high":
		return "high", nil
	default:
		return "", NewExitError(2, "invalid priority %q: must be l/low, m/medium, or h/high", priority)
	}
}
