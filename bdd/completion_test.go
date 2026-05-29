package bdd

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// --- completion.feature support -------------------------------------------

func initializeCompletionSteps(ctx *godog.ScenarioContext, w *world) {
	// "the current directory has no task ledger" is registered by init_test.go.
	ctx.Step(`^the completion suggestion "([^"]*)" is present$`, w.completionContains)
	ctx.Step(`^the completion suggestion "([^"]*)" is absent$`, w.completionAbsent)
	ctx.Step(`^the completion suggestion "([^"]*)" appears with description "([^"]*)"$`, w.completionContainsWithDescription)
	ctx.Step(`^the completion directive is "([^"]*)"$`, w.completionDirective)
	ctx.Step(`^no task-ID completion suggestions are returned$`, w.completionNoTaskSuggestions)
}

// completionLines returns the suggestion lines emitted by `tl __complete`,
// dropping the trailing ":<n>" directive line.
func (w *world) completionLines() []string {
	var lines []string
	for _, line := range strings.Split(w.stdout.String(), "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func (w *world) completionContains(id string) error {
	for _, line := range w.completionLines() {
		candidate := line
		if i := strings.IndexByte(line, '\t'); i >= 0 {
			candidate = line[:i]
		}
		if candidate == id {
			return nil
		}
	}
	return fmt.Errorf("completion suggestions do not include %q; got:\n%s", id, w.stdout.String())
}

func (w *world) completionAbsent(id string) error {
	for _, line := range w.completionLines() {
		candidate := line
		if i := strings.IndexByte(line, '\t'); i >= 0 {
			candidate = line[:i]
		}
		if candidate == id {
			return fmt.Errorf("completion suggestions unexpectedly include %q; got:\n%s", id, w.stdout.String())
		}
	}
	return nil
}

func (w *world) completionContainsWithDescription(id, description string) error {
	want := id + "\t" + description
	for _, line := range w.completionLines() {
		if line == want {
			return nil
		}
	}
	return fmt.Errorf("completion suggestions do not include %q with description %q; got:\n%s", id, description, w.stdout.String())
}

func (w *world) completionDirective(directive string) error {
	if !strings.Contains(w.stderr.String(), directive) {
		return fmt.Errorf("completion directive %q not found in stderr; got:\n%s", directive, w.stderr.String())
	}
	return nil
}

func (w *world) completionNoTaskSuggestions() error {
	for _, line := range w.completionLines() {
		if strings.HasPrefix(line, "task-") {
			return fmt.Errorf("expected no task-ID suggestions, got:\n%s", w.stdout.String())
		}
	}
	return nil
}
