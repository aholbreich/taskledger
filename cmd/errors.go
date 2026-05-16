package cmd

import (
	"errors"
	"fmt"
)

type ExitError struct {
	Code int
	Err  error
}

func NewExitError(code int, format string, args ...any) *ExitError {
	return &ExitError{
		Code: code,
		Err:  fmt.Errorf(format, args...),
	}
}

func (e *ExitError) Error() string {
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	return e.Err
}

func (e *ExitError) ExitCode() int {
	return e.Code
}

func ErrorExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr interface{ ExitCode() int }
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}
