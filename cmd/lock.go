package cmd

import (
	"github.com/aholbreich/tl/internal/lock"
)

// acquireLock wraps lock.Acquire so cmd-layer callers see lock contention as
// ExitError code 7 (matching the reserved code in the README and PRD).
//
// Usage:
//
//	release, err := acquireLock(ledger)
//	if err != nil {
//	    return err
//	}
//	defer release()
func acquireLock(ledger string) (func() error, error) {
	release, err := lock.Acquire(ledger)
	if err != nil {
		return nil, NewExitError(7, "%v", err)
	}
	return release, nil
}
