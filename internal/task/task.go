// Package task models a TaskLedger task and its on-disk Markdown+YAML form.
package task

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Task is the in-memory representation of a single task file.
type Task struct {
	ID        string    `yaml:"id" json:"id"`
	Title     string    `yaml:"title" json:"title"`
	Status    string    `yaml:"status" json:"status"`
	Priority  string    `yaml:"priority" json:"priority"`
	Type      string    `yaml:"type,omitempty" json:"type,omitempty"`
	CreatedAt time.Time `yaml:"created_at" json:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at" json:"updated_at"`
	CreatedBy string    `yaml:"created_by" json:"created_by"`
	Assignee  *string   `yaml:"assignee" json:"assignee"`
	DependsOn []string  `yaml:"depends_on" json:"depends_on"`
	Claim     Claim     `yaml:"claim" json:"claim"`
	Verify    Verify    `yaml:"verify" json:"verify"`
	Tags      []string  `yaml:"tags" json:"tags"`

	// Body is the Markdown content after the YAML frontmatter. It is excluded
	// from the YAML encoder (the frontmatter never embeds the body) but is
	// included in JSON output so consumers see descriptions, notes, etc.
	Body string `yaml:"-" json:"body,omitempty"`
}

type Claim struct {
	Actor       *string    `yaml:"actor" json:"actor"`
	ClaimedAt   *time.Time `yaml:"claimed_at" json:"claimed_at"`
	ExpiresAt   *time.Time `yaml:"expires_at" json:"expires_at"`
	HeartbeatAt *time.Time `yaml:"heartbeat_at" json:"heartbeat_at"`
}

type Verify struct {
	Commands         []string `yaml:"commands" json:"commands"`
	EvidenceRequired []string `yaml:"evidence_required" json:"evidence_required"`
}

const frontmatterSep = "---"

// MarshalMarkdown serializes the task as a Markdown file with a YAML
// frontmatter block followed by the Body content.
func (t *Task) MarshalMarkdown() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(frontmatterSep)
	buf.WriteByte('\n')
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(t); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	buf.WriteString(frontmatterSep)
	buf.WriteByte('\n')
	if t.Body != "" {
		if !strings.HasPrefix(t.Body, "\n") {
			buf.WriteByte('\n')
		}
		buf.WriteString(t.Body)
		if !strings.HasSuffix(t.Body, "\n") {
			buf.WriteByte('\n')
		}
	}
	return buf.Bytes(), nil
}

// UnmarshalMarkdown parses a task file written by MarshalMarkdown.
func UnmarshalMarkdown(data []byte) (*Task, error) {
	s := string(data)
	if !strings.HasPrefix(s, frontmatterSep+"\n") {
		return nil, errors.New("task file missing frontmatter opener")
	}
	rest := s[len(frontmatterSep)+1:]
	end := strings.Index(rest, "\n"+frontmatterSep+"\n")
	bodyOffset := -1
	if end >= 0 {
		bodyOffset = end + len("\n"+frontmatterSep+"\n")
	} else {
		// Trailing closer without newline.
		end = strings.LastIndex(rest, "\n"+frontmatterSep)
		if end < 0 {
			return nil, errors.New("task file frontmatter not terminated")
		}
		bodyOffset = len(rest)
	}
	var t Task
	if err := yaml.Unmarshal([]byte(rest[:end]), &t); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}
	if bodyOffset < len(rest) {
		t.Body = strings.TrimPrefix(rest[bodyOffset:], "\n")
	}
	return &t, nil
}
