// Package lock provides an advisory file lock that mutating commands acquire
// before any read-modify-write on the ledger. Without it, two concurrent tl
// invocations can clobber each other's writes — defeating the claim-safety
// property the project promises.
package lock

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

const (
	// LockFile is the name of the lock file under the ledger directory.
	LockFile = ".lock"
	// DefaultTimeout is how long Acquire waits before giving up.
	DefaultTimeout = 5 * time.Second
	// pollInterval is how often Acquire retries while waiting on a held lock.
	pollInterval = 50 * time.Millisecond
)

// Acquire takes an exclusive flock on the ledger's lock file. The returned
// release function must be called (typically via defer) to release the lock.
// On Unix, the kernel also releases the lock automatically if the process
// exits without calling release.
//
// If another process holds the lock for longer than DefaultTimeout, Acquire
// returns an error; callers map it to exit code 7.
func Acquire(ledger string) (func() error, error) {
	return AcquireWithTimeout(ledger, DefaultTimeout)
}

// AcquireWithTimeout is Acquire with a caller-chosen wait limit. Exposed for
// tests; production code uses Acquire.
func AcquireWithTimeout(ledger string, timeout time.Duration) (func() error, error) {
	path := filepath.Join(ledger, LockFile)
	f := flock.New(path)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	locked, err := f.TryLockContext(ctx, pollInterval)
	if err != nil {
		return nil, fmt.Errorf("acquire ledger lock: %w", err)
	}
	if !locked {
		return nil, fmt.Errorf("ledger lock contention: another tl process held %s for more than %s", path, timeout)
	}
	return f.Unlock, nil
}
