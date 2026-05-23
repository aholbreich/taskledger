package bdd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/aholbreich/tl/internal/events"
)

// --- history.feature support ---------------------------------------------

func initializeHistorySteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^the history output lists events in this order:$`, w.historyOutputListsEventsInOrder)
	ctx.Step(`^the history output for the "([^"]*)" event shows actor "([^"]*)"$`, w.historyEventShowsActor)
	ctx.Step(`^the history output for the "([^"]*)" event has a timestamp$`, w.historyEventHasTimestamp)
	ctx.Step(`^the JSON output is an array of event objects for "([^"]*)"$`, w.jsonOutputIsEventArrayForTask)
	ctx.Step(`^each event object contains a type, a timestamp, and an actor$`, w.eachEventObjectContainsRequiredFields)
	ctx.Step(`^the output reports that a task ID or --since duration is required$`, w.outputReportsHistorySelectorRequired)
	ctx.Step(`^the following events were recorded:$`, w.followingEventsWereRecorded)
	ctx.Step(`^the following events were recorded for "([^"]*)":$`, w.followingEventsWereRecordedForTask)
	ctx.Step(`^the history output contains an event for "([^"]*)"$`, w.historyOutputContainsEventForTask)
	ctx.Step(`^the history output does not contain an event for "([^"]*)"$`, w.historyOutputDoesNotContainEventForTask)
	ctx.Step(`^the history output contains the event "([^"]*)"$`, w.historyOutputContainsEvent)
	ctx.Step(`^the history output does not contain the event "([^"]*)"$`, w.historyOutputDoesNotContainEvent)
}

func recordFixtureEvent(eventName, taskID, actor string, eventTime time.Time) error {
	return events.Append(".taskledger", events.Event{
		Time:   eventTime,
		Event:  eventName,
		TaskID: taskID,
		Actor:  actor,
	})
}

func (w *world) historyOutputListsEventsInOrder(table *godog.Table) error {
	var expected []string
	for _, row := range table.Rows[1:] {
		expected = append(expected, row.Cells[0].Value)
	}
	actual := historyOutputEvents(w.stdout.String())
	if strings.Join(actual, ",") != strings.Join(expected, ",") {
		return fmt.Errorf("history events = %v, expected %v; output:\n%s", actual, expected, w.stdout.String())
	}
	return nil
}

func (w *world) historyEventShowsActor(eventName, actor string) error {
	fields, ok := historyLineFieldsForEvent(w.stdout.String(), eventName)
	if !ok {
		return fmt.Errorf("history output does not contain event %q; got:\n%s", eventName, w.stdout.String())
	}
	if len(fields) < 4 || fields[3] != actor {
		return fmt.Errorf("history event %q actor = %q, expected %q; line fields: %v", eventName, historyField(fields, 3), actor, fields)
	}
	return nil
}

func (w *world) historyEventHasTimestamp(eventName string) error {
	fields, ok := historyLineFieldsForEvent(w.stdout.String(), eventName)
	if !ok {
		return fmt.Errorf("history output does not contain event %q; got:\n%s", eventName, w.stdout.String())
	}
	if _, err := time.Parse(time.RFC3339, fields[0]); err != nil {
		return fmt.Errorf("history event %q timestamp is invalid: %q", eventName, fields[0])
	}
	return nil
}

func (w *world) jsonOutputIsEventArrayForTask(taskID string) error {
	history, err := w.jsonEventArray()
	if err != nil {
		return err
	}
	if len(history) == 0 {
		return fmt.Errorf("JSON event array is empty")
	}
	for _, e := range history {
		if e.TaskID != taskID {
			return fmt.Errorf("event task_id = %q, expected %q; got: %s", e.TaskID, taskID, w.stdout.String())
		}
	}
	return nil
}

func (w *world) eachEventObjectContainsRequiredFields() error {
	history, err := w.jsonEventArray()
	if err != nil {
		return err
	}
	for _, e := range history {
		if e.Event == "" || e.Time.IsZero() || e.Actor == "" {
			data, _ := json.Marshal(e)
			return fmt.Errorf("event object is missing type, timestamp, or actor: %s", string(data))
		}
	}
	return nil
}

func (w *world) outputReportsHistorySelectorRequired() error {
	return w.outputContainsAll("task id", "--since", "required")
}

func (w *world) followingEventsWereRecorded(table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		values := map[string]string{}
		for i, cell := range row.Cells {
			values[table.Rows[0].Cells[i].Value] = strings.TrimSpace(cell.Value)
		}
		taskID := values["task"]
		if taskID == "" {
			return fmt.Errorf("event fixture row is missing task")
		}
		if err := w.taskExists(taskID); err != nil {
			return err
		}
		eventTime, err := relativeFixtureTime(values["when"])
		if err != nil {
			return err
		}
		if err := recordFixtureEvent(values["event"], taskID, "human", eventTime); err != nil {
			return err
		}
	}
	return nil
}

func (w *world) followingEventsWereRecordedForTask(taskID string, table *godog.Table) error {
	if err := w.taskExists(taskID); err != nil {
		return err
	}
	for _, row := range table.Rows[1:] {
		values := map[string]string{}
		for i, cell := range row.Cells {
			values[table.Rows[0].Cells[i].Value] = strings.TrimSpace(cell.Value)
		}
		eventTime, err := relativeFixtureTime(values["when"])
		if err != nil {
			return err
		}
		if err := recordFixtureEvent(values["event"], taskID, "human", eventTime); err != nil {
			return err
		}
	}
	return nil
}

func (w *world) historyOutputContainsEventForTask(taskID string) error {
	for _, fields := range historyOutputRows(w.stdout.String()) {
		if len(fields) >= 3 && fields[2] == taskID {
			return nil
		}
	}
	return fmt.Errorf("history output does not contain an event for %q; got:\n%s", taskID, w.stdout.String())
}

func (w *world) historyOutputDoesNotContainEventForTask(taskID string) error {
	for _, fields := range historyOutputRows(w.stdout.String()) {
		if len(fields) >= 3 && fields[2] == taskID {
			return fmt.Errorf("history output unexpectedly contains an event for %q in row %v; got:\n%s", taskID, fields, w.stdout.String())
		}
	}
	return nil
}

func (w *world) historyOutputContainsEvent(eventName string) error {
	if !historyOutputHasEvent(w.stdout.String(), eventName) {
		return fmt.Errorf("history output does not contain event %q; got:\n%s", eventName, w.stdout.String())
	}
	return nil
}

func (w *world) historyOutputDoesNotContainEvent(eventName string) error {
	if historyOutputHasEvent(w.stdout.String(), eventName) {
		return fmt.Errorf("history output unexpectedly contains event %q; got:\n%s", eventName, w.stdout.String())
	}
	return nil
}

func (w *world) jsonEventArray() ([]events.Event, error) {
	var history []events.Event
	if err := json.Unmarshal(w.stdout.Bytes(), &history); err != nil {
		return nil, fmt.Errorf("stdout is not a JSON event array (%v); got: %s", err, w.stdout.String())
	}
	return history, nil
}

func historyOutputEvents(output string) []string {
	var out []string
	for _, fields := range historyOutputRows(output) {
		if len(fields) >= 2 {
			out = append(out, fields[1])
		}
	}
	return out
}

func historyLineFieldsForEvent(output, eventName string) ([]string, bool) {
	for _, fields := range historyOutputRows(output) {
		if len(fields) >= 2 && fields[1] == eventName {
			return fields, true
		}
	}
	return nil, false
}

func historyOutputRows(output string) [][]string {
	var rows [][]string
	for _, line := range nonEmptyLines(output) {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if _, err := time.Parse(time.RFC3339, fields[0]); err != nil {
			continue
		}
		rows = append(rows, fields)
	}
	return rows
}

func historyOutputHasEvent(output, eventName string) bool {
	_, ok := historyLineFieldsForEvent(output, eventName)
	return ok
}

func historyField(fields []string, index int) string {
	if index >= len(fields) {
		return ""
	}
	return fields[index]
}

func relativeFixtureTime(when string) (time.Time, error) {
	parts := strings.Fields(when)
	if len(parts) != 3 || parts[2] != "ago" {
		return time.Time{}, fmt.Errorf("invalid relative time %q", when)
	}
	amount, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid relative time %q: %w", when, err)
	}
	var unit time.Duration
	switch strings.TrimSuffix(parts[1], "s") {
	case "minute":
		unit = time.Minute
	case "hour":
		unit = time.Hour
	case "day":
		unit = 24 * time.Hour
	default:
		return time.Time{}, fmt.Errorf("invalid relative time unit %q", parts[1])
	}
	return time.Now().UTC().Truncate(time.Second).Add(-time.Duration(amount) * unit), nil
}
