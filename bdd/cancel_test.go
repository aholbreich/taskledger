package bdd

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// --- cancel.feature support -----------------------------------------------

func initializeCancelSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^the command reports the task is already cancelled$`, w.outputReportsTaskAlreadyCancelled)
}

func (w *world) outputReportsTaskAlreadyCancelled() error {
	combined := w.stdout.String() + w.stderr.String()
	if !strings.Contains(combined, "already cancelled") {
		return fmt.Errorf("expected output to report task already cancelled; got: %s", combined)
	}
	return nil
}
