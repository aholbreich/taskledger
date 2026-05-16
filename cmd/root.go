package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootVersion = "dev"

// SetVersion records the version string baked in by main; cobra surfaces it
// via `tl --version` and `tl version`.
func SetVersion(v string) {
	if v != "" {
		rootVersion = v
	}
}

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "tl",
		Short:         "TaskLedger — a Git-native task ledger for humans and AI coding agents",
		Version:       rootVersion,
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	root.AddCommand(newInitCmd())
	root.AddCommand(newCreateCmd())
	root.AddCommand(newShowCmd())
	root.AddCommand(newListCmd())
	return root
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(ErrorExitCode(err))
	}
}
