// Package bdd runs the project's Gherkin features under godog as the BDD
// acceptance test suite. Step definitions invoke the cobra CLI in-process so
// scenarios exercise the full command surface without spawning subprocesses.
//
// Step definitions are organised by feature file in steps_*.go.
// This file holds the test runner, world struct, shared helpers,
// and CLI invocation plumbing.
package bdd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"github.com/aholbreich/tl/cmd"
	"github.com/aholbreich/tl/internal/events"
	"github.com/aholbreich/tl/internal/task"
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

// InitializeScenario registers every step definition and wires Before/After
// hooks that isolate each scenario inside a fresh temp directory.
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
		// Clear env that leaks from the host (notably CI runners, which set
		// XDG_CONFIG_HOME). The completion-install command resolves paths
		// from these before falling back to $HOME, so an inherited value
		// makes scenarios that only override HOME non-deterministic.
		for _, key := range []string{"XDG_CONFIG_HOME", "XDG_DATA_HOME", "ZDOTDIR"} {
			if err := os.Unsetenv(key); err != nil {
				return ctx, err
			}
		}
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
	ctx.Step(`^an initialized task ledger repository$`, w.ledgerInitialized)
	ctx.Step(`^no tasks exist$`, w.noTasksExist)

	// Shared CLI invocation (handles developer/agent phrasings).
	ctx.Step("^the (?:developer|agent|agent or developer) runs `tl (.+)`$", w.runTl)

	// --- per-feature step registration ----------------------------------
	initializeInitSteps(ctx, w)
	initializeCreateSteps(ctx, w)
	initializeBulkJSONSteps(ctx, w)
	initializeListSteps(ctx, w)
	initializeShowSteps(ctx, w)
	initializeLedgerRequiredSteps(ctx, w)
	initializeClaimSteps(ctx, w)
	initializeActorSteps(ctx, w)
	initializeReadySteps(ctx, w)
	initializeNoteSteps(ctx, w)
	initializeNotesFormatSteps(ctx, w)
	initializeStaleSteps(ctx, w)
	initializeHistorySteps(ctx, w)
	initializeAgentsSteps(ctx, w)
	initializeBlockSteps(ctx, w)
	initializeCancelSteps(ctx, w)
	initializeResolveSteps(ctx, w)
	initializeRefineSteps(ctx, w)
	initializeCompletionSteps(ctx, w)
	initializeCompletionInstallSteps(ctx, w)
	initializeReferencesSteps(ctx, w)
	initializeDoctorSteps(ctx, w)
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
// An explicit "" is preserved as an empty positional argument — required for
// completion scenarios like `tl __complete show ""`.
func splitArgs(s string) []string {
	var out []string
	var cur strings.Builder
	inQuote := false
	started := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch == '"':
			inQuote = !inQuote
			started = true
		case ch == ' ' && !inQuote:
			if started {
				out = append(out, cur.String())
				cur.Reset()
				started = false
			}
		default:
			cur.WriteByte(ch)
			started = true
		}
	}
	if started {
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
	if id == "" {
		return fmt.Errorf("JSON output has empty id")
	}
	return nil
}

func (w *world) jsonStringField(field string) func(string) error {
	return func(expected string) error {
		got, err := w.jsonStringValue(field)
		if err != nil {
			return err
		}
		if got != expected {
			return fmt.Errorf("JSON %s = %q, expected %q", field, got, expected)
		}
		return nil
	}
}

func (w *world) jsonStringValue(field string) (string, error) {
	// Try single JSON object first.
	var data map[string]json.RawMessage
	if err := json.Unmarshal(w.stdout.Bytes(), &data); err == nil {
		if raw, ok := data[field]; ok {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return "", fmt.Errorf("JSON field %q is not a string (%v)", field, err)
			}
			return s, nil
		}
		return "", fmt.Errorf("JSON output missing field %q", field)
	}

	// Try JSON array. Search for first element with the field.
	var tasks []task.Task
	if err := json.Unmarshal(w.stdout.Bytes(), &tasks); err != nil {
		return "", fmt.Errorf("stdout is not JSON (%v); got: %s", err, w.stdout.String())
	}
	if len(tasks) == 0 {
		return "", fmt.Errorf("JSON array is empty")
	}
	// Use first task's data to extract the field via re-marshal.
	d, _ := json.Marshal(tasks[0])
	var data2 map[string]json.RawMessage
	json.Unmarshal(d, &data2)
	if raw, ok := data2[field]; ok {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return "", fmt.Errorf("JSON field %q is not a string (%v)", field, err)
		}
		return s, nil
	}
	return "", fmt.Errorf("JSON output missing field %q", field)
}

// --- shared utilities -----------------------------------------------------

func loadOnlyTask() (*task.Task, error) {
	entries, err := os.ReadDir(filepath.Join(".tl", "tasks"))
	if err != nil {
		return nil, err
	}
	var taskFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			taskFiles = append(taskFiles, e.Name())
		}
	}
	if len(taskFiles) == 0 {
		return nil, fmt.Errorf("no task files found")
	}
	if len(taskFiles) > 1 {
		return nil, fmt.Errorf("expected exactly 1 task, got %d: %v", len(taskFiles), taskFiles)
	}
	return loadFixtureTask(strings.TrimSuffix(taskFiles[0], ".md"))
}

func loadFixtureTask(id string) (*task.Task, error) {
	data, err := os.ReadFile(filepath.Join(".tl", "tasks", id+".md"))
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

func assertEventRecordedWithValue(eventName, taskID, value string) error {
	f, err := os.Open(filepath.Join(".tl", "events.jsonl"))
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e events.Event
		if err := json.Unmarshal(line, &e); err != nil {
			return fmt.Errorf("parse event line %q: %w", string(line), err)
		}
		if e.Event == eventName && e.TaskID == taskID && e.Value == value {
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return fmt.Errorf("no %q event for %q with value %q in journal", eventName, taskID, value)
}

// containsString reports whether list contains v (bdd-local; mirrors the
// helper in package cmd, which is not importable here).
func containsString(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func assertEventRecordedMatching(eventName, taskID, actor string) error {
	f, err := os.Open(filepath.Join(".tl", "events.jsonl"))
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
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
	return os.WriteFile(filepath.Join(".tl", "tasks", t.ID+".md"), data, 0o644)
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
