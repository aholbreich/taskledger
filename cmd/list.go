package cmd

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/aholbreich/taskledger/internal/store"
)

func newListCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "list",
		Short: "List tasks in the ledger",
		RunE: func(cmd *cobra.Command, args []string) error {
			ledger, err := requireLedger()
			if err != nil {
				return err
			}
			tasks, err := store.List(ledger)
			if err != nil {
				return err
			}

			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(tasks)
			}

			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tStatus\tPriority\tTitle")
			for _, t := range tasks {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", t.ID, t.Status, t.Priority, t.Title)
			}
			return tw.Flush()
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	return c
}
