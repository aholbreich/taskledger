package cmd

import (
	"os"

	"github.com/spf13/cobra"

	internalcolor "github.com/aholbreich/tl/internal/color"
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
	var (
		colorMode   string
		title       string
		description string
		priority    string
		tags        []string
		refs        []string
		asJSON      bool
	)
	var root *cobra.Command
	root = &cobra.Command{
		Use:           "tl",
		Short:         "tl — a Git-native task ledger for humans and AI coding agents",
		Version:       rootVersion,
		SilenceUsage:  true,
		SilenceErrors: false,
		Args:          cobra.ArbitraryArgs,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return internalcolor.ValidateMode(colorMode)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Bare `tl "Title" [flags]` is shorthand for `tl create "Title" [flags]`.
			if len(args) == 0 || cmd != root {
				return cmd.Help()
			}
			createArgs := []string{args[0]}
			if title != "" {
				createArgs = append(createArgs, "--title", title)
			}
			if description != "" {
				createArgs = append(createArgs, "--description", description)
			}
			if priority != "" {
				createArgs = append(createArgs, "--priority", priority)
			}
			for _, t := range tags {
				createArgs = append(createArgs, "--tag", t)
			}
			for _, r := range refs {
				createArgs = append(createArgs, "--ref", r)
			}
			if asJSON {
				createArgs = append(createArgs, "--json")
			}
			createCmd := newCreateCmd()
			createCmd.SetArgs(createArgs)
			createCmd.SetOut(cmd.OutOrStdout())
			createCmd.SetErr(cmd.ErrOrStderr())
			return createCmd.Execute()
		},
	}
	root.PersistentFlags().StringVar(&colorMode, "color", internalcolor.ModeAuto, "When to use ANSI color (auto|never|always)")
	root.Flags().StringVarP(&description, "description", "d", "", "Task description")
	root.Flags().StringVar(&priority, "priority", "", "Task priority (high|medium|low)")
	root.Flags().StringArrayVar(&tags, "tag", nil, "Tag for the task (repeatable)")
	root.Flags().StringArrayVar(&refs, "ref", nil, "Reference for the task: file path, URL, ticket ID, … (repeatable)")
	root.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	root.Flags().StringVarP(&title, "title", "t", "", "Task title (if not positional)")
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
	root.AddCommand(newRefineCmd())
	root.AddCommand(newCompletionCmd())
	return root
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(ErrorExitCode(err))
	}
}
