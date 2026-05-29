package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/aholbreich/tl/internal/events"
	"github.com/aholbreich/tl/internal/store"
	"github.com/aholbreich/tl/internal/task"
)

func newRefineCmd() *cobra.Command {
	var (
		title       string
		description string
		taskType    string
		priority    string
		addRefs     []string
		removeRefs  []string
		editMode    bool
		asJSON      bool
	)
	c := &cobra.Command{
		Use:               "refine TASK_ID",
		Short:             "Update editable task fields",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeTaskIDs,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			updateTitle := flags.Changed("title")
			updateDescription := flags.Changed("description")
			updateType := flags.Changed("type")
			updatePriority := flags.Changed("priority")
			hasFieldFlag := updateTitle || updateDescription || updateType || updatePriority
			hasRefFlag := flags.Changed("add-ref") || flags.Changed("remove-ref")

			taskID := store.NormalizeID(args[0])
			ledger, err := requireLedger()
			if err != nil {
				return err
			}

			if editMode {
				if hasFieldFlag || hasRefFlag {
					return NewExitError(2, "--edit cannot be combined with field flags")
				}
				return refineWithEditor(cmd, ledger, taskID, asJSON)
			}

			if !hasFieldFlag && !hasRefFlag {
				return NewExitError(2, "no fields were given to refine")
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

			if updateTitle {
				t.Title = title
			}
			if updateDescription {
				t.Body = task.SetDescription(t.Body, description)
			}
			if updateType {
				t.Type = taskType
			}
			if updatePriority {
				normalized, err := normalizePriority(priority)
				if err != nil {
					return err
				}
				t.Priority = normalized
			}

			// Reference mutations are idempotent and, like `dep add/remove`,
			// emit their own events rather than a generic `refined`.
			refEvents := applyRefMutations(t, addRefs, removeRefs)

			// A refine with only no-op ref flags changes nothing: don't write
			// or emit events, just report success.
			if !hasFieldFlag && len(refEvents) == 0 {
				return reportRefineNoChange(cmd, t, asJSON)
			}

			t.UpdatedAt = time.Now().UTC().Truncate(time.Second)
			if err := store.Write(ledger, t); err != nil {
				return err
			}
			if hasFieldFlag {
				if err := events.Append(ledger, events.Event{Event: "refined", TaskID: t.ID}); err != nil {
					return err
				}
			}
			for _, e := range refEvents {
				if err := events.Append(ledger, e); err != nil {
					return err
				}
			}
			return reportRefined(cmd, t, asJSON)
		},
	}
	c.Flags().StringVarP(&title, "title", "t", "", "Task title")
	c.Flags().StringVarP(&description, "description", "d", "", "Task description (stored under ## Description)")
	c.Flags().StringVar(&taskType, "type", "", "Task type")
	c.Flags().StringVarP(&priority, "priority", "p", "", "Task priority (l|low, m|medium, h|high)")
	c.Flags().StringArrayVar(&addRefs, "add-ref", nil, "Add a reference (repeatable; idempotent)")
	c.Flags().StringArrayVar(&removeRefs, "remove-ref", nil, "Remove a reference (repeatable; idempotent)")
	c.Flags().BoolVar(&editMode, "edit", false, "Open $VISUAL or $EDITOR to edit task fields")
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	return c
}

func refineWithEditor(cmd *cobra.Command, ledger, taskID string, asJSON bool) error {
	t, err := store.Read(ledger, taskID)
	if errors.Is(err, store.ErrTaskNotFound) {
		return NewExitError(3, "task %s not found", taskID)
	}
	if err != nil {
		return err
	}

	original := renderRefineEditBuffer(t)
	edited, err := editRefineBuffer(original)
	if err != nil {
		return err
	}
	if edited == original {
		if asJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(t)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "No changes for %s\n", t.ID)
		return nil
	}

	update, err := parseRefineEditBuffer(edited)
	if err != nil {
		return err
	}

	release, err := acquireLock(ledger)
	if err != nil {
		return err
	}
	defer release()

	t, err = store.Read(ledger, taskID)
	if errors.Is(err, store.ErrTaskNotFound) {
		return NewExitError(3, "task %s not found", taskID)
	}
	if err != nil {
		return err
	}
	if t.Title == update.title && t.Priority == update.priority && t.Type == update.taskType && t.Body == update.body {
		if asJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(t)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "No changes for %s\n", t.ID)
		return nil
	}

	t.Title = update.title
	t.Priority = update.priority
	t.Type = update.taskType
	t.Body = update.body
	return writeRefinedTask(cmd, ledger, t, asJSON)
}

func writeRefinedTask(cmd *cobra.Command, ledger string, t *task.Task, asJSON bool) error {
	t.UpdatedAt = time.Now().UTC().Truncate(time.Second)
	if err := store.Write(ledger, t); err != nil {
		return err
	}
	if err := events.Append(ledger, events.Event{
		Event:  "refined",
		TaskID: t.ID,
	}); err != nil {
		return err
	}

	return reportRefined(cmd, t, asJSON)
}

// applyRefMutations adds and removes references in place, returning one event
// per *actual* change. Adding an existing reference or removing a missing one
// is a no-op and produces no event (mirrors `dep add/remove` idempotency).
func applyRefMutations(t *task.Task, add, remove []string) []events.Event {
	var evs []events.Event
	for _, r := range add {
		if containsString(t.References, r) {
			continue
		}
		t.References = append(t.References, r)
		evs = append(evs, events.Event{Event: "reference_added", TaskID: t.ID, Value: r})
	}
	for _, r := range remove {
		if !containsString(t.References, r) {
			continue
		}
		t.References = removeString(t.References, r)
		evs = append(evs, events.Event{Event: "reference_removed", TaskID: t.ID, Value: r})
	}
	return evs
}

func containsString(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// removeString returns list with the first occurrence of v removed, preserving
// order. The result is non-nil so References stays an empty array, not null.
func removeString(list []string, v string) []string {
	out := make([]string, 0, len(list))
	removed := false
	for _, s := range list {
		if !removed && s == v {
			removed = true
			continue
		}
		out = append(out, s)
	}
	return out
}

func reportRefined(cmd *cobra.Command, t *task.Task, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(t)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Refined task %s\n", t.ID)
	return nil
}

func reportRefineNoChange(cmd *cobra.Command, t *task.Task, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(t)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "No changes for %s\n", t.ID)
	return nil
}

func renderRefineEditBuffer(t *task.Task) string {
	var b strings.Builder
	fmt.Fprintf(&b, "title: %s\n", t.Title)
	fmt.Fprintf(&b, "priority: %s\n", t.Priority)
	fmt.Fprintf(&b, "type: %s\n\n", t.Type)
	if t.Body != "" {
		b.WriteString(strings.TrimRight(t.Body, "\n"))
		b.WriteByte('\n')
	}
	return b.String()
}

type refineEditUpdate struct {
	title    string
	priority string
	taskType string
	body     string
}

func parseRefineEditBuffer(buffer string) (refineEditUpdate, error) {
	header, body, _ := strings.Cut(buffer, "\n\n")
	update := refineEditUpdate{}
	seen := map[string]bool{}
	for _, line := range strings.Split(header, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return update, NewExitError(2, "invalid editor buffer line %q: expected key: value", line)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		seen[key] = true
		switch key {
		case "title":
			update.title = value
		case "priority":
			normalized, err := normalizePriority(value)
			if err != nil {
				return update, err
			}
			update.priority = normalized
		case "type":
			update.taskType = value
		default:
			return update, NewExitError(2, "unknown editor buffer field %q", key)
		}
	}
	if !seen["title"] {
		return update, NewExitError(2, "editor buffer missing title")
	}
	if !seen["priority"] {
		return update, NewExitError(2, "editor buffer missing priority")
	}
	if !seen["type"] {
		return update, NewExitError(2, "editor buffer missing type")
	}
	update.body = strings.TrimRight(body, "\n")
	if update.body != "" {
		update.body += "\n"
	}
	return update, nil
}

func editRefineBuffer(initial string) (string, error) {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		return "", NewExitError(2, "no editor is configured: set VISUAL or EDITOR, or pass refine fields as flags")
	}

	f, err := os.CreateTemp("", "tl-refine-*.md")
	if err != nil {
		return "", err
	}
	path := f.Name()
	defer os.Remove(path)
	if _, err := f.WriteString(initial); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}

	command := exec.Command("sh", "-c", editor+" "+shellQuote(path))
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
