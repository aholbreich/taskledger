package bdd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"

	internalcolor "github.com/aholbreich/tl/internal/color"
	"github.com/aholbreich/tl/internal/task"
)

// --- show.feature support -------------------------------------------------

func initializeShowSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^a task "([^"]*)" titled "([^"]*)" with status "([^"]*)"$`, w.taskWithTitleAndStatus)
	ctx.Step(`^a task "([^"]*)" titled "([^"]*)" with status "([^"]*)" and priority "([^"]*)"$`, w.taskWithTitleStatusAndPriority)
	ctx.Step(`^"([^"]*)" depends on "([^"]*)"$`, w.taskDependsOn)
	ctx.Step(`^"([^"]*)" does not depend on "([^"]*)"$`, w.taskDoesNotDependOn)
	ctx.Step(`^"([^"]*)" has a description "([^"]*)"$`, w.taskHasDescription)
	ctx.Step(`^"([^"]*)" has a note from "([^"]*)" saying "([^"]*)"$`, w.taskHasNote)
	ctx.Step(`^no task with identifier "([^"]*)" exists$`, w.noTaskWithIdentifierExists)
	ctx.Step(`^a task "([^"]*)" with no dependencies$`, w.taskWithNoDependencies)
	ctx.Step(`^a task "([^"]*)" exists$`, w.taskExists)
	ctx.Step(`^an event "([^"]*)" is recorded for "([^"]*)"$`, w.eventRecordedFor)
	ctx.Step(`^"([^"]*)" has no dependencies$`, w.taskHasNoDependencies)
	ctx.Step(`^the output contains identifier "([^"]*)"$`, w.outputContains)
	ctx.Step(`^the output contains title "([^"]*)"$`, w.outputContains)
	ctx.Step(`^the output contains status "([^"]*)"$`, w.outputContains)
	ctx.Step(`^the output contains dependency "([^"]*)"$`, w.outputContains)
	ctx.Step(`^the output contains the note from "([^"]*)"$`, w.outputContains)
	ctx.Step(`^the command exits with code (\d+)$`, w.commandExitsWithCode)
	ctx.Step(`^the output reports that "([^"]*)" was not found$`, w.outputReportsNotFound)
	ctx.Step(`^the output colorizes "([^"]*)" with "([^"]*)"$`, w.outputColorizesTextWith)
	ctx.Step(`^the output colorizes the new task identifier$`, w.outputColorizesNewTaskIdentifier)
	ctx.Step(`^the output colorizes the line for "([^"]*)" with "([^"]*)"$`, w.outputColorizesLineForWith)
	ctx.Step(`^the output does not contain ANSI color$`, w.outputDoesNotContainANSIColor)
}

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
	return w.taskWithTitleStatusAndPriority(id, title, status, "medium")
}

func (w *world) taskWithTitleStatusAndPriority(id, title, status, priority string) error {
	return writeFixtureTask(&task.Task{
		ID:        id,
		Title:     title,
		Status:    status,
		Priority:  priority,
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

func (w *world) taskHasDescription(id, description string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	t.Body = "## Description\n\n" + description + "\n"
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

func (w *world) outputColorizesTextWith(text, colorName string) error {
	return w.outputContains(coloredText(colorName, text))
}

func (w *world) outputColorizesNewTaskIdentifier() error {
	t, err := loadOnlyTask()
	if err != nil {
		return err
	}
	return w.outputContains(internalcolor.ID(t.ID))
}

func (w *world) outputColorizesLineForWith(id, colorName string) error {
	line, ok := lineContaining(w.stdout.String(), id)
	if !ok {
		return fmt.Errorf("output does not contain line for %q; got:\n%s", id, w.stdout.String())
	}
	if !strings.HasPrefix(line, colorCode(colorName)) || !strings.HasSuffix(line, internalcolor.Reset) {
		return fmt.Errorf("line for %q is not colorized with %s; line: %q", id, colorName, line)
	}
	return nil
}

func (w *world) outputDoesNotContainANSIColor() error {
	combined := w.stdout.String() + w.stderr.String()
	if strings.Contains(combined, "\x1b[") {
		return fmt.Errorf("output contains ANSI color; got:\n%s", combined)
	}
	return nil
}

func coloredText(colorName, text string) string {
	return internalcolor.Apply(colorCode(colorName), text)
}

func colorCode(colorName string) string {
	switch colorName {
	case "red":
		return internalcolor.Red
	case "green":
		return internalcolor.Green
	case "yellow":
		return internalcolor.Yellow
	case "magenta":
		return internalcolor.Magenta
	case "blue":
		return internalcolor.Blue
	case "bright blue":
		return internalcolor.BrightBlue
	case "cyan":
		return internalcolor.Cyan
	case "bold":
		return internalcolor.Bold
	case "dim":
		return internalcolor.Dim
	default:
		return ""
	}
}
