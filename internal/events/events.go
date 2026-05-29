// Package events appends audit records to .tl/events.jsonl.
package events

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/aholbreich/tl/internal/repo"
)

// Event is a single line in events.jsonl.
type Event struct {
	Time   time.Time `json:"time"`
	Event  string    `json:"event"`
	TaskID string    `json:"task_id"`
	Actor  string    `json:"actor,omitempty"`
	// Value carries an event-specific payload, e.g. the reference string on
	// reference_added / reference_removed. Omitted when empty.
	Value string `json:"value,omitempty"`
}

// Append appends e to the event journal under ledger, stamping the current
// time if e.Time is zero.
func Append(ledger string, e Event) error {
	if e.Time.IsZero() {
		e.Time = time.Now().UTC().Truncate(time.Second)
	}
	p := filepath.Join(ledger, repo.EventsJournal)
	f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

// ReadAll reads every event from the ledger journal in append order.
func ReadAll(ledger string) ([]Event, error) {
	p := filepath.Join(ledger, repo.EventsJournal)
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
