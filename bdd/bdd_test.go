// Package bdd runs the project's Gherkin features under godog as the BDD
// acceptance test suite. Step definitions invoke the cobra CLI in-process so
// scenarios exercise the full command surface without spawning subprocesses.
package bdd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"github.com/aholbreich/taskledger/cmd"
	"github.com/aholbreich/taskledger/internal/events"
	"github.com/aholbreich/taskledger/internal/task"
)

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features"},
			Tags:     "@implemented",
			TestingT: t,
			Strict:   true,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("godog suite failed")
	}
}

type world struct {
	tempDir      string
	origWd       string
	stdout       *bytes.Buffer
	stderr       *bytes.Buffer
	cmdErr       error
	envOverrides []string // env keys set during scenario, restored in After
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	w := &world{}

	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		td, err := os.MkdirTemp("", "tl-bdd-*")
		if err != nil {
			return ctx, err
		}
		owd, err := os.Getwd()
		if err != nil {
			return ctx, err
		}
		if err := os.Chdir(td); err != nil {
			return ctx, err
		}
		w.tempDir = td
		w.origWd = owd
		w.stdout = &bytes.Buffer{}
		w.stderr = &bytes.Buffer{}
		w.cmdErr = nil
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		_ = os.Chdir(w.origWd)
		_ = os.RemoveAll(w.tempDir)
		cmd.DetectedActor = cmd.DefaultDetectActor
		for _, key := range w.envOverrides {
			_ = os.Unsetenv(key)
		}
		w.envOverrides = nil
		return ctx, nil
	})

	// Background / shared preconditions
	ctx.Step(`^an initialized TaskLedger repository$`, w.ledgerInitialized)
	ctx.Step(`^no tasks exist$`, w.noTasksExist)

	// init.feature preconditions
	ctx.Step(`^the current directory has no TaskLedger ledger$`, w.currentDirHasNoLedger)
	ctx.Step(`^the current directory already has a TaskLedger ledger$`, w.currentDirAlreadyHasLedger)

	// Shared CLI invocation (handles both "developer" and "agent" actors).
	ctx.Step("^the (?:developer|agent) runs `tl (.+)`$", w.runTl)

	// init.feature outcomes
	ctx.Step(`^the directory contains a TaskLedger config file$`, w.dirContainsConfigFile)
	ctx.Step(`^the directory contains an empty tasks folder$`, w.dirContainsEmptyTasksFolder)
	ctx.Step(`^the directory contains an empty event journal$`, w.dirContainsEmptyEventJournal)
	ctx.Step(`^the command reports that the ledger already exists$`, w.cmdReportsAlreadyExists)
	ctx.Step(`^the existing config file is unchanged$`, w.existingConfigUnchanged)

	// create.feature outcomes
	ctx.Step(`^a new task with title "([^"]*)" exists$`, w.newTaskWithTitleExists)
	ctx.Step(`^the new task has status "([^"]*)"$`, w.newTaskHasStatus)
	ctx.Step(`^the new task has type "([^"]*)"$`, w.newTaskHasType)
	ctx.Step(`^the new task has priority "([^"]*)"$`, w.newTaskHasPriority)
	ctx.Step(`^the new task has tags "([^"]*)" and "([^"]*)"$`, w.newTaskHasTwoTags)
	ctx.Step(`^the new task has no dependencies$`, w.newTaskHasNoDependencies)
	ctx.Step(`^the new task description is "([^"]*)"$`, w.newTaskDescriptionIs)
	ctx.Step(`^an event "([^"]*)" is recorded for the new task$`, w.eventRecordedForNewTask)
	ctx.Step(`^the JSON output body contains "([^"]*)"$`, w.jsonBodyContains)
	ctx.Step(`^the output reports that the priority is invalid$`, w.outputReportsPriorityInvalid)

	// list.feature preconditions and outcomes
	ctx.Step(`^the following tasks exist:$`, w.followingTasksExist)
	ctx.Step(`^the output lists "([^"]*)"$`, w.outputListsTaskID)
	ctx.Step(`^the output lists "([^"]*)" with status "([^"]*)" and title "([^"]*)"$`, w.outputListsTask)
	ctx.Step(`^the output lists "([^"]*)" with status "([^"]*)", priority "([^"]*)", claimed by "([^"]*)", and title "([^"]*)"$`, w.outputListsTaskWithColumns)
	ctx.Step(`^the output does not list "([^"]*)"$`, w.outputDoesNotListTask)
	ctx.Step(`^the list output columns are exactly:$`, w.listOutputColumnsAreExactly)
	ctx.Step(`^the JSON output is an array of (\d+) tasks$`, w.jsonOutputIsArrayOfTasks)
	ctx.Step(`^the JSON output contains a task with identifier "([^"]*)"$`, w.jsonArrayContainsTaskID)
	ctx.Step(`^the JSON output does not contain a task with identifier "([^"]*)"$`, w.jsonArrayDoesNotContainTaskID)
	ctx.Step(`^the listed task identifiers appear in this order:$`, w.listedTaskIDsAppearInOrder)

	// show.feature preconditions and outcomes
	ctx.Step(`^a task "([^"]*)" titled "([^"]*)" with status "([^"]*)"$`, w.taskWithTitleAndStatus)
	ctx.Step(`^"([^"]*)" depends on "([^"]*)"$`, w.taskDependsOn)
	ctx.Step(`^"([^"]*)" does not depend on "([^"]*)"$`, w.taskDoesNotDependOn)
	ctx.Step(`^"([^"]*)" has a note from "([^"]*)" saying "([^"]*)"$`, w.taskHasNote)
	ctx.Step(`^no task with identifier "([^"]*)" exists$`, w.noTaskWithIdentifierExists)

	// dep-add.feature preconditions and outcomes
	ctx.Step(`^a task "([^"]*)" with no dependencies$`, w.taskWithNoDependencies)
	ctx.Step(`^a task "([^"]*)" exists$`, w.taskExists)
	ctx.Step(`^an event "([^"]*)" is recorded for "([^"]*)"$`, w.eventRecordedFor)
	ctx.Step(`^"([^"]*)" has no dependencies$`, w.taskHasNoDependencies)
	ctx.Step(`^the output contains identifier "([^"]*)"$`, w.outputContains)
	ctx.Step(`^the output contains title "([^"]*)"$`, w.outputContains)
	ctx.Step(`^the output contains status "([^"]*)"$`, w.outputContains)
	ctx.Step(`^the output contains dependency "([^"]*)"$`, w.outputContains)
	ctx.Step(`^the output contains the note from "([^"]*)"$`, w.outputContains)
	ctx.Step(`^the JSON output contains identifier "([^"]*)"$`, w.jsonStringField("id"))
	ctx.Step(`^the command exits with code (\d+)$`, w.commandExitsWithCode)
	ctx.Step(`^the output reports that "([^"]*)" was not found$`, w.outputReportsNotFound)

	// ledger-required.feature outcomes
	ctx.Step(`^the output reports that TaskLedger is not initialized$`, w.outputReportsLedgerNotInitialized)
	ctx.Step(`^the output suggests running "([^"]*)"$`, w.outputSuggestsRunning)

	// JSON output assertions on stdout
	ctx.Step(`^the JSON output contains the new task identifier$`, w.jsonContainsTaskIdentifier)
	ctx.Step(`^the JSON output contains title "([^"]*)"$`, w.jsonStringField("title"))
	ctx.Step(`^the JSON output contains status "([^"]*)"$`, w.jsonStringField("status"))

	// claim.feature preconditions
	ctx.Step(`^a ready task "([^"]*)" titled "([^"]*)"$`, w.readyTaskTitled)
	ctx.Step(`^a ready task "([^"]*)"$`, w.readyTask)
	ctx.Step(`^a task "([^"]*)" claimed by "([^"]*)" with an active lease$`, w.taskClaimedByWithActiveLease)
	ctx.Step(`^a task "([^"]*)" claimed by "([^"]*)"$`, w.taskClaimedByWithActiveLease)
	ctx.Step(`^a task "([^"]*)" with status "([^"]*)"$`, w.taskWithStatus)

	// claim.feature outcomes
	ctx.Step(`^"([^"]*)" is claimed by "([^"]*)"$`, w.taskIsClaimedBy)
	ctx.Step(`^"([^"]*)" is still claimed by "([^"]*)"$`, w.taskIsClaimedBy)
	ctx.Step(`^"([^"]*)" is no longer claimed by "([^"]*)"$`, w.taskIsNoLongerClaimedBy)
	ctx.Step(`^"([^"]*)" has status "([^"]*)"$`, w.taskHasSpecificStatus)
	ctx.Step(`^"([^"]*)" still has status "([^"]*)"$`, w.taskHasSpecificStatus)
	ctx.Step(`^"([^"]*)" has a non-empty claim expiry$`, w.taskHasNonEmptyClaimExpiry)
	ctx.Step(`^"([^"]*)" is not claimed$`, w.taskIsNotClaimed)
	ctx.Step(`^"([^"]*)" was closed by "([^"]*)"$`, w.taskWasClosedBy)
	ctx.Step(`^"([^"]*)" does not have status "([^"]*)"$`, w.taskDoesNotHaveStatus)
	ctx.Step(`^the command reports the task is blocked$`, w.outputReportsTaskBlocked)
	ctx.Step(`^the command reports the claim is held by a different actor$`, w.outputReportsClaimHeldByDifferentActor)
	ctx.Step(`^the command reports the task is already closed$`, w.outputReportsTaskAlreadyClosed)
	ctx.Step(`^the JSON output contains actor "([^"]*)"$`, w.jsonClaimActor)
	ctx.Step(`^the JSON output contains a claim expiry (\d+) minutes after the claim time$`, w.jsonClaimExpiryAfter)

	// actor.feature preconditions and outcomes
	ctx.Step(`^environment variable "([^"]*)" is "([^"]*)"$`, w.setEnv)
	ctx.Step(`^the detected agent is "([^"]*)"$`, w.setDetectedAgent)
	ctx.Step(`^the claim expiry for "([^"]*)" is extended$`, w.claimExpiryIsExtended)

	// ready.feature preconditions and outcomes
	ctx.Step(`^a task "([^"]*)" with status "([^"]*)" and no dependencies$`, w.taskWithStatusAndNoDeps)
	ctx.Step(`^a task "([^"]*)" with status "([^"]*)" and tag "([^"]*)"$`, w.taskWithStatusAndTag)
	ctx.Step(`^a task "([^"]*)" titled "([^"]*)" with status "([^"]*)" and no dependencies$`, w.taskTitledWithStatusAndNoDeps)
	ctx.Step(`^a task "([^"]*)" with an expired claim by "([^"]*)"$`, w.taskWithExpiredClaim)
	ctx.Step(`^the ready output contains "([^"]*)"$`, w.readyOutputContains)
	ctx.Step(`^the ready output does not contain "([^"]*)"$`, w.readyOutputDoesNotContain)
	ctx.Step(`^the JSON output is an array containing a task with identifier "([^"]*)"$`, w.jsonArrayContainsTaskID)
	ctx.Step(`^the JSON output contains a priority for "([^"]*)"$`, w.jsonArrayTaskHasPriority)

	// note.feature preconditions and outcomes
	ctx.Step(`^a task "([^"]*)" titled "([^"]*)"$`, w.taskTitled)
	ctx.Step(`^"([^"]*)" has a note from "([^"]*)"$`, w.taskHasNoteFrom)
	ctx.Step(`^the note contains the message "([^"]*)"$`, w.noteContainsMessage)
	ctx.Step(`^the note has a timestamp$`, w.noteHasTimestamp)

	// stale.feature preconditions and outcomes
	ctx.Step(`^the stale output contains "([^"]*)"$`, w.staleOutputContains)
	ctx.Step(`^the stale output does not contain "([^"]*)"$`, w.staleOutputDoesNotContain)
	ctx.Step(`^the JSON output is an array containing a stale claim for "([^"]*)"$`, w.jsonArrayContainsStaleClaim)

	// agents.feature preconditions and outcomes
	ctx.Step(`^the file "([^"]*)" exists with content "([^"]*)"$`, w.fileExistsWithContent)
	ctx.Step(`^the file "([^"]*)" still has content "([^"]*)"$`, w.fileStillHasContent)
	ctx.Step(`^the output contains a "([^"]*)" heading$`, w.outputContainsHeading)
	ctx.Step(`^the output describes the ready, claim, show, note, and close steps$`, w.outputDescribesWorkflowSteps)
	ctx.Step(`^the output formats task commands as Markdown code spans$`, w.outputFormatsCommandsAsMarkdownCodeSpans)
}

// --- init.feature support -------------------------------------------------

const sentinelConfig = "# existing config — DO NOT TOUCH\n"

func (w *world) currentDirHasNoLedger() error {
	if _, err := os.Stat(".taskledger"); err == nil {
		return fmt.Errorf(".taskledger already exists in fresh temp dir (setup bug)")
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (w *world) currentDirAlreadyHasLedger() error {
	if err := os.MkdirAll(filepath.Join(".taskledger", "tasks"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(".taskledger", "config.yaml"), []byte(sentinelConfig), 0o644)
}

func (w *world) dirContainsConfigFile() error {
	info, err := os.Stat(filepath.Join(".taskledger", "config.yaml"))
	if err != nil {
		return fmt.Errorf("config file missing: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("config file is empty")
	}
	return nil
}

func (w *world) dirContainsEmptyTasksFolder() error {
	entries, err := os.ReadDir(filepath.Join(".taskledger", "tasks"))
	if err != nil {
		return fmt.Errorf("tasks folder missing: %w", err)
	}
	if len(entries) != 0 {
		return fmt.Errorf("tasks folder is not empty (%d entries)", len(entries))
	}
	return nil
}

func (w *world) dirContainsEmptyEventJournal() error {
	info, err := os.Stat(filepath.Join(".taskledger", "events.jsonl"))
	if err != nil {
		return fmt.Errorf("events journal missing: %w", err)
	}
	if info.Size() != 0 {
		return fmt.Errorf("events journal is not empty (%d bytes)", info.Size())
	}
	return nil
}

func (w *world) cmdReportsAlreadyExists() error {
	combined := strings.ToLower(w.stdout.String() + w.stderr.String())
	if w.cmdErr != nil {
		combined += " " + strings.ToLower(w.cmdErr.Error())
	}
	if w.cmdErr == nil {
		return fmt.Errorf("expected init to fail, but it succeeded; output: %s", combined)
	}
	if !strings.Contains(combined, "already") {
		return fmt.Errorf("expected output to mention 'already', got: %s", combined)
	}
	return nil
}

func (w *world) existingConfigUnchanged() error {
	data, err := os.ReadFile(filepath.Join(".taskledger", "config.yaml"))
	if err != nil {
		return err
	}
	if string(data) != sentinelConfig {
		return fmt.Errorf("config file was modified; content: %q", string(data))
	}
	return nil
}

// --- create.feature support -----------------------------------------------

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

// --- list.feature support -------------------------------------------------

func (w *world) followingTasksExist(table *godog.Table) error {
	allowedColumns := map[string]bool{
		"id": true, "status": true, "priority": true, "claimed by": true, "title": true, "tags": true,
	}
	for _, header := range table.Rows[0].Cells {
		if !allowedColumns[header.Value] {
			return fmt.Errorf("unsupported task fixture column %q", header.Value)
		}
	}

	for rowIdx, row := range table.Rows[1:] {
		values := map[string]string{}
		for i, cell := range row.Cells {
			values[table.Rows[0].Cells[i].Value] = strings.TrimSpace(cell.Value)
		}
		id := values["id"]
		if id == "" {
			return fmt.Errorf("task row %d is missing id", rowIdx+1)
		}
		status := values["status"]
		if status == "" {
			status = "open"
		}
		priority := values["priority"]
		if priority == "" {
			priority = "medium"
		}
		title := values["title"]
		if title == "" {
			title = id
		}

		fixture := &task.Task{
			ID:        id,
			Title:     title,
			Status:    status,
			Priority:  priority,
			CreatedAt: fixtureTime,
			UpdatedAt: fixtureTime,
			CreatedBy: "human",
			DependsOn: []string{},
			Tags:      []string{},
		}
		if actor := values["claimed by"]; actor != "" {
			claimedAt := fixtureTime
			expiresAt := fixtureTime.Add(time.Hour)
			fixture.Claim = task.Claim{
				Actor:       &actor,
				ClaimedAt:   &claimedAt,
				ExpiresAt:   &expiresAt,
				HeartbeatAt: &claimedAt,
			}
		}
		if tagsVal := values["tags"]; tagsVal != "" {
			for _, tag := range strings.Split(tagsVal, ",") {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					fixture.Tags = append(fixture.Tags, tag)
				}
			}
		}
		if err := writeFixtureTask(fixture); err != nil {
			return err
		}
	}
	return nil
}

func (w *world) outputListsTaskID(id string) error {
	if _, ok := lineContaining(w.stdout.String(), id); !ok {
		return fmt.Errorf("output does not list %q; got:\n%s", id, w.stdout.String())
	}
	return nil
}

func (w *world) outputListsTask(id, status, title string) error {
	line, ok := lineContaining(w.stdout.String(), id)
	if !ok {
		return fmt.Errorf("output does not list %q; got:\n%s", id, w.stdout.String())
	}
	if !strings.Contains(line, status) {
		return fmt.Errorf("line for %q does not contain status %q: %s", id, status, line)
	}
	if !strings.Contains(line, title) {
		return fmt.Errorf("line for %q does not contain title %q: %s", id, title, line)
	}
	return nil
}

func (w *world) outputListsTaskWithColumns(id, status, priority, claimedBy, title string) error {
	line, ok := lineContaining(w.stdout.String(), id)
	if !ok {
		return fmt.Errorf("output does not list %q; got:\n%s", id, w.stdout.String())
	}
	columns := splitListLine(line)
	expected := []string{id, status, priority, claimedBy, title}
	if len(columns) != len(expected) {
		return fmt.Errorf("line for %q has columns %v, expected %v; line: %s", id, columns, expected, line)
	}
	for i := range expected {
		if columns[i] != expected[i] {
			return fmt.Errorf("line for %q column %d = %q, expected %q; columns: %v", id, i+1, columns[i], expected[i], columns)
		}
	}
	return nil
}

func (w *world) outputDoesNotListTask(id string) error {
	if line, ok := lineContaining(w.stdout.String(), id); ok {
		return fmt.Errorf("output unexpectedly lists %q in line: %s\nfull output:\n%s", id, line, w.stdout.String())
	}
	return nil
}

func (w *world) listOutputColumnsAreExactly(table *godog.Table) error {
	lines := nonEmptyLines(w.stdout.String())
	if len(lines) == 0 {
		return fmt.Errorf("list output is empty")
	}
	actual := splitListLine(lines[0])
	var expected []string
	for _, row := range table.Rows[1:] {
		expected = append(expected, row.Cells[0].Value)
	}
	if strings.Join(actual, "|") != strings.Join(expected, "|") {
		return fmt.Errorf("list columns = %v, expected %v; output:\n%s", actual, expected, w.stdout.String())
	}
	return nil
}

func (w *world) jsonOutputIsArrayOfTasks(count int) error {
	tasks, err := w.jsonTaskArray()
	if err != nil {
		return err
	}
	if len(tasks) != count {
		return fmt.Errorf("JSON array has %d tasks, expected %d; got: %s", len(tasks), count, w.stdout.String())
	}
	return nil
}

func (w *world) jsonArrayContainsTaskID(id string) error {
	tasks, err := w.jsonTaskArray()
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if t.ID == id {
			return nil
		}
	}
	return fmt.Errorf("JSON array does not contain task %q; got: %s", id, w.stdout.String())
}

func (w *world) jsonArrayDoesNotContainTaskID(id string) error {
	tasks, err := w.jsonTaskArray()
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if t.ID == id {
			return fmt.Errorf("JSON array unexpectedly contains task %q; got: %s", id, w.stdout.String())
		}
	}
	return nil
}

func (w *world) listedTaskIDsAppearInOrder(table *godog.Table) error {
	var expected []string
	for _, row := range table.Rows[1:] {
		expected = append(expected, row.Cells[0].Value)
	}
	var actual []string
	for _, line := range nonEmptyLines(w.stdout.String()) {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] == "ID" {
			continue
		}
		actual = append(actual, fields[0])
	}
	if strings.Join(actual, ",") != strings.Join(expected, ",") {
		return fmt.Errorf("listed task ids = %v, expected %v; output:\n%s", actual, expected, w.stdout.String())
	}
	return nil
}

func (w *world) jsonTaskArray() ([]task.Task, error) {
	var tasks []task.Task
	if err := json.Unmarshal(w.stdout.Bytes(), &tasks); err != nil {
		return nil, fmt.Errorf("stdout is not a JSON task array (%v); got: %s", err, w.stdout.String())
	}
	return tasks, nil
}

// --- show.feature support -------------------------------------------------

func (w *world) taskWithNoDependencies(id string) error {
	return writeFixtureTask(&task.Task{
		ID:        id,
		Title:     id,
		Status:    "open",
		Priority:  "medium",
		CreatedAt: fixtureTime,
		UpdatedAt: fixtureTime,
		CreatedBy: "human",
		DependsOn: []string{},
		Tags:      []string{},
	})
}

func (w *world) taskExists(id string) error {
	return w.taskWithNoDependencies(id)
}

func (w *world) taskWithTitleAndStatus(id, title, status string) error {
	return writeFixtureTask(&task.Task{
		ID:        id,
		Title:     title,
		Status:    status,
		Priority:  "medium",
		CreatedAt: fixtureTime,
		UpdatedAt: fixtureTime,
		CreatedBy: "human",
		DependsOn: []string{},
		Tags:      []string{},
	})
}

func (w *world) taskDependsOn(id, dependencyID string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	// If already present, treat as an assertion (Then step). Otherwise add
	// it (Given setup for scenarios like show.feature and dep-add.feature).
	for _, d := range t.DependsOn {
		if d == dependencyID {
			return nil
		}
	}
	t.DependsOn = append(t.DependsOn, dependencyID)
	return writeFixtureTask(t)
}

func (w *world) taskDoesNotDependOn(id, dep string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	for _, d := range t.DependsOn {
		if d == dep {
			return fmt.Errorf("task %s still depends on %s; depends_on: %v", id, dep, t.DependsOn)
		}
	}
	return nil
}

func (w *world) taskHasNote(id, actor, message string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	t.Body = strings.TrimRight(t.Body, "\n")
	if t.Body != "" {
		t.Body += "\n\n"
	}
	t.Body += fmt.Sprintf("## Notes\n\n### %s - %s\n\n%s\n", fixtureTime.Format(time.RFC3339), actor, message)
	return writeFixtureTask(t)
}

func (w *world) taskHasNoDependencies(id string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if len(t.DependsOn) != 0 {
		return fmt.Errorf("task %s has dependencies %v, expected none", id, t.DependsOn)
	}
	return nil
}

func (w *world) noTaskWithIdentifierExists(id string) error {
	if err := os.Remove(filepath.Join(".taskledger", "tasks", id+".md")); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (w *world) outputContains(needle string) error {
	combined := w.stdout.String() + w.stderr.String()
	if w.cmdErr != nil {
		combined += "\n" + w.cmdErr.Error()
	}
	if !strings.Contains(combined, needle) {
		return fmt.Errorf("output does not contain %q; got:\n%s", needle, combined)
	}
	return nil
}

func (w *world) outputContainsAll(needles ...string) error {
	combined := strings.ToLower(w.stdout.String() + w.stderr.String())
	if w.cmdErr != nil {
		combined += "\n" + strings.ToLower(w.cmdErr.Error())
	}
	for _, needle := range needles {
		if !strings.Contains(combined, strings.ToLower(needle)) {
			return fmt.Errorf("output does not contain %q; got:\n%s", needle, combined)
		}
	}
	return nil
}

func (w *world) commandExitsWithCode(expected int) error {
	got := 0
	if w.cmdErr != nil {
		got = 1
		var exitErr interface{ ExitCode() int }
		if errors.As(w.cmdErr, &exitErr) {
			got = exitErr.ExitCode()
		}
	}
	if got != expected {
		return fmt.Errorf("exit code = %d, expected %d; error: %v", got, expected, w.cmdErr)
	}
	return nil
}

func (w *world) outputReportsNotFound(id string) error {
	combined := strings.ToLower(w.stdout.String() + w.stderr.String())
	if w.cmdErr != nil {
		combined += "\n" + strings.ToLower(w.cmdErr.Error())
	}
	if !strings.Contains(combined, strings.ToLower(id)) || !strings.Contains(combined, "not found") {
		return fmt.Errorf("output does not report %q as not found; got:\n%s", id, combined)
	}
	return nil
}

// --- ledger-required.feature support --------------------------------------

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

// --- claim.feature support -----------------------------------------------

func (w *world) readyTaskTitled(id, title string) error {
	return writeFixtureTask(&task.Task{
		ID:        id,
		Title:     title,
		Status:    "open",
		Priority:  "medium",
		CreatedAt: fixtureTime,
		UpdatedAt: fixtureTime,
		CreatedBy: "human",
		DependsOn: []string{},
		Tags:      []string{},
	})
}

func (w *world) readyTask(id string) error {
	return w.readyTaskTitled(id, id)
}

func (w *world) taskClaimedByWithActiveLease(id, actor string) error {
	now := time.Now().UTC().Truncate(time.Second)
	expires := now.Add(1 * time.Hour)
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
			ClaimedAt:   &now,
			ExpiresAt:   &expires,
			HeartbeatAt: &now,
		},
	})
}

func (w *world) taskWithStatus(id, status string) error {
	t := &task.Task{
		ID:        id,
		Title:     id,
		Status:    status,
		Priority:  "medium",
		CreatedAt: fixtureTime,
		UpdatedAt: fixtureTime,
		CreatedBy: "human",
		DependsOn: []string{},
		Tags:      []string{},
	}
	return writeFixtureTask(t)
}

func (w *world) taskIsClaimedBy(id, actor string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Claim.Actor == nil {
		return fmt.Errorf("task %s has no claim", id)
	}
	if *t.Claim.Actor != actor {
		return fmt.Errorf("task %s claimed by %q, expected %q", id, *t.Claim.Actor, actor)
	}
	return nil
}

func (w *world) taskIsNoLongerClaimedBy(id, actor string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Claim.Actor != nil && *t.Claim.Actor == actor {
		return fmt.Errorf("task %s is still claimed by %q", id, actor)
	}
	return nil
}

func (w *world) taskHasSpecificStatus(id, status string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Status != status {
		return fmt.Errorf("task %s status is %q, expected %q", id, t.Status, status)
	}
	return nil
}

func (w *world) taskHasNonEmptyClaimExpiry(id string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Claim.ExpiresAt == nil {
		return fmt.Errorf("task %s has no claim expiry", id)
	}
	if t.Claim.ExpiresAt.IsZero() {
		return fmt.Errorf("task %s claim expiry is zero", id)
	}
	return nil
}

func (w *world) taskIsNotClaimed(id string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Claim.Actor != nil {
		return fmt.Errorf("task %s is claimed by %q, expected none", id, *t.Claim.Actor)
	}
	return nil
}

func (w *world) taskWasClosedBy(id, actor string) error {
	return assertEventRecordedBy("closed", id, actor)
}

func (w *world) taskDoesNotHaveStatus(id, status string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Status == status {
		return fmt.Errorf("task %s status is %q, expected a different status", id, status)
	}
	return nil
}

func (w *world) outputReportsTaskBlocked() error {
	return w.outputContainsAll("blocked")
}

func (w *world) outputReportsClaimHeldByDifferentActor() error {
	return w.outputContainsAll("claimed", "different")
}

func (w *world) outputReportsTaskAlreadyClosed() error {
	return w.outputContainsAll("already", "closed")
}

func (w *world) jsonClaimActor(expected string) error {
	dec := json.NewDecoder(bytes.NewReader(w.stdout.Bytes()))

	// Try single object.
	var single struct {
		Claim struct {
			Actor *string `json:"actor"`
		} `json:"claim"`
	}
	if err := json.Unmarshal(w.stdout.Bytes(), &single); err == nil && single.Claim.Actor != nil {
		if *single.Claim.Actor != expected {
			return fmt.Errorf("JSON claim actor = %v, expected %q", single.Claim.Actor, expected)
		}
		return nil
	}

	// Try array.
	dec = json.NewDecoder(bytes.NewReader(w.stdout.Bytes()))
	tok, err := dec.Token()
	if err != nil || tok != json.Delim('[') {
		return fmt.Errorf("stdout is neither JSON object nor array; got: %s", w.stdout.String())
	}
	for dec.More() {
		var t task.Task
		if err := dec.Decode(&t); err != nil {
			return fmt.Errorf("parse JSON array element: %w", err)
		}
		if t.Claim.Actor != nil && *t.Claim.Actor == expected {
			return nil
		}
	}
	return fmt.Errorf("JSON array does not contain a claim by %q; got: %s", expected, w.stdout.String())
}

func (w *world) jsonClaimExpiryAfter(minutes int) error {
	var data struct {
		Claim struct {
			ClaimedAt time.Time `json:"claimed_at"`
			ExpiresAt time.Time `json:"expires_at"`
		} `json:"claim"`
	}
	if err := json.Unmarshal(w.stdout.Bytes(), &data); err != nil {
		return fmt.Errorf("stdout is not JSON (%v); got: %s", err, w.stdout.String())
	}
	diff := data.Claim.ExpiresAt.Sub(data.Claim.ClaimedAt)
	expected := time.Duration(minutes) * time.Minute
	if diff != expected {
		return fmt.Errorf("claim expiry is %v after claim time, expected %d minutes (%v)", diff, minutes, expected)
	}
	return nil
}

// --- actor.feature support ------------------------------------------------

func (w *world) setEnv(key, value string) error {
	w.envOverrides = append(w.envOverrides, key)
	return os.Setenv(key, value)
}

func (w *world) setDetectedAgent(agent string) error {
	cmd.DetectedActor = func() string { return agent }
	return nil
}

func (w *world) claimExpiryIsExtended(id string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Claim.ExpiresAt == nil {
		return fmt.Errorf("task %s has no claim expiry", id)
	}
	// Fixture tasks set UpdatedAt to fixtureTime (2026-05-16T12:00Z).
	// A successful renewal bumps UpdatedAt to time.Now().
	if t.UpdatedAt.Equal(fixtureTime) {
		return fmt.Errorf("task %s was not renewed (updated_at still at fixture time)", id)
	}
	if !t.Claim.ExpiresAt.After(time.Now().UTC()) {
		return fmt.Errorf("task %s claim expiry %v is not in the future", id, t.Claim.ExpiresAt)
	}
	return nil
}

// --- ready.feature support ------------------------------------------------

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

// --- note.feature support -------------------------------------------------

func (w *world) taskTitled(id, title string) error {
	return w.taskWithTitleAndStatus(id, title, "open")
}

func (w *world) taskHasNoteFrom(id, actor string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	// Look for a note header containing the actor name.
	idx := strings.Index(t.Body, "## Notes")
	if idx < 0 {
		return fmt.Errorf("task %s has no ## Notes section; body:\n%s", id, t.Body)
	}
	notesSection := t.Body[idx:]
	found := false
	for _, line := range strings.Split(notesSection, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "### ") && strings.Contains(line, " - "+actor) {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("task %s has no note from %q; body:\n%s", id, actor, t.Body)
	}
	return nil
}

func (w *world) noteContainsMessage(message string) error {
	t, err := loadOnlyTask()
	if err != nil {
		return err
	}
	if !strings.Contains(t.Body, message) {
		return fmt.Errorf("note body does not contain %q; body:\n%s", message, t.Body)
	}
	return nil
}

func (w *world) noteHasTimestamp() error {
	t, err := loadOnlyTask()
	if err != nil {
		return err
	}
	// Look for an RFC 3339 timestamp in the Notes section.
	idx := strings.Index(t.Body, "## Notes")
	if idx < 0 {
		return fmt.Errorf("body has no ## Notes section; body:\n%s", t.Body)
	}
	notesSection := t.Body[idx:]
	hasTimestamp := false
	for _, line := range strings.Split(notesSection, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "### ") {
			// Format: "### 2026-05-17T10:30:00Z - actor"
			parts := strings.SplitN(strings.TrimPrefix(line, "### "), " - ", 2)
			if len(parts) >= 1 {
				if _, err := time.Parse(time.RFC3339, strings.TrimSpace(parts[0])); err == nil {
					hasTimestamp = true
					break
				}
			}
		}
	}
	if !hasTimestamp {
		return fmt.Errorf("no RFC 3339 timestamp found in ## Notes section; body:\n%s", t.Body)
	}
	return nil
}

// --- stale.feature support ------------------------------------------------

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

// --- agents.feature support -----------------------------------------------

func (w *world) fileExistsWithContent(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

func (w *world) fileStillHasContent(path, expected string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if string(data) != expected {
		return fmt.Errorf("file %s content = %q, expected %q", path, string(data), expected)
	}
	return nil
}

func (w *world) outputContainsHeading(heading string) error {
	needle := "## " + heading
	if !strings.Contains(w.stdout.String(), needle) {
		return fmt.Errorf("output does not contain heading %q; got:\n%s", needle, w.stdout.String())
	}
	return nil
}

func (w *world) outputDescribesWorkflowSteps() error {
	for _, command := range []string{"tl ready", "tl claim", "tl show", "tl note", "tl close"} {
		if !strings.Contains(w.stdout.String(), command) {
			return fmt.Errorf("output does not describe %s; got:\n%s", command, w.stdout.String())
		}
	}
	return nil
}

func (w *world) outputFormatsCommandsAsMarkdownCodeSpans() error {
	for _, command := range []string{
		"`tl ready --json`",
		"`tl claim <task-id> --actor <your-agent-name>`",
		"`tl show <task-id>`",
		"`tl note <task-id> --actor <your-agent-name> -m \"...\"`",
		"`tl close <task-id> --actor <your-agent-name>`",
	} {
		if !strings.Contains(w.stdout.String(), command) {
			return fmt.Errorf("output does not contain Markdown code span %s; got:\n%s", command, w.stdout.String())
		}
	}
	return nil
}

// --- shared CLI invocation ------------------------------------------------

func (w *world) runTl(args string) error {
	root := cmd.NewRootCmd()
	root.SetOut(w.stdout)
	root.SetErr(w.stderr)
	root.SetArgs(splitArgs(args))
	w.cmdErr = root.Execute()
	return nil
}

// splitArgs splits a CLI argument string while honoring "double-quoted"
// values (so titles with spaces survive `tl create "Add login form"`).
func splitArgs(s string) []string {
	var out []string
	var cur strings.Builder
	inQuote := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch == '"':
			inQuote = !inQuote
		case ch == ' ' && !inQuote:
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(ch)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

// --- JSON output helpers --------------------------------------------------

func (w *world) jsonContainsTaskIdentifier() error {
	id, err := w.jsonStringValue("id")
	if err != nil {
		return err
	}
	if !strings.HasPrefix(id, "task-") {
		return fmt.Errorf("JSON id %q does not look like a task identifier", id)
	}
	return nil
}

// jsonStringField returns a step handler asserting that the named JSON
// string field equals the provided value.
func (w *world) jsonStringField(field string) func(string) error {
	return func(expected string) error {
		got, err := w.jsonStringValue(field)
		if err != nil {
			return err
		}
		if got != expected {
			return fmt.Errorf("JSON %q = %q, expected %q", field, got, expected)
		}
		return nil
	}
}

func (w *world) jsonStringValue(field string) (string, error) {
	// Try array first (e.g. ready --json, list --json).
	var arr []map[string]any
	if err := json.Unmarshal(w.stdout.Bytes(), &arr); err == nil {
		for _, item := range arr {
			if v, ok := item[field].(string); ok && v != "" {
				return v, nil
			}
		}
		return "", fmt.Errorf("JSON array has no entry with non-empty %q; got: %s", field, w.stdout.String())
	}
	// Fall back to single object (e.g. show --json, create --json).
	var data map[string]any
	if err := json.Unmarshal(w.stdout.Bytes(), &data); err != nil {
		return "", fmt.Errorf("stdout is not JSON (%v); got: %s", err, w.stdout.String())
	}
	v, ok := data[field].(string)
	if !ok {
		return "", fmt.Errorf("JSON missing string field %q; got: %s", field, w.stdout.String())
	}
	return v, nil
}

// --- shared utilities -----------------------------------------------------

func loadOnlyTask() (*task.Task, error) {
	dir := filepath.Join(".taskledger", "tasks")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read tasks dir: %w", err)
	}
	if len(entries) != 1 {
		return nil, fmt.Errorf("expected exactly 1 task, found %d", len(entries))
	}
	data, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		return nil, err
	}
	return task.UnmarshalMarkdown(data)
}

func loadFixtureTask(id string) (*task.Task, error) {
	data, err := os.ReadFile(filepath.Join(".taskledger", "tasks", id+".md"))
	if err != nil {
		return nil, err
	}
	return task.UnmarshalMarkdown(data)
}

func assertEventRecorded(eventName, taskID string) error {
	return assertEventRecordedMatching(eventName, taskID, "")
}

func assertEventRecordedBy(eventName, taskID, actor string) error {
	return assertEventRecordedMatching(eventName, taskID, actor)
}

func assertEventRecordedMatching(eventName, taskID, actor string) error {
	f, err := os.Open(filepath.Join(".taskledger", "events.jsonl"))
	if err != nil {
		return fmt.Errorf("open events journal: %w", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var e events.Event
		if err := json.Unmarshal(line, &e); err != nil {
			return fmt.Errorf("parse event line %q: %w", string(line), err)
		}
		if e.Event == eventName && e.TaskID == taskID && (actor == "" || e.Actor == actor) {
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if actor != "" {
		return fmt.Errorf("no %q event for %q by %q in journal", eventName, taskID, actor)
	}
	return fmt.Errorf("no %q event for %q in journal", eventName, taskID)
}

var fixtureTime = time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)

func writeFixtureTask(t *task.Task) error {
	data, err := t.MarshalMarkdown()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(".taskledger", "tasks", t.ID+".md"), data, 0o644)
}

func lineContaining(s, needle string) (string, bool) {
	for _, line := range strings.Split(s, "\n") {
		if strings.Contains(line, needle) {
			return line, true
		}
	}
	return "", false
}

func nonEmptyLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

var listColumnSeparator = regexp.MustCompile(`\s{2,}`)

func splitListLine(line string) []string {
	return listColumnSeparator.Split(strings.TrimSpace(line), -1)
}
