// Package store handles task-file I/O: locating the ledger, generating task
// IDs, and atomically writing task files.
package store

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aholbreich/taskledger/internal/repo"
	"github.com/aholbreich/taskledger/internal/task"
)

const (
	idCharset     = "abcdefghijklmnopqrstuvwxyz0123456789"
	idLen         = 3
	maxIDAttempts = 100
)

var (
	ErrLedgerNotFound = errors.New("ledger not found")
	ErrTaskNotFound   = errors.New("task not found")
)

// NewID returns a fresh task identifier of the form "task-<3 alphanumeric>".
// It retries on collision with an existing task file under ledger and gives up
// after maxIDAttempts (the 36^3 namespace fills up well before that matters
// for the project sizes TaskLedger targets).
func NewID(ledger string) (string, error) {
	for attempt := 0; attempt < maxIDAttempts; attempt++ {
		id, err := randomShortID()
		if err != nil {
			return "", err
		}
		_, err = os.Stat(TaskPath(ledger, id))
		if errors.Is(err, os.ErrNotExist) {
			return id, nil
		}
		if err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("could not generate a unique task id after %d attempts (id namespace too full)", maxIDAttempts)
}

func randomShortID() (string, error) {
	buf := make([]byte, idLen)
	max := big.NewInt(int64(len(idCharset)))
	for i := range buf {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		buf[i] = idCharset[n.Int64()]
	}
	return "task-" + string(buf), nil
}

// LedgerDir walks upward from start to find the nearest .taskledger directory.
func LedgerDir(start string) (string, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	cur := abs
	for {
		candidate := filepath.Join(cur, repo.LedgerDir)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", fmt.Errorf("%w: searched from %s", ErrLedgerNotFound, abs)
		}
		cur = parent
	}
}

// TaskPath returns the on-disk path for a task with the given id under ledger.
func TaskPath(ledger, id string) string {
	return filepath.Join(ledger, repo.TasksDir, id+".md")
}

// Write atomically writes the task to its file under ledger.
func Write(ledger string, t *task.Task) error {
	data, err := t.MarshalMarkdown()
	if err != nil {
		return err
	}
	p := TaskPath(ledger, t.ID)
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// Read loads one task by identifier from ledger.
func Read(ledger, id string) (*task.Task, error) {
	p := TaskPath(ledger, id)
	data, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}
	if err != nil {
		return nil, err
	}
	t, err := task.UnmarshalMarkdown(data)
	if err != nil {
		return nil, fmt.Errorf("parse task %s: %w", p, err)
	}
	return t, nil
}

// List reads every task file under ledger and returns tasks sorted by priority
// and then identifier.
func List(ledger string) ([]*task.Task, error) {
	dir := filepath.Join(ledger, repo.TasksDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	tasks := make([]*task.Task, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		p := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		t, err := task.UnmarshalMarkdown(data)
		if err != nil {
			return nil, fmt.Errorf("parse task %s: %w", p, err)
		}
		tasks = append(tasks, t)
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		left, right := priorityRank(tasks[i].Priority), priorityRank(tasks[j].Priority)
		if left != right {
			return left < right
		}
		return tasks[i].ID < tasks[j].ID
	})
	return tasks, nil
}

func priorityRank(priority string) int {
	switch priority {
	case "high":
		return 0
	case "medium":
		return 1
	case "low":
		return 2
	default:
		return 3
	}
}
