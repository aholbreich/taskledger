package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aholbreich/tl/internal/doctor"
)

func newDoctorCmd() *cobra.Command {
	var (
		asJSON bool
		fix    bool
		force  bool
	)
	c := &cobra.Command{
		Use:   "doctor",
		Short: "Scan the ledger for integrity issues (optionally repair them)",
		Long: "Scan task files, the event journal, config, and the filesystem for " +
			"structural and data-integrity problems.\n\n" +
			"doctor is diagnostic: it always exits 0 when it can read the ledger, " +
			"regardless of what it finds.\n\n" +
			"Use --fix to repair issues that can be safely auto-fixed. Add --force " +
			"to also apply destructive repairs (e.g. removing orphaned event lines " +
			"from the journal).",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ledger, err := requireLedger()
			if err != nil {
				return err
			}
			if fix {
				return runDoctorFix(cmd, ledger, asJSON, force)
			}
			if force {
				return fmt.Errorf("--force requires --fix")
			}
			return runDoctorReport(cmd, ledger, asJSON)
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	c.Flags().BoolVar(&fix, "fix", false, "Repair issues that can be safely auto-fixed")
	c.Flags().BoolVar(&force, "force", false, "Allow destructive repairs (requires --fix)")
	return c
}

func runDoctorReport(cmd *cobra.Command, ledger string, asJSON bool) error {
	diags, err := doctor.Diagnose(ledger)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()

	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(diags)
	}

	if len(diags) == 0 {
		fmt.Fprintln(out, "No issues found.")
		return nil
	}
	color := commandColorEnabled(cmd)
	fmt.Fprintf(out, "Found %d issue(s):\n\n", len(diags))
	for _, d := range diags {
		fmt.Fprintln(out, formatDiagnostic(color, d))
	}
	return nil
}

func runDoctorFix(cmd *cobra.Command, ledger string, asJSON, force bool) error {
	release, err := acquireLock(ledger)
	if err != nil {
		return err
	}
	defer release()

	applied, unfixable, err := doctor.Fix(ledger, force)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()

	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if unfixable == nil {
			unfixable = []doctor.Diagnostic{}
		}
		return enc.Encode(unfixable)
	}

	if len(applied) == 0 && len(unfixable) == 0 {
		fmt.Fprintln(out, "No issues found.")
		return nil
	}
	color := commandColorEnabled(cmd)
	for _, r := range applied {
		fmt.Fprintln(out, formatRepair(color, r))
	}
	for _, d := range unfixable {
		fmt.Fprintf(out, "%s (not fixable)\n", formatDiagnostic(color, d))
	}
	return nil
}

func formatDiagnostic(color bool, d doctor.Diagnostic) string {
	tag := colorSeverity(color, d.Severity, "["+d.Severity+"]")
	if d.TaskID != "" {
		return fmt.Sprintf("%s %s %s: %s", tag, d.Category, d.TaskID, d.Message)
	}
	return fmt.Sprintf("%s %s: %s", tag, d.Category, d.Message)
}

func formatRepair(color bool, r doctor.Repair) string {
	tag := colorRepairVerb(color, "["+r.Verb+"]")
	d := r.Diagnostic
	if d.TaskID != "" {
		return fmt.Sprintf("%s %s %s: %s", tag, d.Category, d.TaskID, d.Message)
	}
	return fmt.Sprintf("%s %s: %s", tag, d.Category, d.Message)
}
