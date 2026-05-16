package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/aholbreich/taskledger/internal/repo"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a TaskLedger ledger in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			ledger, err := repo.Init(wd)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Initialized TaskLedger ledger at %s\n", ledger)
			return nil
		},
	}
}
