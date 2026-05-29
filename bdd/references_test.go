package bdd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/aholbreich/tl/internal/task"
)

// --- references.feature support -------------------------------------------

func initializeReferencesSteps(ctx *godog.ScenarioContext, w *world) {
	// Fixtures
	ctx.Step(`^a task "([^"]*)" with no references$`, w.fixtureTaskWithNoRefs)
	ctx.Step(`^a task "([^"]*)" with reference "([^"]*)"$`, w.fixtureTaskWithRef)
	ctx.Step(`^a task "([^"]*)" with references "([^"]*)" and "([^"]*)"$`, w.fixtureTaskWithTwoRefs)

	// New-task assertions (single task in the ledger)
	ctx.Step(`^the new task has references containing "([^"]*)"$`, w.newTaskHasReference)
	ctx.Step(`^the new task has no references$`, w.newTaskHasNoReferences)
	ctx.Step(`^the new task has tag "([^"]*)"$`, w.newTaskHasTag)

	// Named-task assertions
	ctx.Step(`^"([^"]*)" has references containing "([^"]*)"$`, w.taskHasReference)
	ctx.Step(`^"([^"]*)" has exactly one reference matching "([^"]*)"$`, w.taskHasExactlyOneReference)
	ctx.Step(`^"([^"]*)" does not have reference "([^"]*)"$`, w.taskDoesNotHaveReference)
	ctx.Step(`^"([^"]*)" has no references$`, w.taskHasNoReferences)

	// Display ordering
	ctx.Step(`^the "([^"]*)" line appears after the "([^"]*)" line$`, w.lineAppearsAfter)
	ctx.Step(`^the "([^"]*)" line appears before the "([^"]*)" line$`, w.lineAppearsBefore)

	// JSON
	ctx.Step(`^the JSON output has a "references" array containing "([^"]*)"$`, w.jsonReferencesContains)
	ctx.Step(`^the JSON output has an empty "references" array$`, w.jsonReferencesEmpty)

	// Events
	ctx.Step(`^a "([^"]*)" event is recorded for the new task with value "([^"]*)"$`, w.eventRecordedForNewTaskWithValue)
	ctx.Step(`^a "([^"]*)" event is recorded for "([^"]*)" with value "([^"]*)"$`, w.eventRecordedForTaskWithValue)
	ctx.Step(`^no "([^"]*)" event is recorded for "([^"]*)" in this invocation$`, w.noEventRecordedFor)
}

// --- fixtures -------------------------------------------------------------

func (w *world) writeRefFixture(id string, refs []string) error {
	if err := writeFixtureTask(&task.Task{
		ID:         id,
		Title:      id,
		Status:     "open",
		Priority:   "medium",
		CreatedAt:  fixtureTime,
		UpdatedAt:  fixtureTime,
		CreatedBy:  "human",
		DependsOn:  []string{},
		Tags:       []string{},
		References: refs,
	}); err != nil {
		return err
	}
	// Preconditions seed only the creation event — never reference_added —
	// so idempotency scenarios that scan the journal stay accurate.
	return recordFixtureEvent("created", id, "human", fixtureTime)
}

func (w *world) fixtureTaskWithNoRefs(id string) error {
	return w.writeRefFixture(id, nil)
}

func (w *world) fixtureTaskWithRef(id, ref string) error {
	return w.writeRefFixture(id, []string{ref})
}

func (w *world) fixtureTaskWithTwoRefs(id, a, b string) error {
	return w.writeRefFixture(id, []string{a, b})
}

// --- task-state assertions ------------------------------------------------

func (w *world) newTaskHasReference(ref string) error {
	t, err := loadOnlyTask()
	if err != nil {
		return err
	}
	return assertHasReference(t, ref)
}

func (w *world) newTaskHasNoReferences() error {
	t, err := loadOnlyTask()
	if err != nil {
		return err
	}
	if len(t.References) != 0 {
		return fmt.Errorf("task has references %v, expected none", t.References)
	}
	return nil
}

func (w *world) newTaskHasTag(tag string) error {
	t, err := loadOnlyTask()
	if err != nil {
		return err
	}
	if !containsString(t.Tags, tag) {
		return fmt.Errorf("task tags %v do not contain %q", t.Tags, tag)
	}
	return nil
}

func (w *world) taskHasReference(id, ref string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	return assertHasReference(t, ref)
}

func (w *world) taskHasExactlyOneReference(id, ref string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	n := 0
	for _, r := range t.References {
		if r == ref {
			n++
		}
	}
	if n != 1 {
		return fmt.Errorf("task %s has %d references matching %q (refs: %v), expected exactly 1", id, n, ref, t.References)
	}
	return nil
}

func (w *world) taskDoesNotHaveReference(id, ref string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if containsString(t.References, ref) {
		return fmt.Errorf("task %s still has reference %q (refs: %v)", id, ref, t.References)
	}
	return nil
}

func (w *world) taskHasNoReferences(id string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if len(t.References) != 0 {
		return fmt.Errorf("task %s has references %v, expected none", id, t.References)
	}
	return nil
}

func assertHasReference(t *task.Task, ref string) error {
	if !containsString(t.References, ref) {
		return fmt.Errorf("references %v do not contain %q", t.References, ref)
	}
	return nil
}

// --- display ordering -----------------------------------------------------

func (w *world) lineAppearsAfter(label, before string) error {
	li, bi, err := w.labelLineIndices(label, before)
	if err != nil {
		return err
	}
	if li <= bi {
		return fmt.Errorf("%q (line %d) does not appear after %q (line %d)\n%s", label, li, before, bi, w.stdout.String())
	}
	return nil
}

func (w *world) lineAppearsBefore(label, after string) error {
	li, ai, err := w.labelLineIndices(label, after)
	if err != nil {
		return err
	}
	if li >= ai {
		return fmt.Errorf("%q (line %d) does not appear before %q (line %d)\n%s", label, li, after, ai, w.stdout.String())
	}
	return nil
}

// labelLineIndices returns the 0-based line index of the lines beginning with
// "<a>" and "<b>" labels in stdout.
func (w *world) labelLineIndices(a, b string) (int, int, error) {
	ai, bi := -1, -1
	for i, line := range strings.Split(w.stdout.String(), "\n") {
		trimmed := strings.TrimSpace(line)
		if ai < 0 && strings.HasPrefix(trimmed, a) {
			ai = i
		}
		if bi < 0 && strings.HasPrefix(trimmed, b) {
			bi = i
		}
	}
	if ai < 0 {
		return 0, 0, fmt.Errorf("no line starting with %q in output:\n%s", a, w.stdout.String())
	}
	if bi < 0 {
		return 0, 0, fmt.Errorf("no line starting with %q in output:\n%s", b, w.stdout.String())
	}
	return ai, bi, nil
}

// --- JSON -----------------------------------------------------------------

func (w *world) jsonReferencesContains(value string) error {
	refs, err := w.jsonReferences()
	if err != nil {
		return err
	}
	if !containsString(refs, value) {
		return fmt.Errorf("JSON references %v do not contain %q", refs, value)
	}
	return nil
}

func (w *world) jsonReferencesEmpty() error {
	refs, err := w.jsonReferences()
	if err != nil {
		return err
	}
	if len(refs) != 0 {
		return fmt.Errorf("JSON references %v, expected empty array", refs)
	}
	return nil
}

func (w *world) jsonReferences() ([]string, error) {
	var data map[string]json.RawMessage
	if err := json.Unmarshal(w.stdout.Bytes(), &data); err != nil {
		return nil, fmt.Errorf("stdout is not a JSON object (%v); got: %s", err, w.stdout.String())
	}
	raw, ok := data["references"]
	if !ok {
		return nil, fmt.Errorf("JSON output missing field \"references\"; got: %s", w.stdout.String())
	}
	var refs []string
	if err := json.Unmarshal(raw, &refs); err != nil {
		return nil, fmt.Errorf("JSON \"references\" is not a string array (%v)", err)
	}
	return refs, nil
}

// --- events ---------------------------------------------------------------

func (w *world) eventRecordedForNewTaskWithValue(eventName, value string) error {
	t, err := loadOnlyTask()
	if err != nil {
		return err
	}
	return assertEventRecordedWithValue(eventName, t.ID, value)
}

func (w *world) eventRecordedForTaskWithValue(eventName, taskID, value string) error {
	return assertEventRecordedWithValue(eventName, taskID, value)
}
