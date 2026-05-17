package bdd

import (
	"fmt"
	"github.com/cucumber/godog"
	"io"
	"os"
	"strings"

	"github.com/aholbreich/taskledger/cmd"
	"path/filepath"
)

// --- create.feature support -----------------------------------------------

func initializeCreateSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^a new task with title "([^"]*)" exists$`, w.newTaskWithTitleExists)
	ctx.Step(`^the new task has status "([^"]*)"$`, w.newTaskHasStatus)
	ctx.Step(`^the new task has type "([^"]*)"$`, w.newTaskHasType)
	ctx.Step(`^the new task has priority "([^"]*)"$`, w.newTaskHasPriority)
	ctx.Step(`^the new task has tags "([^"]*)" and "([^"]*)"$`, w.newTaskHasTwoTags)
	ctx.Step(`^the new task has no dependencies$`, w.newTaskHasNoDependencies)
	ctx.Step(`^the new task description is "([^"]*)"$`, w.newTaskDescriptionIs)
	ctx.Step(`^an event "([^"]*)" is recorded for the new task$`, w.eventRecordedForNewTask)
	ctx.Step(`^the JSON output body contains "([^"]*)"$`, w.jsonBodyContains)
	ctx.Step(`^the JSON output contains the new task identifier$`, w.jsonContainsTaskIdentifier)
	ctx.Step(`^the JSON output contains title "([^"]*)"$`, w.jsonStringField("title"))
	ctx.Step(`^the JSON output contains status "([^"]*)"$`, w.jsonStringField("status"))
	ctx.Step(`^the output reports that the priority is invalid$`, w.outputReportsPriorityInvalid)
}

func (w *world) ledgerInitialized() error {
	root := cmd.NewRootCmd()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"init"})
	return root.Execute()
}

func (w *world) noTasksExist() error {
	entries, err := os.ReadDir(filepath.Join(".taskledger", "tasks"))
	if err != nil {
		return fmt.Errorf("tasks folder missing: %w", err)
	}
	if len(entries) != 0 {
		return fmt.Errorf("expected no tasks, found %d", len(entries))
	}
	return nil
}

func (w *world) newTaskWithTitleExists(title string) error {
	t, err := loadOnlyTask()
	if err != nil {
		return err
	}
	if t.Title != title {
		return fmt.Errorf("task title is %q, expected %q", t.Title, title)
	}
	return nil
}

func (w *world) newTaskHasStatus(status string) error {
	t, err := loadOnlyTask()
	if err != nil {
		return err
	}
	if t.Status != status {
		return fmt.Errorf("task status is %q, expected %q", t.Status, status)
	}
	return nil
}

func (w *world) newTaskHasType(taskType string) error {
	t, err := loadOnlyTask()
	if err != nil {
		return err
	}
	if t.Type != taskType {
		return fmt.Errorf("task type is %q, expected %q", t.Type, taskType)
	}
	return nil
}

func (w *world) newTaskHasPriority(priority string) error {
	t, err := loadOnlyTask()
	if err != nil {
		return err
	}
	if t.Priority != priority {
		return fmt.Errorf("task priority is %q, expected %q", t.Priority, priority)
	}
	return nil
}

func (w *world) newTaskHasTwoTags(a, b string) error {
	t, err := loadOnlyTask()
	if err != nil {
		return err
	}
	have := map[string]bool{}
	for _, tg := range t.Tags {
		have[tg] = true
	}
	if !have[a] || !have[b] {
		return fmt.Errorf("task tags %v missing %q or %q", t.Tags, a, b)
	}
	return nil
}

func (w *world) newTaskHasNoDependencies() error {
	t, err := loadOnlyTask()
	if err != nil {
		return err
	}
	if len(t.DependsOn) != 0 {
		return fmt.Errorf("task has dependencies %v, expected none", t.DependsOn)
	}
	return nil
}

func (w *world) newTaskDescriptionIs(desc string) error {
	t, err := loadOnlyTask()
	if err != nil {
		return err
	}
	got := strings.TrimSpace(extractMarkdownSection(t.Body, "Description"))
	if got != desc {
		return fmt.Errorf("task description = %q, expected %q (body: %q)", got, desc, t.Body)
	}
	return nil
}

func (w *world) outputReportsPriorityInvalid() error {
	combined := strings.ToLower(w.stdout.String() + w.stderr.String())
	if w.cmdErr != nil {
		combined += "\n" + strings.ToLower(w.cmdErr.Error())
	}
	if !strings.Contains(combined, "invalid priority") {
		return fmt.Errorf("output does not report invalid priority; got:\n%s", combined)
	}
	return nil
}

func (w *world) jsonBodyContains(needle string) error {
	body, err := w.jsonStringValue("body")
	if err != nil {
		return err
	}
	if !strings.Contains(body, needle) {
		return fmt.Errorf("JSON body %q does not contain %q", body, needle)
	}
	return nil
}

// extractMarkdownSection returns the content under a "## <heading>" line up to
// the next "## " heading or end of body.
func extractMarkdownSection(body, heading string) string {
	needle := "## " + heading
	idx := strings.Index(body, needle)
	if idx < 0 {
		return ""
	}
	rest := body[idx+len(needle):]
	if i := strings.Index(rest, "\n"); i >= 0 {
		rest = rest[i+1:]
	} else {
		return ""
	}
	if next := strings.Index(rest, "\n## "); next >= 0 {
		rest = rest[:next]
	}
	return rest
}

func (w *world) eventRecordedForNewTask(eventName string) error {
	t, err := loadOnlyTask()
	if err != nil {
		return err
	}
	return assertEventRecorded(eventName, t.ID)
}

func (w *world) eventRecordedFor(eventName, taskID string) error {
	return assertEventRecorded(eventName, taskID)
}
