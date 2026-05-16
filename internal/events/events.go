// Package events appends audit records to .taskledger/events.jsonl.
package events

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/aholbreich/taskledger/internal/repo"
)

// Event is a single line in events.jsonl.
type Event struct {
	Time   time.Time `json:"time"`
	Event  string    `json:"event"`
	TaskID string    `json:"task_id"`
	Actor  string    `json:"actor,omitempty"`
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
