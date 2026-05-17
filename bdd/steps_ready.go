package bdd

import (
	"fmt"
	"github.com/cucumber/godog"
	"strings"
	"time"

	"github.com/aholbreich/taskledger/internal/task"
)

// --- ready.feature support ------------------------------------------------

func initializeReadySteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^a task "([^"]*)" with status "([^"]*)" and no dependencies$`, w.taskWithStatusAndNoDeps)
	ctx.Step(`^a task "([^"]*)" with status "([^"]*)" and tag "([^"]*)"$`, w.taskWithStatusAndTag)
	ctx.Step(`^a task "([^"]*)" titled "([^"]*)" with status "([^"]*)" and no dependencies$`, w.taskTitledWithStatusAndNoDeps)
	ctx.Step(`^a task "([^"]*)" with an expired claim by "([^"]*)"$`, w.taskWithExpiredClaim)
	ctx.Step(`^the ready output contains "([^"]*)"$`, w.readyOutputContains)
	ctx.Step(`^the ready output does not contain "([^"]*)"$`, w.readyOutputDoesNotContain)
	ctx.Step(`^the JSON output is an array containing a task with identifier "([^"]*)"$`, w.jsonArrayContainsTaskID)
	ctx.Step(`^the JSON output contains a priority for "([^"]*)"$`, w.jsonArrayTaskHasPriority)
}

func (w *world) taskWithStatusAndNoDeps(id, status string) error {
	return w.taskWithStatus(id, status)
}

func (w *world) taskWithStatusAndTag(id, status, tag string) error {
	return writeFixtureTask(&task.Task{
		ID:        id,
		Title:     id,
		Status:    status,
		Priority:  "medium",
		CreatedAt: fixtureTime,
		UpdatedAt: fixtureTime,
		CreatedBy: "human",
		DependsOn: []string{},
		Tags:      []string{tag},
	})
}

func (w *world) taskTitledWithStatusAndNoDeps(id, title, status string) error {
	return w.taskWithTitleAndStatus(id, title, status)
}

func (w *world) taskWithExpiredClaim(id, actor string) error {
	now := time.Now().UTC().Truncate(time.Second)
	expired := now.Add(-1 * time.Hour)
	a := actor
	return writeFixtureTask(&task.Task{
		ID:        id,
		Title:     id,
		Status:    "in_progress",
		Priority:  "medium",
		CreatedAt: fixtureTime,
		UpdatedAt: fixtureTime,
		CreatedBy: "human",
		DependsOn: []string{},
		Tags:      []string{},
		Claim: task.Claim{
			Actor:       &a,
			ClaimedAt:   &expired,
			ExpiresAt:   &expired,
			HeartbeatAt: &expired,
		},
	})
}

func (w *world) readyOutputContains(id string) error {
	combined := w.stdout.String()
	if !strings.Contains(combined, id) {
		return fmt.Errorf("ready output does not contain %q; got:\n%s", id, combined)
	}
	return nil
}

func (w *world) readyOutputDoesNotContain(id string) error {
	combined := w.stdout.String()
	if strings.Contains(combined, id) {
		return fmt.Errorf("ready output contains %q but should not; got:\n%s", id, combined)
	}
	return nil
}

func (w *world) jsonArrayTaskHasPriority(id string) error {
	tasks, err := w.jsonTaskArray()
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if t.ID == id {
			if t.Priority == "" {
				return fmt.Errorf("task %s in JSON array has no priority", id)
			}
			return nil
		}
	}
	return fmt.Errorf("JSON array does not contain task %q", id)
}
