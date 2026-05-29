package cmd

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// completeTaskIDs must never error or print when invoked outside an
// initialized ledger — shell completion has to fail silently.
func TestCompleteTaskIDsReturnsNoSuggestionsWithoutLedger(t *testing.T) {
	tmp := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	suggestions, directive := completeTaskIDs(nil, nil, "")
	if len(suggestions) != 0 {
		t.Fatalf("expected no suggestions outside a ledger, got %v", suggestions)
	}
	want := cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	if directive != want {
		t.Fatalf("expected NoFileComp|KeepOrder (%d), got %d", want, directive)
	}
}
