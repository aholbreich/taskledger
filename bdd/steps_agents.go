package bdd

import (
	"fmt"
	"github.com/cucumber/godog"
	"os"
	"strings"
)

// --- agents.feature support -----------------------------------------------

func initializeAgentsSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^the file "([^"]*)" exists with content "([^"]*)"$`, w.fileExistsWithContent)
	ctx.Step(`^the file "([^"]*)" still has content "([^"]*)"$`, w.fileStillHasContent)
	ctx.Step(`^the output contains a "([^"]*)" heading$`, w.outputContainsHeading)
	ctx.Step(`^the output describes the ready, claim, show, note, and close steps$`, w.outputDescribesWorkflowSteps)
	ctx.Step(`^the output formats task commands as Markdown code spans$`, w.outputFormatsCommandsAsMarkdownCodeSpans)
}

func (w *world) fileExistsWithContent(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

func (w *world) fileStillHasContent(path, expected string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if string(data) != expected {
		return fmt.Errorf("file %s content = %q, expected %q", path, string(data), expected)
	}
	return nil
}

func (w *world) outputContainsHeading(heading string) error {
	needle := "## " + heading
	if !strings.Contains(w.stdout.String(), needle) {
		return fmt.Errorf("output does not contain heading %q; got:\n%s", needle, w.stdout.String())
	}
	return nil
}

func (w *world) outputDescribesWorkflowSteps() error {
	for _, command := range []string{"tl ready", "tl claim", "tl show", "tl note", "tl close"} {
		if !strings.Contains(w.stdout.String(), command) {
			return fmt.Errorf("output does not describe %s; got:\n%s", command, w.stdout.String())
		}
	}
	return nil
}

func (w *world) outputFormatsCommandsAsMarkdownCodeSpans() error {
	for _, command := range []string{
		"`tl ready --json`",
		"`tl claim <task-id> --actor <your-agent-name>`",
		"`tl show <task-id>`",
		"`tl note <task-id> --actor <your-agent-name> -m \"...\"`",
		"`tl close <task-id> --actor <your-agent-name>`",
	} {
		if !strings.Contains(w.stdout.String(), command) {
			return fmt.Errorf("output does not contain Markdown code span %s; got:\n%s", command, w.stdout.String())
		}
	}
	return nil
}
