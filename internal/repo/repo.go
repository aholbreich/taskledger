// Package repo manages the on-disk .taskledger layout: locating an existing
// ledger and creating a fresh one.
package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	LedgerDir     = ".taskledger"
	ConfigFile    = "config.yaml"
	TasksDir      = "tasks"
	EventsJournal = "events.jsonl"
)

// ErrAlreadyInitialized is returned by Init when a ledger directory already
// exists at the target location.
var ErrAlreadyInitialized = errors.New("ledger already exists")

const defaultConfig = `version: 1
default_claim_ttl: 60m
id_prefix: task
default_verify_commands: []
actors:
  require_actor: true
`

// Init creates the .taskledger layout under dir and returns the absolute path
// to the created ledger directory. It refuses to touch an existing ledger.
func Init(dir string) (string, error) {
	ledger := filepath.Join(dir, LedgerDir)

	if _, err := os.Stat(ledger); err == nil {
		return "", fmt.Errorf("%w at %s", ErrAlreadyInitialized, ledger)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	if err := os.MkdirAll(filepath.Join(ledger, TasksDir), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(ledger, ConfigFile), []byte(defaultConfig), 0o644); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(ledger, EventsJournal), nil, 0o644); err != nil {
		return "", err
	}
	return ledger, nil
}
