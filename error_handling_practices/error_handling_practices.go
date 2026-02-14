package error_handling_practices

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
)

// OpError keeps operation context and supports unwrapping.
type OpError struct {
	Op  string
	Err error
}

func (e *OpError) Error() string {
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *OpError) Unwrap() error {
	return e.Err
}

// ParseUserID shows early-return + wrapped error.
func ParseUserID(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("parse user id: %w", ErrInvalidInput)
	}
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parse user id(%q): %w", s, err)
	}
	if id <= 0 {
		return 0, fmt.Errorf("parse user id(%q): %w", s, ErrInvalidInput)
	}
	return id, nil
}

// FindUserName demonstrates sentinel + wrapping at boundary.
func FindUserName(store map[int]string, id int) (string, error) {
	name, ok := store[id]
	if !ok {
		return "", &OpError{Op: "find user name", Err: ErrNotFound}
	}
	return name, nil
}

// CloseFunc is dependency-injected close behavior for testing.
type CloseFunc func() error

// RunWithCleanup demonstrates cleanup error handling.
func RunWithCleanup(run func() error, closeFn CloseFunc) error {
	if run == nil || closeFn == nil {
		return fmt.Errorf("run with cleanup: %w", ErrInvalidInput)
	}

	runErr := run()
	closeErr := closeFn()

	// Go1.18 without errors.Join: preserve primary failure, include cleanup detail.
	if runErr != nil && closeErr != nil {
		return fmt.Errorf("run err=%v; cleanup err=%v", runErr, closeErr)
	}
	if runErr != nil {
		return runErr
	}
	if closeErr != nil {
		return fmt.Errorf("cleanup: %w", closeErr)
	}
	return nil
}


