package bdd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/aholbreich/tl/internal/task"
)

// --- notes-format.feature support ----------------------------------------

func initializeNotesFormatSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^"([^"]*)" has a canonical "([^"]*)" note from "([^"]*)" with message "([^"]*)"$`, w.taskHasCanonicalNote)
	ctx.Step(`^the developer asks for JSON for "([^"]*)"$`, w.developerAsksForTaskJSON)
	ctx.Step(`^the JSON output contains a parsed "([^"]*)" note from "([^"]*)" with message "([^"]*)"$`, w.jsonContainsParsedNote)
}

func (w *world) taskHasCanonicalNote(id, kind, actor, message string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	idx := strings.Index(t.Body, "## Notes")
	if idx < 0 {
		return fmt.Errorf("task %s has no ## Notes section; body:\n%s", id, t.Body)
	}
	for _, line := range strings.Split(t.Body[idx:], "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") && strings.Contains(line, "["+actor+"] "+kind+": "+message) {
			return nil
		}
	}
	return fmt.Errorf("task %s has no canonical %q note from %q with message %q; body:\n%s", id, kind, actor, message, t.Body)
}

func (w *world) developerAsksForTaskJSON(id string) error {
	w.stdout.Reset()
	w.stderr.Reset()
	return w.runTl("show " + id + " --json")
}

func (w *world) jsonContainsParsedNote(kind, actor, message string) error {
	var data struct {
		Notes []task.Note `json:"notes"`
	}
	if err := json.Unmarshal(w.stdout.Bytes(), &data); err != nil {
		return fmt.Errorf("stdout is not task JSON (%v); got: %s", err, w.stdout.String())
	}
	for _, note := range data.Notes {
		if note.Kind == kind && note.Actor == actor && note.Message == message {
			return nil
		}
	}
	return fmt.Errorf("JSON notes do not contain %q note from %q with message %q; notes: %#v", kind, actor, message, data.Notes)
}
