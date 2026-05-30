package doctor

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aholbreich/tl/internal/repo"
)

// newLedger creates an initialized ledger in a temp dir and returns its path.
func newLedger(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	ledger, err := repo.Init(dir)
	if err != nil {
		t.Fatalf("init ledger: %v", err)
	}
	return ledger
}

func writeTask(t *testing.T, ledger, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(ledger, repo.TasksDir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write task %s: %v", name, err)
	}
}

// cleanTask is a minimal valid task file body.
func cleanTask(id string) string {
	return "---\n" +
		"id: " + id + "\n" +
		"title: Valid task\n" +
		"status: open\n" +
		"priority: medium\n" +
		"type: feature\n" +
		"created_at: 2026-01-01T00:00:00Z\n" +
		"updated_at: 2026-01-01T00:00:00Z\n" +
		"created_by: human\n" +
		"assignee: null\n" +
		"depends_on: []\n" +
		"claim:\n  actor: null\n  claimed_at: null\n  expires_at: null\n  heartbeat_at: null\n" +
		"tags: []\n" +
		"---\n"
}

func countCategory(diags []Diagnostic, category string) int {
	n := 0
	for _, d := range diags {
		if d.Category == category {
			n++
		}
	}
	return n
}

func findFor(diags []Diagnostic, category, taskID string) *Diagnostic {
	for i := range diags {
		if diags[i].Category == category && diags[i].TaskID == taskID {
			return &diags[i]
		}
	}
	return nil
}

func TestCleanLedgerHasNoIssues(t *testing.T) {
	ledger := newLedger(t)
	writeTask(t, ledger, "task-abc.md", cleanTask("task-abc"))

	diags, err := Diagnose(ledger)
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected 0 issues on a clean ledger, got %d: %+v", len(diags), diags)
	}
}

func TestFrontmatterChecks(t *testing.T) {
	ledger := newLedger(t)
	bad := "---\nid: task-bad\ntitle: \"\"\nstatus: super-duper\npriority: urgent\ntype: \"\"\n" +
		"created_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\ncreated_by: human\n" +
		"assignee: null\ndepends_on: []\nclaim:\n  actor: null\n  claimed_at: null\n  expires_at: null\n  heartbeat_at: null\ntags: []\n---\n"
	writeTask(t, ledger, "task-bad.md", bad)

	diags, err := Diagnose(ledger)
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}
	// Expect exactly the four frontmatter errors: title, status, priority, type.
	if got := countCategory(diags, CategoryFrontmatter); got != 4 {
		t.Fatalf("expected 4 frontmatter issues, got %d: %+v", got, diags)
	}
	d := findFor(diags, CategoryFrontmatter, "task-bad")
	if d == nil || d.Severity != SeverityError {
		t.Fatalf("expected frontmatter error for task-bad, got %+v", d)
	}
}

func TestSelfDependencyIsFixable(t *testing.T) {
	ledger := newLedger(t)
	selfDep := "---\nid: task-abc\ntitle: T\nstatus: open\npriority: medium\ntype: feature\n" +
		"created_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\ncreated_by: human\n" +
		"assignee: null\ndepends_on:\n  - task-abc\nclaim:\n  actor: null\n  claimed_at: null\n  expires_at: null\n  heartbeat_at: null\ntags: []\n---\n"
	writeTask(t, ledger, "task-abc.md", selfDep)

	diags, _ := Diagnose(ledger)
	d := findFor(diags, CategoryDependency, "task-abc")
	if d == nil || !d.Fixable {
		t.Fatalf("expected fixable self-dependency, got %+v", d)
	}

	applied, _, err := Fix(ledger)
	if err != nil {
		t.Fatalf("Fix: %v", err)
	}
	if len(applied) != 1 || applied[0].Verb != "fixed" {
		t.Fatalf("expected one 'fixed' repair, got %+v", applied)
	}
	// After fix, no dependency issue remains.
	diags2, _ := Diagnose(ledger)
	if findFor(diags2, CategoryDependency, "task-abc") != nil {
		t.Fatalf("self-dependency still present after fix: %+v", diags2)
	}
}

func TestCyclicDependencyReportsEveryNode(t *testing.T) {
	ledger := newLedger(t)
	mk := func(id, dep string) string {
		return "---\nid: " + id + "\ntitle: T\nstatus: open\npriority: medium\ntype: feature\n" +
			"created_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\ncreated_by: human\n" +
			"assignee: null\ndepends_on:\n  - " + dep + "\nclaim:\n  actor: null\n  claimed_at: null\n  expires_at: null\n  heartbeat_at: null\ntags: []\n---\n"
	}
	writeTask(t, ledger, "task-a.md", mk("task-a", "task-b"))
	writeTask(t, ledger, "task-b.md", mk("task-b", "task-a"))

	diags, _ := Diagnose(ledger)
	if findFor(diags, CategoryDependency, "task-a") == nil || findFor(diags, CategoryDependency, "task-b") == nil {
		t.Fatalf("expected dependency cycle reported for both task-a and task-b: %+v", diags)
	}
}

func TestOrphanedTmpIsFixableWarning(t *testing.T) {
	ledger := newLedger(t)
	writeTask(t, ledger, "task-abc.md", cleanTask("task-abc"))
	writeTask(t, ledger, "task-abc.md.tmp", "junk")

	diags, _ := Diagnose(ledger)
	d := findFor(diags, CategoryFilesystem, "")
	if d == nil || d.Severity != SeverityWarning || !d.Fixable {
		t.Fatalf("expected fixable filesystem warning for orphan tmp, got %+v", d)
	}

	if _, _, err := Fix(ledger); err != nil {
		t.Fatalf("Fix: %v", err)
	}
	if _, err := os.Stat(filepath.Join(ledger, repo.TasksDir, "task-abc.md.tmp")); !os.IsNotExist(err) {
		t.Fatalf("orphan tmp not removed by fix")
	}
}

func TestExpiredClaimIsReleased(t *testing.T) {
	ledger := newLedger(t)
	past := time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)
	expired := "---\nid: task-abc\ntitle: T\nstatus: in_progress\npriority: medium\ntype: feature\n" +
		"created_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\ncreated_by: human\n" +
		"assignee: null\ndepends_on: []\nclaim:\n  actor: agent\n  claimed_at: 2026-01-01T00:00:00Z\n  expires_at: " + past + "\n  heartbeat_at: null\ntags: []\n---\n"
	writeTask(t, ledger, "task-abc.md", expired)

	diags, _ := Diagnose(ledger)
	d := findFor(diags, CategoryClaims, "task-abc")
	if d == nil || d.Severity != SeverityWarning || !d.Fixable {
		t.Fatalf("expected fixable claims warning for expired claim, got %+v", d)
	}

	if _, _, err := Fix(ledger); err != nil {
		t.Fatalf("Fix: %v", err)
	}
	diags2, _ := Diagnose(ledger)
	if findFor(diags2, CategoryClaims, "task-abc") != nil {
		t.Fatalf("expired claim still present after fix: %+v", diags2)
	}
}

func TestOpenTaskWithClaimDataCleared(t *testing.T) {
	ledger := newLedger(t)
	future := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	openClaim := "---\nid: task-abc\ntitle: T\nstatus: open\npriority: medium\ntype: feature\n" +
		"created_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\ncreated_by: human\n" +
		"assignee: null\ndepends_on: []\nclaim:\n  actor: agent\n  claimed_at: 2026-01-01T00:00:00Z\n  expires_at: " + future + "\n  heartbeat_at: null\ntags: []\n---\n"
	writeTask(t, ledger, "task-abc.md", openClaim)

	diags, _ := Diagnose(ledger)
	d := findFor(diags, CategoryClaims, "task-abc")
	if d == nil || d.Severity != SeverityWarning || !d.Fixable {
		t.Fatalf("expected fixable claims warning for open task with claim data, got %+v", d)
	}
}

func TestInProgressWithoutClaimIsError(t *testing.T) {
	ledger := newLedger(t)
	noClaim := "---\nid: task-abc\ntitle: T\nstatus: in_progress\npriority: medium\ntype: feature\n" +
		"created_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\ncreated_by: human\n" +
		"assignee: null\ndepends_on: []\nclaim:\n  actor: null\n  claimed_at: null\n  expires_at: null\n  heartbeat_at: null\ntags: []\n---\n"
	writeTask(t, ledger, "task-abc.md", noClaim)

	diags, _ := Diagnose(ledger)
	d := findFor(diags, CategoryClaims, "task-abc")
	if d == nil || d.Severity != SeverityError || d.Fixable {
		t.Fatalf("expected non-fixable claims error, got %+v", d)
	}
}

func TestUnparseableFileIsFilesystemOrFrontmatterError(t *testing.T) {
	ledger := newLedger(t)
	writeTask(t, ledger, "task-bad.md", "this is not a task file at all")

	diags, _ := Diagnose(ledger)
	d := findFor(diags, CategoryFrontmatter, "task-bad")
	if d == nil || d.Severity != SeverityError || d.Fixable {
		t.Fatalf("expected non-fixable frontmatter error for unparseable file, got %+v", d)
	}
}

func TestMergeConflictMarkersInBody(t *testing.T) {
	ledger := newLedger(t)
	body := cleanTask("task-abc") + "\n## Description\n\n<<<<<<< HEAD\nfoo\n=======\nbar\n>>>>>>> branch\n"
	writeTask(t, ledger, "task-abc.md", body)

	diags, _ := Diagnose(ledger)
	d := findFor(diags, CategoryBody, "task-abc")
	if d == nil || d.Severity != SeverityError {
		t.Fatalf("expected body error for merge markers, got %+v", d)
	}
}

func TestOrphanedEventDetected(t *testing.T) {
	ledger := newLedger(t)
	writeTask(t, ledger, "task-abc.md", cleanTask("task-abc"))
	journal := filepath.Join(ledger, repo.EventsJournal)
	line := `{"time":"2026-01-01T00:00:00Z","event":"created","task_id":"task-ghost"}` + "\n"
	if err := os.WriteFile(journal, []byte(line), 0o644); err != nil {
		t.Fatalf("write journal: %v", err)
	}

	diags, _ := Diagnose(ledger)
	d := findFor(diags, CategoryEvents, "task-ghost")
	if d == nil || d.Severity != SeverityError {
		t.Fatalf("expected events error for orphaned event, got %+v", d)
	}
}

func TestDeadFileReferenceIsFixable(t *testing.T) {
	ledger := newLedger(t)
	withRef := "---\nid: task-abc\ntitle: T\nstatus: open\npriority: medium\ntype: feature\n" +
		"created_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\ncreated_by: human\n" +
		"assignee: null\ndepends_on: []\nclaim:\n  actor: null\n  claimed_at: null\n  expires_at: null\n  heartbeat_at: null\ntags: []\n" +
		"references:\n  - src/missing/file.go\n  - https://example.com/pr/1\n  - JIRA-123\n---\n"
	writeTask(t, ledger, "task-abc.md", withRef)

	diags, _ := Diagnose(ledger)
	if got := countCategory(diags, CategoryReferences); got != 1 {
		t.Fatalf("expected exactly 1 references issue (dead path only; URL and bare id skipped), got %d: %+v", got, diags)
	}
	d := findFor(diags, CategoryReferences, "task-abc")
	if d == nil || d.Severity != SeverityWarning || !d.Fixable {
		t.Fatalf("expected fixable references warning, got %+v", d)
	}

	if _, _, err := Fix(ledger); err != nil {
		t.Fatalf("Fix: %v", err)
	}
	diags2, _ := Diagnose(ledger)
	if findFor(diags2, CategoryReferences, "task-abc") != nil {
		t.Fatalf("dead reference still present after fix: %+v", diags2)
	}
}

func TestExistingReferenceFileIsClean(t *testing.T) {
	ledger := newLedger(t)
	// Create a real file relative to the repo root (parent of .tl).
	repoRoot := filepath.Dir(ledger)
	if err := os.MkdirAll(filepath.Join(repoRoot, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "src", "real.go"), []byte("package src"), 0o644); err != nil {
		t.Fatal(err)
	}
	withRef := "---\nid: task-abc\ntitle: T\nstatus: open\npriority: medium\ntype: feature\n" +
		"created_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\ncreated_by: human\n" +
		"assignee: null\ndepends_on: []\nclaim:\n  actor: null\n  claimed_at: null\n  expires_at: null\n  heartbeat_at: null\ntags: []\n" +
		"references:\n  - src/real.go\n---\n"
	writeTask(t, ledger, "task-abc.md", withRef)

	diags, _ := Diagnose(ledger)
	if countCategory(diags, CategoryReferences) != 0 {
		t.Fatalf("expected no references issue for existing file, got %+v", diags)
	}
}

func TestDuplicateIDsDetected(t *testing.T) {
	ledger := newLedger(t)
	writeTask(t, ledger, "task-dup.md", cleanTask("task-dup"))
	writeTask(t, ledger, "task-dup-copy.md", cleanTask("task-dup"))

	diags, _ := Diagnose(ledger)
	d := findFor(diags, CategoryIdentity, "task-dup")
	if d == nil || d.Severity != SeverityError {
		t.Fatalf("expected identity error for duplicate id, got %+v", d)
	}
}
