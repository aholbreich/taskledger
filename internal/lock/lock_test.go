package lock

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestAcquireReleaseRoundTrip verifies the happy path: a single Acquire
// against a fresh ledger directory returns a release function with no error,
// and the release can be called without error.
func TestAcquireReleaseRoundTrip(t *testing.T) {
	dir := t.TempDir()

	release, err := Acquire(dir)
	if err != nil {
		t.Fatalf("first Acquire: %v", err)
	}
	if err := release(); err != nil {
		t.Fatalf("release: %v", err)
	}

	// After release a fresh Acquire should also succeed.
	release2, err := Acquire(dir)
	if err != nil {
		t.Fatalf("second Acquire after release: %v", err)
	}
	if err := release2(); err != nil {
		t.Fatalf("second release: %v", err)
	}

	// The lock file should have been created.
	if _, err := os.Stat(filepath.Join(dir, LockFile)); err != nil {
		t.Fatalf("lock file should exist after Acquire: %v", err)
	}
}

// TestCrossProcessContention verifies the lock actually excludes a second
// process. This is the property the whole package exists to provide; without
// this test we would only know the function can be called.
//
// Strategy: spawn a child Go process that acquires the lock and holds it for
// a known duration. From the parent, try to acquire with a short timeout.
// The parent's Acquire must fail.
func TestCrossProcessContention(t *testing.T) {
	if os.Getenv("TL_LOCK_CHILD") == "1" {
		// Child path: acquire and hold for 2 seconds, then exit cleanly.
		dir := os.Getenv("TL_LOCK_DIR")
		release, err := Acquire(dir)
		if err != nil {
			os.Stderr.WriteString("child Acquire failed: " + err.Error())
			os.Exit(2)
		}
		// Signal readiness by writing a sentinel file the parent polls for.
		if err := os.WriteFile(filepath.Join(dir, "child.ready"), nil, 0o644); err != nil {
			os.Exit(3)
		}
		time.Sleep(2 * time.Second)
		_ = release()
		os.Exit(0)
	}

	dir := t.TempDir()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("Executable: %v", err)
	}

	// Re-run this test binary, isolated to just this test, with env that
	// triggers the child path above.
	child := exec.Command(exe, "-test.run", "^TestCrossProcessContention$", "-test.v")
	child.Env = append(os.Environ(), "TL_LOCK_CHILD=1", "TL_LOCK_DIR="+dir)
	var childOut strings.Builder
	child.Stdout = &childOut
	child.Stderr = &childOut
	if err := child.Start(); err != nil {
		t.Fatalf("start child: %v", err)
	}
	t.Cleanup(func() { _ = child.Wait() })

	// Wait for the child to signal it has the lock.
	ready := filepath.Join(dir, "child.ready")
	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if time.Now().After(deadline) {
			_ = child.Process.Kill()
			t.Fatalf("child never signalled readiness; output:\n%s", childOut.String())
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Parent attempts Acquire with a short timeout — must fail because the
	// child is holding it.
	start := time.Now()
	_, err = AcquireWithTimeout(dir, 300*time.Millisecond)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatalf("parent Acquire succeeded while child held the lock (after %s)", elapsed)
	}
	if elapsed < 250*time.Millisecond {
		t.Fatalf("Acquire returned too quickly (%s); did it actually wait?", elapsed)
	}
}

// TestSequentialAcquiresInSameProcess covers the BDD-suite usage pattern:
// many commands run sequentially in the same process, each acquires and
// releases the lock. Each must succeed.
func TestSequentialAcquiresInSameProcess(t *testing.T) {
	dir := t.TempDir()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		i := i
		// Run sequentially via WaitGroup pattern; we just want each Acquire to
		// release cleanly before the next.
		go func() {
			defer wg.Done()
			release, err := Acquire(dir)
			if err != nil {
				t.Errorf("Acquire %d: %v", i, err)
				return
			}
			if err := release(); err != nil {
				t.Errorf("release %d: %v", i, err)
			}
		}()
		wg.Wait()
	}
}
