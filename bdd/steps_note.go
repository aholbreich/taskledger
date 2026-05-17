package bdd

import (
	"fmt"
	"github.com/cucumber/godog"
	"strings"
	"time"
)

// --- note.feature support -------------------------------------------------

func initializeNoteSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^a task "([^"]*)" titled "([^"]*)"$`, w.taskTitled)
	ctx.Step(`^"([^"]*)" has a note from "([^"]*)"$`, w.taskHasNoteFrom)
	ctx.Step(`^the note contains the message "([^"]*)"$`, w.noteContainsMessage)
	ctx.Step(`^the note has a timestamp$`, w.noteHasTimestamp)
}

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
