package task

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

// Note is a parsed entry from a task's ## Notes Markdown section.
type Note struct {
	Time    time.Time `json:"time"`
	Actor   string    `json:"actor,omitempty"`
	Kind    string    `json:"kind"`
	Message string    `json:"message"`
	Raw     string    `json:"raw,omitempty"`
}

// ParsedBody is the structured view of a task Markdown body.
type ParsedBody struct {
	Description string            `json:"description,omitempty"`
	Notes       []Note            `json:"notes,omitempty"`
	Sections    map[string]string `json:"sections,omitempty"`
}

var (
	notesHeadingRE = regexp.MustCompile(`(?m)^## Notes\s*$`)
	h2HeadingRE    = regexp.MustCompile(`(?m)^## `)
	noteLineRE     = regexp.MustCompile(`^-\s+(\S+)\s+\[([^\]]*)\]\s+([A-Za-z_][A-Za-z0-9_-]*):\s*(.*)$`)
)

// SetDescription replaces the ## Description section, inserts it before other
// body content if missing, or removes it when description is empty.
func SetDescription(body, description string) string {
	section := ""
	if description != "" {
		section = "## Description\n\n" + strings.TrimRight(description, "\n")
	}

	loc := regexp.MustCompile(`(?m)^## Description\s*$`).FindStringIndex(body)
	if loc == nil {
		if section == "" {
			return body
		}
		if strings.TrimSpace(body) == "" {
			return section + "\n"
		}
		return section + "\n\n" + strings.TrimLeft(body, "\n")
	}

	end := len(body)
	if next := h2HeadingRE.FindStringIndex(body[loc[1]:]); next != nil {
		end = loc[1] + next[0]
	}

	var parts []string
	if prefix := strings.TrimRight(body[:loc[0]], "\n"); prefix != "" {
		parts = append(parts, prefix)
	}
	if section != "" {
		parts = append(parts, section)
	}
	if suffix := strings.TrimLeft(body[end:], "\n"); suffix != "" {
		parts = append(parts, suffix)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n\n") + "\n"
}

// AppendNote appends a normalized one-line bullet note under ## Notes.
func AppendNote(body string, when time.Time, actor, kind, message string) string {
	trimmed := strings.TrimRight(body, "\n")
	line := FormatNote(when, actor, kind, message)

	loc := notesHeadingRE.FindStringIndex(trimmed)
	if loc == nil {
		if trimmed == "" {
			return "## Notes\n\n" + line + "\n"
		}
		return trimmed + "\n\n## Notes\n\n" + line + "\n"
	}

	insertAt := len(trimmed)
	if next := h2HeadingRE.FindStringIndex(trimmed[loc[1]:]); next != nil {
		insertAt = loc[1] + next[0]
	}

	prefix := strings.TrimRight(trimmed[:insertAt], "\n")
	suffix := trimmed[insertAt:]
	if suffix == "" {
		return prefix + "\n" + line + "\n"
	}
	return prefix + "\n" + line + "\n\n" + strings.TrimLeft(suffix, "\n")
}

// FormatNote returns the canonical Markdown representation for one note.
func FormatNote(when time.Time, actor, kind, message string) string {
	if kind == "" {
		kind = "note"
	}
	if actor == "" {
		actor = "unknown"
	}
	return "- " + when.UTC().Format(time.RFC3339) + " [" + actor + "] " + kind + ": " + normalizeNoteMessage(message)
}

func normalizeNoteMessage(message string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(message)), " ")
}

// ParseBody extracts conventional sections from a Markdown task body.
func ParseBody(body string) ParsedBody {
	sections := splitH2Sections(body)
	parsed := ParsedBody{}
	for heading, content := range sections {
		switch heading {
		case "Description":
			parsed.Description = content
		case "Notes":
			parsed.Notes = ParseNotes(content)
		case "":
			if content != "" {
				if parsed.Sections == nil {
					parsed.Sections = map[string]string{}
				}
				parsed.Sections["preamble"] = content
			}
		default:
			if parsed.Sections == nil {
				parsed.Sections = map[string]string{}
			}
			parsed.Sections[heading] = content
		}
	}
	return parsed
}

// ParseNotes extracts canonical bullet notes and legacy note headings.
func ParseNotes(notesSection string) []Note {
	var notes []Note
	lines := strings.Split(notesSection, "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if n, ok := parseCanonicalNoteLine(line); ok {
			notes = append(notes, n)
			continue
		}
		if strings.HasPrefix(line, "### ") {
			note, next := parseLegacyNote(lines, i)
			notes = append(notes, note)
			i = next - 1
		}
	}
	return notes
}

func parseCanonicalNoteLine(line string) (Note, bool) {
	m := noteLineRE.FindStringSubmatch(line)
	if m == nil {
		return Note{}, false
	}
	when, err := time.Parse(time.RFC3339, m[1])
	if err != nil {
		return Note{}, false
	}
	return Note{
		Time:    when,
		Actor:   strings.TrimSpace(m[2]),
		Kind:    strings.TrimSpace(m[3]),
		Message: strings.TrimSpace(m[4]),
	}, true
}

func parseLegacyNote(lines []string, start int) (Note, int) {
	heading := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[start]), "### "))
	end := start + 1
	var message []string
	for end < len(lines) {
		line := strings.TrimSpace(lines[end])
		if strings.HasPrefix(line, "### ") || strings.HasPrefix(line, "- ") {
			break
		}
		message = append(message, lines[end])
		end++
	}

	note := Note{Kind: "note", Message: strings.TrimSpace(strings.Join(message, "\n")), Raw: strings.TrimSpace(strings.Join(lines[start:end], "\n"))}
	if idx := strings.Index(heading, " - "); idx >= 0 {
		if when, err := time.Parse(time.RFC3339, strings.TrimSpace(heading[:idx])); err == nil {
			note.Time = when
			note.Actor = strings.TrimSpace(heading[idx+3:])
		}
		return note, end
	}
	if idx := strings.Index(heading, " — "); idx >= 0 {
		if when, err := time.Parse(time.RFC3339, strings.TrimSpace(heading[:idx])); err == nil {
			note.Time = when
			note.Kind = strings.TrimSpace(heading[idx+len(" — "):])
		}
		return note, end
	}
	if when, err := time.Parse(time.RFC3339, strings.TrimSpace(heading)); err == nil {
		note.Time = when
	}
	return note, end
}

func splitH2Sections(body string) map[string]string {
	sections := map[string]string{}
	current := ""
	var buf strings.Builder
	flush := func() {
		content := strings.Trim(buf.String(), "\n")
		if content != "" || current != "" {
			sections[current] = content
		}
		buf.Reset()
	}

	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
			flush()
			current = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			continue
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	flush()
	return sections
}

// MarshalJSON includes a parsed view of Body while keeping the raw body for
// compatibility with existing JSON consumers.
func (t Task) MarshalJSON() ([]byte, error) {
	type alias Task
	parsed := ParseBody(t.Body)
	out := struct {
		alias
		Description string            `json:"description,omitempty"`
		Notes       []Note            `json:"notes,omitempty"`
		Sections    map[string]string `json:"sections,omitempty"`
	}{
		alias:       alias(t),
		Description: parsed.Description,
		Notes:       parsed.Notes,
		Sections:    parsed.Sections,
	}
	return json.Marshal(out)
}
