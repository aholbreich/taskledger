package bdd

import (
	"fmt"
	"strings"

	"github.com/aholbreich/tl/internal/task"
	"github.com/cucumber/godog"
)

// --- resolve.feature support ----------------------------------------------

func initializeResolveSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^"([^"]*)" has the question "([^"]*)"$`, w.taskHasQuestion)
	ctx.Step(`^the command reports the task is not pending_human$`, w.outputReportsTaskNotPending)
}

func (w *world) taskHasQuestion(id, question string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	// Ensure the Pending struct is set up (this step acts as a Given setup).
	if t.Pending == nil {
		t.Pending = &task.Pending{}
	}
	t.Pending.Question = question
	t.Pending.Requester = "claude-code:frontend"
	return writeFixtureTask(t)
}

func (w *world) outputReportsTaskNotPending() error {
	combined := w.stdout.String() + w.stderr.String()
	if !strings.Contains(combined, "not pending_human") {
		return fmt.Errorf("expected output to report task not pending_human; got: %s", combined)
	}
	return nil
}
