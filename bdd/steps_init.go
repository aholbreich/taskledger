package bdd

import (
	"fmt"
	"github.com/cucumber/godog"
	"os"
	"strings"

	"path/filepath"
)

// --- init.feature support -------------------------------------------------

func initializeInitSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^the current directory has no TaskLedger ledger$`, w.currentDirHasNoLedger)
	ctx.Step(`^the current directory already has a TaskLedger ledger$`, w.currentDirAlreadyHasLedger)
	ctx.Step(`^the directory contains a TaskLedger config file$`, w.dirContainsConfigFile)
	ctx.Step(`^the directory contains an empty tasks folder$`, w.dirContainsEmptyTasksFolder)
	ctx.Step(`^the directory contains an empty event journal$`, w.dirContainsEmptyEventJournal)
	ctx.Step(`^the command reports that the ledger already exists$`, w.cmdReportsAlreadyExists)
	ctx.Step(`^the existing config file is unchanged$`, w.existingConfigUnchanged)
}

const sentinelConfig = "# existing config — DO NOT TOUCH\n"

func (w *world) currentDirHasNoLedger() error {
	if _, err := os.Stat(".taskledger"); err == nil {
		return fmt.Errorf(".taskledger already exists in fresh temp dir (setup bug)")
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (w *world) currentDirAlreadyHasLedger() error {
	if err := os.MkdirAll(filepath.Join(".taskledger", "tasks"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(".taskledger", "config.yaml"), []byte(sentinelConfig), 0o644)
}

func (w *world) dirContainsConfigFile() error {
	info, err := os.Stat(filepath.Join(".taskledger", "config.yaml"))
	if err != nil {
		return fmt.Errorf("config file missing: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("config file is empty")
	}
	return nil
}

func (w *world) dirContainsEmptyTasksFolder() error {
	entries, err := os.ReadDir(filepath.Join(".taskledger", "tasks"))
	if err != nil {
		return fmt.Errorf("tasks folder missing: %w", err)
	}
	if len(entries) != 0 {
		return fmt.Errorf("tasks folder is not empty (%d entries)", len(entries))
	}
	return nil
}

func (w *world) dirContainsEmptyEventJournal() error {
	info, err := os.Stat(filepath.Join(".taskledger", "events.jsonl"))
	if err != nil {
		return fmt.Errorf("events journal missing: %w", err)
	}
	if info.Size() != 0 {
		return fmt.Errorf("events journal is not empty (%d bytes)", info.Size())
	}
	return nil
}

func (w *world) cmdReportsAlreadyExists() error {
	combined := strings.ToLower(w.stdout.String() + w.stderr.String())
	if w.cmdErr != nil {
		combined += " " + strings.ToLower(w.cmdErr.Error())
	}
	if w.cmdErr == nil {
		return fmt.Errorf("expected init to fail, but it succeeded; output: %s", combined)
	}
	if !strings.Contains(combined, "already") {
		return fmt.Errorf("expected output to mention 'already', got: %s", combined)
	}
	return nil
}

func (w *world) existingConfigUnchanged() error {
	data, err := os.ReadFile(filepath.Join(".taskledger", "config.yaml"))
	if err != nil {
		return err
	}
	if string(data) != sentinelConfig {
		return fmt.Errorf("config file was modified; content: %q", string(data))
	}
	return nil
}
