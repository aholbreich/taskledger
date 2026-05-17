package bdd

import (
	"fmt"
	"github.com/cucumber/godog"
	"strings"
)

// --- ledger-required.feature support --------------------------------------

func initializeLedgerRequiredSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^the output reports that TaskLedger is not initialized$`, w.outputReportsLedgerNotInitialized)
	ctx.Step(`^the output suggests running "([^"]*)"$`, w.outputSuggestsRunning)
}

func (w *world) outputReportsLedgerNotInitialized() error {
	combined := strings.ToLower(w.stdout.String() + w.stderr.String())
	if w.cmdErr != nil {
		combined += "\n" + strings.ToLower(w.cmdErr.Error())
	}
	if !strings.Contains(combined, "taskledger") || !strings.Contains(combined, "not initialized") {
		return fmt.Errorf("output does not report TaskLedger as not initialized; got:\n%s", combined)
	}
	return nil
}

func (w *world) outputSuggestsRunning(command string) error {
	combined := w.stdout.String() + w.stderr.String()
	if w.cmdErr != nil {
		combined += "\n" + w.cmdErr.Error()
	}
	if !strings.Contains(combined, command) {
		return fmt.Errorf("output does not suggest running %q; got:\n%s", command, combined)
	}
	return nil
}
