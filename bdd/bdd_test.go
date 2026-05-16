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
	tempDir string
	origWd  string
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer
	cmdErr  error
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
	ctx.Step(`^the new task has priority "([^"]*)"$`, w.newTaskHasPriority)
	ctx.Step(`^the new task has tags "([^"]*)" and "([^"]*)"$`, w.newTaskHasTwoTags)
	ctx.Step(`^the new task has no dependencies$`, w.newTaskHasNoDependencies)
	ctx.Step(`^the new task description is "([^"]*)"$`, w.newTaskDescriptionIs)
	ctx.Step(`^an event "([^"]*)" is recorded for the new task$`, w.eventRecordedForNewTask)
	ctx.Step(`^the JSON output body contains "([^"]*)"$`, w.jsonBodyContains)

	// list.feature preconditions and outcomes
	ctx.Step(`^the following tasks exist:$`, w.followingTasksExist)
	ctx.Step(`^the output lists "([^"]*)" with status "([^"]*)" and title "([^"]*)"$`, w.outputListsTask)
	ctx.Step(`^the JSON output is an array of (\d+) tasks$`, w.jsonOutputIsArrayOfTasks)
	ctx.Step(`^the JSON output contains a task with identifier "([^"]*)"$`, w.jsonArrayContainsTaskID)
	ctx.Step(`^the listed task identifiers appear in this order:$`, w.listedTaskIDsAppearInOrder)

	// show.feature preconditions and outcomes
	ctx.Step(`^a task "([^"]*)" titled "([^"]*)" with status "([^"]*)"$`, w.taskWithTitleAndStatus)
	ctx.Step(`^"([^"]*)" depends on "([^"]*)"$`, w.taskDependsOn)
	ctx.Step(`^"([^"]*)" has a note from "([^"]*)" saying "([^"]*)"$`, w.taskHasNote)
	ctx.Step(`^no task with identifier "([^"]*)" exists$`, w.noTaskWithIdentifierExists)
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

// --- list.feature support -------------------------------------------------

func (w *world) followingTasksExist(table *godog.Table) error {
	for rowIdx, row := range table.Rows[1:] {
		values := map[string]string{}
		for i, cell := range row.Cells {
			values[table.Rows[0].Cells[i].Value] = cell.Value
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
			return fmt.Errorf("task row %d is missing title", rowIdx+1)
		}
		if err := writeFixtureTask(&task.Task{
			ID:        id,
			Title:     title,
			Status:    status,
			Priority:  priority,
			CreatedAt: fixtureTime,
			UpdatedAt: fixtureTime,
			CreatedBy: "human",
			DependsOn: []string{},
			Tags:      []string{},
			Verify: task.Verify{
				Commands:         []string{},
				EvidenceRequired: []string{},
			},
		}); err != nil {
			return err
		}
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
		Verify: task.Verify{
			Commands:         []string{},
			EvidenceRequired: []string{},
		},
	})
}

func (w *world) taskDependsOn(id, dependencyID string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	t.DependsOn = append(t.DependsOn, dependencyID)
	return writeFixtureTask(t)
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
// values (so titles with spaces survive `tl create --title "Add login form"`).
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
		if e.Event == eventName && e.TaskID == taskID {
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return err
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
