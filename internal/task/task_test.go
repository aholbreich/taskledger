package task

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestMarshalUnmarshalRoundtrip(t *testing.T) {
	actor := "claude-code:main"
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	orig := &Task{
		ID:        "task-abc123",
		Title:     "Add login validation",
		Status:    "open",
		Priority:  "high",
		Type:      "feature",
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: "human",
		Assignee:  nil,
		DependsOn: []string{"task-def456"},
		Claim: Claim{
			Actor:       &actor,
			ClaimedAt:   &now,
			ExpiresAt:   &now,
			HeartbeatAt: &now,
		},
		Tags: []string{"frontend", "auth"},
		Body: "## Description\n\nValidate email format.\n",
	}

	data, err := orig.MarshalMarkdown()
	if err != nil {
		t.Fatalf("MarshalMarkdown: %v", err)
	}

	parsed, err := UnmarshalMarkdown(data)
	if err != nil {
		t.Fatalf("UnmarshalMarkdown: %v", err)
	}

	if parsed.ID != orig.ID {
		t.Errorf("ID: got %q, want %q", parsed.ID, orig.ID)
	}
	if parsed.Title != orig.Title {
		t.Errorf("Title: got %q, want %q", parsed.Title, orig.Title)
	}
	if parsed.Status != orig.Status {
		t.Errorf("Status: got %q, want %q", parsed.Status, orig.Status)
	}
	if parsed.Priority != orig.Priority {
		t.Errorf("Priority: got %q, want %q", parsed.Priority, orig.Priority)
	}
	if parsed.Type != orig.Type {
		t.Errorf("Type: got %q, want %q", parsed.Type, orig.Type)
	}
	if len(parsed.DependsOn) != 1 || parsed.DependsOn[0] != "task-def456" {
		t.Errorf("DependsOn: got %v, want [task-def456]", parsed.DependsOn)
	}
	if parsed.Claim.Actor == nil || *parsed.Claim.Actor != "claude-code:main" {
		t.Errorf("Claim.Actor: got %v, want claude-code:main", parsed.Claim.Actor)
	}
	if len(parsed.Tags) != 2 {
		t.Errorf("Tags: got %v, want 2 tags", parsed.Tags)
	}
	if !parsed.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt: got %v, want %v", parsed.CreatedAt, now)
	}
}

func TestMarshalNoClaim(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	task := &Task{
		ID:        "task-min",
		Title:     "Minimal task",
		Status:    "open",
		Priority:  "medium",
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: "human",
		DependsOn: []string{},
		Tags:      []string{},
	}

	data, err := task.MarshalMarkdown()
	if err != nil {
		t.Fatalf("MarshalMarkdown: %v", err)
	}

	parsed, err := UnmarshalMarkdown(data)
	if err != nil {
		t.Fatalf("UnmarshalMarkdown: %v", err)
	}

	if parsed.ID != "task-min" {
		t.Errorf("ID: got %q", parsed.ID)
	}
	if parsed.Claim.Actor != nil {
		t.Errorf("Claim should be nil for minimal task, got %v", parsed.Claim.Actor)
	}
}

func TestMarshalWithBody(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	task := &Task{
		ID:        "task-body",
		Title:     "Body test",
		Status:    "open",
		Priority:  "medium",
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: "human",
		DependsOn: []string{},
		Tags:      []string{},
		Body:      "## Description\n\nSome description.\n\n## Notes\n\nSome note.\n",
	}

	data, err := task.MarshalMarkdown()
	if err != nil {
		t.Fatalf("MarshalMarkdown: %v", err)
	}

	parsed, err := UnmarshalMarkdown(data)
	if err != nil {
		t.Fatalf("UnmarshalMarkdown: %v", err)
	}

	if parsed.Body != task.Body {
		t.Errorf("Body: got %q, want %q", parsed.Body, task.Body)
	}
}

func TestUnmarshalMissingFrontmatter(t *testing.T) {
	_, err := UnmarshalMarkdown([]byte("not a task file"))
	if err == nil {
		t.Error("expected error for missing frontmatter")
	}
}

func TestSetDescriptionReplacesDescriptionAndPreservesNotes(t *testing.T) {
	body := "## Description\n\nOld description.\n\n## Notes\n\n- 2026-05-24T15:36:20Z [pi] note: Keep me.\n"

	got := SetDescription(body, "New description.")

	if !strings.Contains(got, "## Description\n\nNew description.") {
		t.Fatalf("SetDescription() did not replace description; got:\n%s", got)
	}
	if strings.Contains(got, "Old description.") {
		t.Fatalf("SetDescription() kept old description; got:\n%s", got)
	}
	if !strings.Contains(got, "## Notes\n\n- 2026-05-24T15:36:20Z [pi] note: Keep me.") {
		t.Fatalf("SetDescription() did not preserve notes; got:\n%s", got)
	}
}

func TestAppendNoteUsesCanonicalBulletFormat(t *testing.T) {
	when := time.Date(2026, 5, 24, 15, 36, 20, 0, time.UTC)
	body := "## Description\n\nSome description.\n"

	got := AppendNote(body, when, "pi:notes-format", "blocked", "Waiting\nfor input.")

	want := "## Notes\n\n- 2026-05-24T15:36:20Z [pi:notes-format] blocked: Waiting for input."
	if !strings.Contains(got, want) {
		t.Fatalf("AppendNote() missing canonical note %q; got:\n%s", want, got)
	}
}

func TestParseBodyExtractsDescriptionAndCanonicalNotes(t *testing.T) {
	body := "## Description\n\nSome description.\n\n## Notes\n\n- 2026-05-24T15:36:20Z [pi:notes-format] note: Verified locally.\n"

	parsed := ParseBody(body)

	if parsed.Description != "Some description." {
		t.Fatalf("Description = %q", parsed.Description)
	}
	if len(parsed.Notes) != 1 {
		t.Fatalf("Notes len = %d", len(parsed.Notes))
	}
	note := parsed.Notes[0]
	if note.Actor != "pi:notes-format" || note.Kind != "note" || note.Message != "Verified locally." {
		t.Fatalf("Note = %#v", note)
	}
	if !note.Time.Equal(time.Date(2026, 5, 24, 15, 36, 20, 0, time.UTC)) {
		t.Fatalf("Note time = %s", note.Time)
	}
}

func TestParseNotesSupportsLegacyHeadings(t *testing.T) {
	body := "## Notes\n\n### 2026-05-24T15:36:20Z - pi:legacy\n\nOld note.\n\n### 2026-05-24T15:40:00Z — cancelled\n\nNo longer needed.\n"

	notes := ParseBody(body).Notes

	if len(notes) != 2 {
		t.Fatalf("Notes len = %d", len(notes))
	}
	if notes[0].Actor != "pi:legacy" || notes[0].Kind != "note" || notes[0].Message != "Old note." {
		t.Fatalf("legacy actor note = %#v", notes[0])
	}
	if notes[1].Kind != "cancelled" || notes[1].Message != "No longer needed." {
		t.Fatalf("legacy lifecycle note = %#v", notes[1])
	}
}

func TestTaskJSONIncludesParsedBodyFields(t *testing.T) {
	when := time.Date(2026, 5, 24, 15, 36, 20, 0, time.UTC)
	task := &Task{
		ID:        "task-json",
		Title:     "JSON task",
		Status:    "open",
		Priority:  "medium",
		CreatedAt: when,
		UpdatedAt: when,
		CreatedBy: "human",
		DependsOn: []string{},
		Tags:      []string{},
		Body:      "## Description\n\nSome description.\n\n## Notes\n\n- 2026-05-24T15:36:20Z [pi:notes-format] note: Verified locally.\n",
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal JSON: %v", err)
	}
	if got["description"] != "Some description." {
		t.Fatalf("description = %#v", got["description"])
	}
	notes, ok := got["notes"].([]any)
	if !ok || len(notes) != 1 {
		t.Fatalf("notes = %#v", got["notes"])
	}
	note := notes[0].(map[string]any)
	if note["actor"] != "pi:notes-format" || note["kind"] != "note" || note["message"] != "Verified locally." {
		t.Fatalf("note = %#v", note)
	}
}
