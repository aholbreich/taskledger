package bdd

import (
	"fmt"
	"github.com/cucumber/godog"
)

// --- stale.feature support ------------------------------------------------

func initializeStaleSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^the stale output contains "([^"]*)"$`, w.staleOutputContains)
	ctx.Step(`^the stale output does not contain "([^"]*)"$`, w.staleOutputDoesNotContain)
	ctx.Step(`^the JSON output is an array containing a stale claim for "([^"]*)"$`, w.jsonArrayContainsStaleClaim)
}

func (w *world) staleOutputContains(id string) error {
	if _, ok := lineContaining(w.stdout.String(), id); !ok {
		return fmt.Errorf("stale output does not contain %q; got:\n%s", id, w.stdout.String())
	}
	return nil
}

func (w *world) staleOutputDoesNotContain(id string) error {
	if line, ok := lineContaining(w.stdout.String(), id); ok {
		return fmt.Errorf("stale output unexpectedly contains %q in line: %s", id, line)
	}
	return nil
}

func (w *world) jsonArrayContainsStaleClaim(id string) error {
	tasks, err := w.jsonTaskArray()
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if t.ID == id {
			if t.Claim.Actor == nil {
				return fmt.Errorf("task %s has no claim actor", id)
			}
			return nil
		}
	}
	return fmt.Errorf("JSON array does not contain task %q; got: %s", id, w.stdout.String())
}
