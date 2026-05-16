package cmd

import (
	"errors"
	"os"

	"github.com/aholbreich/taskledger/internal/store"
)

func requireLedger() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	ledger, err := store.LedgerDir(wd)
	if errors.Is(err, store.ErrLedgerNotFound) {
		return "", NewExitError(1, "TaskLedger is not initialized in this repository.\nRun `tl init` from the repository root to create .taskledger/.")
	}
	if err != nil {
		return "", err
	}
	return ledger, nil
}
