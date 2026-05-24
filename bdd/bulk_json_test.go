package bdd

import (
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog"
)

// --- bulk-json.feature support -------------------------------------------

func initializeBulkJSONSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^the developer asks for list JSON$`, w.developerAsksForListJSON)
	ctx.Step(`^the developer asks for ready JSON$`, w.developerAsksForReadyJSON)
	ctx.Step(`^the JSON task "([^"]*)" does not include field "([^"]*)"$`, w.jsonTaskDoesNotIncludeField)
	ctx.Step(`^the JSON task "([^"]*)" has description "([^"]*)"$`, w.jsonTaskHasDescription)
	ctx.Step(`^the JSON task "([^"]*)" contains a parsed "([^"]*)" note from "([^"]*)" with message "([^"]*)"$`, w.jsonTaskContainsParsedNote)
}

func (w *world) developerAsksForListJSON() error {
	w.stdout.Reset()
	w.stderr.Reset()
	return w.runTl("list --json")
}

func (w *world) developerAsksForReadyJSON() error {
	w.stdout.Reset()
	w.stderr.Reset()
	return w.runTl("ready --json")
}

func (w *world) jsonTaskDoesNotIncludeField(id, field string) error {
	data, err := w.jsonObjectForTask(id)
	if err != nil {
		return err
	}
	if _, ok := data[field]; ok {
		return fmt.Errorf("JSON task %s includes field %q; task: %s", id, field, string(mustMarshal(data)))
	}
	return nil
}

func (w *world) jsonTaskHasDescription(id, expected string) error {
	data, err := w.jsonObjectForTask(id)
	if err != nil {
		return err
	}
	var got string
	if err := json.Unmarshal(data["description"], &got); err != nil {
		return fmt.Errorf("JSON task %s description is missing or not a string: %v", id, err)
	}
	if got != expected {
		return fmt.Errorf("JSON task %s description = %q, expected %q", id, got, expected)
	}
	return nil
}

func (w *world) jsonTaskContainsParsedNote(id, kind, actor, message string) error {
	data, err := w.jsonObjectForTask(id)
	if err != nil {
		return err
	}
	var notes []struct {
		Actor   string `json:"actor"`
		Kind    string `json:"kind"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data["notes"], &notes); err != nil {
		return fmt.Errorf("JSON task %s notes are missing or invalid: %v", id, err)
	}
	for _, note := range notes {
		if note.Kind == kind && note.Actor == actor && note.Message == message {
			return nil
		}
	}
	return fmt.Errorf("JSON task %s notes do not contain %q note from %q with message %q; notes: %#v", id, kind, actor, message, notes)
}

func (w *world) jsonObjectForTask(id string) (map[string]json.RawMessage, error) {
	var tasks []map[string]json.RawMessage
	if err := json.Unmarshal(w.stdout.Bytes(), &tasks); err != nil {
		return nil, fmt.Errorf("stdout is not a JSON task array (%v); got: %s", err, w.stdout.String())
	}
	for _, item := range tasks {
		var gotID string
		if err := json.Unmarshal(item["id"], &gotID); err != nil {
			continue
		}
		if gotID == id {
			return item, nil
		}
	}
	return nil, fmt.Errorf("JSON task array does not contain %s; got: %s", id, w.stdout.String())
}

func mustMarshal(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}
