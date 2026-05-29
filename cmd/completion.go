package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aholbreich/tl/internal/store"
)

// completionDirective is the directive shared across completion paths.
// NoFileComp suppresses the shell's filename fallback; KeepOrder preserves
// the priority-then-ID order returned by store.List so the menu reads the
// same as `tl list`.
const completionDirective = cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder

// completeTaskIDs is a cobra ValidArgsFunction / flag completion func that
// suggests active task IDs from the current ledger. Closed tasks (status
// done/cancelled) are always excluded — the menu would otherwise grow
// unbounded with archived work. To inspect a closed task with `tl show` or
// `tl history`, type the full ID by hand.
//
// Suggestions are emitted as "id\t[status] Title" so zsh/fish display the
// status alongside the ID; bash shows the ID only. Failures (missing
// ledger, unreadable tasks) yield zero suggestions silently — completion
// must never spam stderr or block input.
func completeTaskIDs(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, completionDirective
	}
	ledger, err := store.LedgerDir(wd)
	if err != nil {
		return nil, completionDirective
	}
	tasks, err := store.List(ledger)
	if err != nil {
		return nil, completionDirective
	}

	// Emit canonical IDs (`task-xxx`) when the user has typed nothing or a
	// `task-` prefix; emit bare short codes (`xxx`) when the user has started
	// with something else, so `abc<TAB>` completes to `abc123` (which the
	// commands then accept via store.NormalizeID).
	bare := toComplete != "" && !strings.HasPrefix(toComplete, "task-") &&
		!strings.HasPrefix("task-", toComplete)

	suggestions := make([]string, 0, len(tasks))
	for _, t := range tasks {
		if t.Status == "done" || t.Status == "cancelled" {
			continue
		}
		id := t.ID
		if bare {
			id = strings.TrimPrefix(t.ID, "task-")
		}
		if toComplete != "" && !strings.HasPrefix(id, toComplete) {
			continue
		}
		description := "[" + t.Status + "]"
		if title := strings.TrimSpace(t.Title); title != "" {
			description += " " + title
		}
		suggestions = append(suggestions, id+"\t"+description)
	}
	return suggestions, completionDirective
}
