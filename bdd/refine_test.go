package bdd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	"github.com/aholbreich/tl/internal/events"
)

// --- refine.feature support ----------------------------------------------

func initializeRefineSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^"([^"]*)" has title "([^"]*)"$`, w.taskHasTitle)
	ctx.Step(`^"([^"]*)" has priority "([^"]*)"$`, w.taskHasPriority)
	ctx.Step(`^"([^"]*)" has type "([^"]*)"$`, w.taskHasType)
	ctx.Step(`^"([^"]*)" has the description "([^"]*)"$`, w.taskHasDescriptionText)
	ctx.Step(`^the developer's editor will save:$`, w.editorWillSave)
	ctx.Step(`^the developer's editor saves the buffer unchanged$`, w.editorSavesUnchanged)
	ctx.Step(`^no system editor is configured$`, w.noSystemEditorConfigured)
	ctx.Step(`^no event "([^"]*)" is recorded for "([^"]*)"$`, w.noEventRecordedFor)
	ctx.Step(`^the output reports that no fields were given to refine$`, w.outputReportsNoFieldsToRefine)
	ctx.Step(`^the output reports that no editor is configured$`, w.outputReportsNoEditorConfigured)
}

func (w *world) taskHasTitle(id, title string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Title != title {
		return fmt.Errorf("task %s title = %q, expected %q", id, t.Title, title)
	}
	return nil
}

func (w *world) taskHasPriority(id, priority string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Priority != priority {
		return fmt.Errorf("task %s priority = %q, expected %q", id, t.Priority, priority)
	}
	return nil
}

func (w *world) taskHasType(id, taskType string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Type != taskType {
		return fmt.Errorf("task %s type = %q, expected %q", id, t.Type, taskType)
	}
	return nil
}

func (w *world) taskHasDescriptionText(id, description string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	got := strings.TrimSpace(extractMarkdownSection(t.Body, "Description"))
	if got != description {
		return fmt.Errorf("task %s description = %q, expected %q (body: %q)", id, got, description, t.Body)
	}
	return nil
}

func (w *world) editorWillSave(content *godog.DocString) error {
	path := filepath.Join(w.tempDir, "fake-editor.sh")
	script := "#!/bin/sh\ncat > \"$1\" <<'TL_EDITOR_EOF'\n" + content.Content + "\nTL_EDITOR_EOF\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		return err
	}
	w.envOverrides = append(w.envOverrides, "VISUAL", "EDITOR")
	_ = os.Unsetenv("VISUAL")
	return os.Setenv("EDITOR", path)
}

func (w *world) editorSavesUnchanged() error {
	path := filepath.Join(w.tempDir, "fake-editor-unchanged.sh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		return err
	}
	w.envOverrides = append(w.envOverrides, "VISUAL", "EDITOR")
	_ = os.Unsetenv("VISUAL")
	return os.Setenv("EDITOR", path)
}

func (w *world) noSystemEditorConfigured() error {
	w.envOverrides = append(w.envOverrides, "VISUAL", "EDITOR")
	_ = os.Unsetenv("VISUAL")
	return os.Unsetenv("EDITOR")
}

func (w *world) noEventRecordedFor(eventName, taskID string) error {
	f, err := os.Open(filepath.Join(".taskledger", "events.jsonl"))
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
		if e.Event == eventName && e.TaskID == taskID {
			return fmt.Errorf("unexpected %q event for %q in journal", eventName, taskID)
		}
	}
	return scanner.Err()
}

func (w *world) outputReportsNoFieldsToRefine() error {
	combined := strings.ToLower(w.stdout.String() + w.stderr.String())
	if w.cmdErr != nil {
		combined += "\n" + strings.ToLower(w.cmdErr.Error())
	}
	if !strings.Contains(combined, "no fields") || !strings.Contains(combined, "refine") {
		return fmt.Errorf("output does not report no fields were given to refine; got:\n%s", combined)
	}
	return nil
}

func (w *world) outputReportsNoEditorConfigured() error {
	combined := strings.ToLower(w.stdout.String() + w.stderr.String())
	if w.cmdErr != nil {
		combined += "\n" + strings.ToLower(w.cmdErr.Error())
	}
	if !strings.Contains(combined, "no editor") || !strings.Contains(combined, "configured") {
		return fmt.Errorf("output does not report that no editor is configured; got:\n%s", combined)
	}
	return nil
}
