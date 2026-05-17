package cmd

import (
	"os"

	"github.com/spf13/cobra"

	internalcolor "github.com/aholbreich/taskledger/internal/color"
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
	var colorMode string
	root := &cobra.Command{
		Use:           "tl",
		Short:         "TaskLedger — a Git-native task ledger for humans and AI coding agents",
		Version:       rootVersion,
		SilenceUsage:  true,
		SilenceErrors: false,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return internalcolor.ValidateMode(colorMode)
		},
	}
	root.PersistentFlags().StringVar(&colorMode, "color", internalcolor.ModeAuto, "When to use ANSI color (auto|never|always)")
	root.AddCommand(newInitCmd())
	root.AddCommand(newCreateCmd())
	root.AddCommand(newShowCmd())
	root.AddCommand(newListCmd())
	root.AddCommand(newReadyCmd())
	root.AddCommand(newNoteCmd())
	root.AddCommand(newDepCmd())
	root.AddCommand(newClaimCmd())
	root.AddCommand(newCloseCmd())
	root.AddCommand(newAgentsCmd())
	root.AddCommand(newReleaseCmd())
	root.AddCommand(newStaleCmd())
	root.AddCommand(newHistoryCmd())
	root.AddCommand(newBlockCmd())
	root.AddCommand(newCancelCmd())
	root.AddCommand(newUnblockCmd())
	root.AddCommand(newPendingCmd())
	root.AddCommand(newResolveCmd())
	return root
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(ErrorExitCode(err))
	}
}
