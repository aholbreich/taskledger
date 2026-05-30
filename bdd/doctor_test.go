package bdd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/aholbreich/tl/cmd"
	"github.com/aholbreich/tl/internal/doctor"
	"github.com/aholbreich/tl/internal/events"
	"github.com/aholbreich/tl/internal/task"
)

// --- doctor.feature support -----------------------------------------------

func initializeDoctorSteps(ctx *godog.ScenarioContext, w *world) {
	// Fixtures — clean and defective task files.
	ctx.Step(`^a task "([^"]*)" with title "([^"]*)" and status "([^"]*)"$`, w.doctorTaskWithTitleAndStatus)
	ctx.Step(`^a task file "([^"]*)" whose content is not valid YAML$`, w.doctorInvalidYAMLFile)
	ctx.Step(`^a task file "([^"]*)" with frontmatter missing the "([^"]*)" field$`, w.doctorFileMissingField)
	ctx.Step(`^a task "([^"]*)" with priority "([^"]*)"$`, w.doctorTaskWithPriority)
	ctx.Step(`^a task "([^"]*)" with type "" \(empty\)$`, w.doctorTaskWithEmptyType)
	ctx.Step(`^a second task file claiming id "([^"]*)" exists in the tasks directory$`, w.doctorSecondFileWithID)
	ctx.Step(`^a task "([^"]*)" with dependency "([^"]*)"$`, w.doctorTaskWithDependency)
	ctx.Step(`^an event in the journal referencing task "([^"]*)"$`, w.doctorEventReferencingTask)
	ctx.Step(`^no task file for "([^"]*)" exists$`, w.doctorNoTaskFileFor)
	ctx.Step(`^a task "([^"]*)" with status "in_progress" and no active claim$`, w.doctorInProgressNoClaim)
	ctx.Step(`^a task "([^"]*)" claimed by "([^"]*)" with an expired lease$`, w.doctorClaimedExpiredLease)
	ctx.Step(`^a task "([^"]*)" with status "open" that still has claim data set$`, w.doctorOpenWithClaimData)
	ctx.Step(`^a task "([^"]*)" with created_at in the year 2099$`, w.doctorCreatedInFuture)
	ctx.Step(`^a task "([^"]*)" with created_at after updated_at$`, w.doctorCreatedAfterUpdated)
	ctx.Step(`^a task "([^"]*)" with claim_expires_at before claim_claimed_at$`, w.doctorExpiryBeforeClaimed)
	ctx.Step(`^an orphaned file "([^"]*)" exists in the tasks directory$`, w.doctorOrphanedTmp)
	ctx.Step(`^a task file "([^"]*)" that cannot be read$`, w.doctorUnreadableFile)
	ctx.Step(`^a task "([^"]*)" whose body contains "([^"]*)"$`, w.doctorBodyContains)
	ctx.Step(`^a task "([^"]*)" whose Notes section contains lines that do not match the canonical format$`, w.doctorMalformedNotes)
	ctx.Step(`^the ledger has no config.yaml file$`, w.doctorNoConfig)
	ctx.Step(`^the config.yaml contains content that is not valid YAML$`, w.doctorInvalidConfig)
	ctx.Step(`^a ledger with (\d+) tasks$`, w.doctorLedgerWithTasks)
	ctx.Step(`^a ledger with (\d+) events$`, w.doctorLedgerWithEvents)

	// Assertions — report shape.
	ctx.Step(`^the doctor reports no issues$`, w.doctorReportsNoIssues)
	ctx.Step(`^the doctor reports an? "([^"]*)" issue for "([^"]*)" with severity "([^"]*)"$`, w.doctorReportsIssueFor)
	ctx.Step(`^the doctor reports an? "([^"]*)" issue with severity "([^"]*)"$`, w.doctorReportsIssue)
	ctx.Step(`^the doctor output contains a scale warning about "([^"]*)" with severity "([^"]*)"$`, w.doctorReportsScaleWarning)

	// Assertions — JSON shape.
	ctx.Step(`^the JSON output is an empty array$`, w.doctorJSONEmptyArray)
	ctx.Step(`^the JSON output is an array of diagnostic objects$`, w.doctorJSONDiagnosticArray)
	ctx.Step(`^a diagnostic object has category "([^"]*)" and severity "([^"]*)"$`, w.doctorJSONHasCategorySeverity)
	ctx.Step(`^a diagnostic object has a non-empty "([^"]*)" field$`, w.doctorJSONHasNonEmptyField)
	ctx.Step(`^a diagnostic object has a "([^"]*)" boolean field$`, w.doctorJSONHasBoolField)

	// Assertions — --fix behaviour.
	ctx.Step(`^the doctor reports the "([^"]*)" issue for "([^"]*)" as fixed$`, w.doctorReportsFixed)
	ctx.Step(`^"([^"]*)" no longer depends on "([^"]*)"$`, w.doctorNoLongerDependsOn)
	ctx.Step(`^the doctor reports the orphaned file as removed$`, w.doctorReportsRemoved)
	ctx.Step(`^the file "([^"]*)" no longer exists$`, w.doctorFileNoLongerExists)
	ctx.Step(`^the doctor reports the claim data as cleared$`, w.doctorReportsCleared)
	ctx.Step(`^"([^"]*)" has no claim data$`, w.doctorTaskHasNoClaimData)
	ctx.Step(`^the doctor reports the stale claim as released$`, w.doctorReportsReleased)
	ctx.Step(`^the doctor reports the "([^"]*)" issue for "([^"]*)" as not fixable$`, w.doctorReportsNotFixable)
}

// --- fixture construction -------------------------------------------------

// doctorValidTask returns a task that passes every doctor check, as a base for
// fixtures that introduce a single defect.
func doctorValidTask(id string) *task.Task {
	return &task.Task{
		ID:         id,
		Title:      id,
		Status:     "open",
		Priority:   "medium",
		Type:       "feature",
		CreatedAt:  fixtureTime,
		UpdatedAt:  fixtureTime,
		CreatedBy:  "human",
		DependsOn:  []string{},
		Tags:       []string{},
		References: []string{},
	}
}

func (w *world) doctorTaskWithTitleAndStatus(id, title, status string) error {
	t := doctorValidTask(id)
	t.Title = title
	t.Status = status
	return writeFixtureTask(t)
}

func (w *world) doctorInvalidYAMLFile(name string) error {
	// Valid frontmatter delimiters wrapping content that is not valid YAML.
	content := "---\ntitle: [unterminated\nstatus: : :\n---\n"
	return os.WriteFile(filepath.Join(".tl", "tasks", name), []byte(content), 0o644)
}

func (w *world) doctorFileMissingField(name, field string) error {
	id := strings.TrimSuffix(name, ".md")
	t := doctorValidTask(id)
	data, err := t.MarshalMarkdown()
	if err != nil {
		return err
	}
	// Drop the requested frontmatter line entirely.
	var kept []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, field+":") {
			continue
		}
		kept = append(kept, line)
	}
	return os.WriteFile(filepath.Join(".tl", "tasks", name), []byte(strings.Join(kept, "\n")), 0o644)
}

func (w *world) doctorTaskWithPriority(id, priority string) error {
	t := doctorValidTask(id)
	t.Priority = priority
	return writeFixtureTask(t)
}

func (w *world) doctorTaskWithEmptyType(id string) error {
	t := doctorValidTask(id)
	t.Type = ""
	return writeFixtureTask(t)
}

func (w *world) doctorSecondFileWithID(id string) error {
	// The first file is created by an earlier step; write a second file with a
	// different name but the same frontmatter id.
	t := doctorValidTask(id)
	t.ID = id
	data, err := t.MarshalMarkdown()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(".tl", "tasks", id+"-copy.md"), data, 0o644)
}

func (w *world) doctorTaskWithDependency(id, dep string) error {
	t := doctorValidTask(id)
	t.DependsOn = []string{dep}
	return writeFixtureTask(t)
}

func (w *world) doctorEventReferencingTask(id string) error {
	return recordFixtureEvent("created", id, "human", fixtureTime)
}

func (w *world) doctorNoTaskFileFor(id string) error {
	// Ensure absence; remove if a prior step created it.
	_ = os.Remove(filepath.Join(".tl", "tasks", id+".md"))
	return nil
}

func (w *world) doctorInProgressNoClaim(id string) error {
	t := doctorValidTask(id)
	t.Status = "in_progress"
	return writeFixtureTask(t)
}

func (w *world) doctorClaimedExpiredLease(id, actor string) error {
	t := doctorValidTask(id)
	t.Status = "in_progress"
	claimed := time.Now().Add(-3 * time.Hour).UTC().Truncate(time.Second)
	expires := time.Now().Add(-2 * time.Hour).UTC().Truncate(time.Second)
	t.CreatedAt = claimed
	t.UpdatedAt = claimed
	t.Claim = task.Claim{Actor: &actor, ClaimedAt: &claimed, ExpiresAt: &expires}
	return writeFixtureTask(t)
}

func (w *world) doctorOpenWithClaimData(id string) error {
	t := doctorValidTask(id)
	actor := "claude-code:main"
	claimed := time.Now().Add(-30 * time.Minute).UTC().Truncate(time.Second)
	expires := time.Now().Add(30 * time.Minute).UTC().Truncate(time.Second)
	t.Claim = task.Claim{Actor: &actor, ClaimedAt: &claimed, ExpiresAt: &expires}
	return writeFixtureTask(t)
}

func (w *world) doctorCreatedInFuture(id string) error {
	t := doctorValidTask(id)
	future := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	t.CreatedAt = future
	t.UpdatedAt = future
	return writeFixtureTask(t)
}

func (w *world) doctorCreatedAfterUpdated(id string) error {
	t := doctorValidTask(id)
	t.CreatedAt = fixtureTime.Add(time.Hour)
	t.UpdatedAt = fixtureTime
	return writeFixtureTask(t)
}

func (w *world) doctorExpiryBeforeClaimed(id string) error {
	t := doctorValidTask(id)
	claimed := fixtureTime
	expires := fixtureTime.Add(-time.Hour)
	// Actor left nil so this isolates the timestamp anomaly from the claims check.
	t.Claim = task.Claim{ClaimedAt: &claimed, ExpiresAt: &expires}
	return writeFixtureTask(t)
}

func (w *world) doctorOrphanedTmp(name string) error {
	return os.WriteFile(filepath.Join(".tl", "tasks", name), []byte("interrupted write"), 0o644)
}

func (w *world) doctorUnreadableFile(name string) error {
	// A dangling symlink reproduces an unreadable file deterministically,
	// independent of the test user's privileges.
	target := filepath.Join(".tl", "tasks", "nonexistent-target")
	return os.Symlink(target, filepath.Join(".tl", "tasks", name))
}

func (w *world) doctorBodyContains(id, marker string) error {
	t := doctorValidTask(id)
	t.Body = "## Description\n\n" + marker + "\n"
	return writeFixtureTask(t)
}

func (w *world) doctorMalformedNotes(id string) error {
	t := doctorValidTask(id)
	t.Body = "## Notes\n\n- this is not a canonical note line\n"
	return writeFixtureTask(t)
}

func (w *world) doctorNoConfig() error {
	return os.Remove(filepath.Join(".tl", "config.yaml"))
}

func (w *world) doctorInvalidConfig() error {
	return os.WriteFile(filepath.Join(".tl", "config.yaml"), []byte("::: not: valid: yaml\n  - broken"), 0o644)
}

func (w *world) doctorLedgerWithTasks(n int) error {
	for i := 0; i < n; i++ {
		t := doctorValidTask(fmt.Sprintf("task-s%04d", i))
		if err := writeFixtureTask(t); err != nil {
			return err
		}
	}
	return nil
}

func (w *world) doctorLedgerWithEvents(n int) error {
	for i := 0; i < n; i++ {
		if err := recordFixtureEvent("noted", "task-s001", "human", fixtureTime); err != nil {
			return err
		}
	}
	return nil
}

// --- assertions: report ---------------------------------------------------

func doctorDiagnose() ([]doctor.Diagnostic, error) {
	return doctor.Diagnose(".tl")
}

func (w *world) doctorReportsNoIssues() error {
	diags, err := doctorDiagnose()
	if err != nil {
		return err
	}
	if len(diags) != 0 {
		return fmt.Errorf("expected no issues, got %d: %+v", len(diags), diags)
	}
	return nil
}

func (w *world) doctorReportsIssueFor(category, taskID, severity string) error {
	diags, err := doctorDiagnose()
	if err != nil {
		return err
	}
	for _, d := range diags {
		if d.Category == category && d.TaskID == taskID && d.Severity == severity {
			return nil
		}
	}
	return fmt.Errorf("no %q issue for %q with severity %q; got: %+v", category, taskID, severity, diags)
}

func (w *world) doctorReportsIssue(category, severity string) error {
	diags, err := doctorDiagnose()
	if err != nil {
		return err
	}
	for _, d := range diags {
		if d.Category == category && d.Severity == severity {
			return nil
		}
	}
	return fmt.Errorf("no %q issue with severity %q; got: %+v", category, severity, diags)
}

func (w *world) doctorReportsScaleWarning(dimension, severity string) error {
	diags, err := doctorDiagnose()
	if err != nil {
		return err
	}
	for _, d := range diags {
		if d.Category == doctor.CategoryScale && d.Severity == severity && strings.Contains(d.Message, dimension) {
			return nil
		}
	}
	return fmt.Errorf("no scale warning about %q with severity %q; got: %+v", dimension, severity, diags)
}

// --- assertions: JSON -----------------------------------------------------

// doctorJSON runs `tl doctor --json` in-process and returns the parsed array.
func doctorJSON() ([]map[string]json.RawMessage, []byte, error) {
	root := cmd.NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"doctor", "--json"})
	if err := root.Execute(); err != nil {
		return nil, buf.Bytes(), err
	}
	var arr []map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &arr); err != nil {
		return nil, buf.Bytes(), fmt.Errorf("output is not a JSON array (%v); got: %s", err, buf.String())
	}
	return arr, buf.Bytes(), nil
}

func (w *world) doctorJSONEmptyArray() error {
	arr, raw, err := doctorJSON()
	if err != nil {
		return err
	}
	if len(arr) != 0 {
		return fmt.Errorf("expected empty array, got: %s", raw)
	}
	return nil
}

func (w *world) doctorJSONDiagnosticArray() error {
	arr, raw, err := doctorJSON()
	if err != nil {
		return err
	}
	if len(arr) == 0 {
		return fmt.Errorf("expected a non-empty diagnostic array, got: %s", raw)
	}
	return nil
}

func (w *world) doctorJSONHasCategorySeverity(category, severity string) error {
	arr, _, err := doctorJSON()
	if err != nil {
		return err
	}
	for _, obj := range arr {
		if jsonStr(obj["category"]) == category && jsonStr(obj["severity"]) == severity {
			return nil
		}
	}
	return fmt.Errorf("no diagnostic object with category %q and severity %q", category, severity)
}

func (w *world) doctorJSONHasNonEmptyField(field string) error {
	arr, _, err := doctorJSON()
	if err != nil {
		return err
	}
	for _, obj := range arr {
		if jsonStr(obj[field]) != "" {
			return nil
		}
	}
	return fmt.Errorf("no diagnostic object with a non-empty %q field", field)
}

func (w *world) doctorJSONHasBoolField(field string) error {
	arr, _, err := doctorJSON()
	if err != nil {
		return err
	}
	for _, obj := range arr {
		raw, ok := obj[field]
		if !ok {
			continue
		}
		var b bool
		if err := json.Unmarshal(raw, &b); err == nil {
			return nil
		}
	}
	return fmt.Errorf("no diagnostic object with a boolean %q field", field)
}

func jsonStr(raw json.RawMessage) string {
	var s string
	_ = json.Unmarshal(raw, &s)
	return s
}

// --- assertions: --fix ----------------------------------------------------

func (w *world) doctorReportsFixed(category, taskID string) error {
	return w.assertFixOutput("fixed", category, taskID)
}

func (w *world) doctorReportsRemoved() error {
	out := w.stdout.String()
	if !strings.Contains(out, "removed") || !strings.Contains(out, doctor.CategoryFilesystem) {
		return fmt.Errorf("fix output does not report a removed filesystem issue; got:\n%s", out)
	}
	return nil
}

func (w *world) doctorReportsCleared() error {
	out := w.stdout.String()
	if !strings.Contains(out, "cleared") || !strings.Contains(out, doctor.CategoryClaims) {
		return fmt.Errorf("fix output does not report cleared claim data; got:\n%s", out)
	}
	return nil
}

func (w *world) doctorReportsReleased() error {
	out := w.stdout.String()
	if !strings.Contains(out, "released") || !strings.Contains(out, doctor.CategoryClaims) {
		return fmt.Errorf("fix output does not report a released claim; got:\n%s", out)
	}
	return nil
}

func (w *world) doctorReportsNotFixable(category, taskID string) error {
	out := w.stdout.String()
	if !strings.Contains(out, "not fixable") || !strings.Contains(out, category) || !strings.Contains(out, taskID) {
		return fmt.Errorf("fix output does not report %q/%q as not fixable; got:\n%s", category, taskID, out)
	}
	return nil
}

func (w *world) assertFixOutput(verb, category, taskID string) error {
	out := w.stdout.String()
	if !strings.Contains(out, verb) || !strings.Contains(out, category) || !strings.Contains(out, taskID) {
		return fmt.Errorf("fix output does not report %q for %q/%q; got:\n%s", verb, category, taskID, out)
	}
	return nil
}

func (w *world) doctorNoLongerDependsOn(id, dep string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if containsString(t.DependsOn, dep) {
		return fmt.Errorf("%s still depends on %s: %v", id, dep, t.DependsOn)
	}
	return nil
}

func (w *world) doctorFileNoLongerExists(name string) error {
	if _, err := os.Stat(filepath.Join(".tl", "tasks", name)); !os.IsNotExist(err) {
		return fmt.Errorf("file %s still exists", name)
	}
	return nil
}

func (w *world) doctorTaskHasNoClaimData(id string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Claim.Actor != nil {
		return fmt.Errorf("%s still has claim data: %+v", id, t.Claim)
	}
	return nil
}

var _ = events.Event{} // keep events import for parity with sibling step files
