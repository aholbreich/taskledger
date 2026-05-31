package bdd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// --- remove.feature support -----------------------------------------------

func initializeRemoveSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^an event "([^"]*)" is recorded for "([^"]*)" with value "([^"]*)"$`, w.eventRecordedForTaskWithValue)
	ctx.Step(`^the task file for "([^"]*)" does not exist$`, w.taskFileDoesNotExist)
	ctx.Step(`^the task file for "([^"]*)" exists$`, w.taskFileExists)
	ctx.Step(`^the output reports that force is required$`, w.outputReportsForceRequired)
}

func (w *world) taskFileDoesNotExist(id string) error {
	_, err := os.Stat(filepath.Join(".tl", "tasks", id+".md"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("task file for %s still exists", id)
}

func (w *world) taskFileExists(id string) error {
	_, err := os.Stat(filepath.Join(".tl", "tasks", id+".md"))
	if err != nil {
		return err
	}
	return nil
}

func (w *world) outputReportsForceRequired() error {
	return w.outputContainsAll("use --force")
}
