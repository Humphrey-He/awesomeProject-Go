package result_wrapper

import (
	"errors"
	"fmt"
)

var ErrResultUnwrapOnFailure = errors.New("unwrap called on failure result")

// Result carries value and error in one typed container.
// ok=true means Value is valid and Err must be nil.
// ok=false means operation failed and Err is non-nil.
type Result[T any] struct {
	value T
	err   error
	ok    bool
}

// Success creates a successful Result.
func Success[T any](v T) Result[T] {
	return Result[T]{value: v, ok: true}
}

// Failure creates a failed Result.
// T's zero value is inferred automatically by compiler context.
func Failure[T any](err error) Result[T] {
	if err == nil {
		err = errors.New("failure requires non-nil error")
	}
	var zero T
	return Result[T]{value: zero, err: err, ok: false}
}

func (r Result[T]) IsSuccess() bool { return r.ok }
func (r Result[T]) IsFailure() bool { return !r.ok }
func (r Result[T]) Err() error      { return r.err }

// ValueOrZero returns value for success or zero value for failure.
func (r Result[T]) ValueOrZero() T {
	if r.ok {
		return r.value
	}
	var zero T
	return zero
}

// ValueOr returns fallback when failed.
func (r Result[T]) ValueOr(fallback T) T {
	if r.ok {
		return r.value
	}
	return fallback
}

// Unpack converts Result back to classic Go (value, error) pair.
func (r Result[T]) Unpack() (T, error) {
	if r.ok {
		return r.value, nil
	}
	var zero T
	return zero, r.err
}

// Must returns value or panics, intended for tests/bootstrap only.
func (r Result[T]) Must() T {
	if !r.ok {
		panic(fmt.Errorf("%w: %v", ErrResultUnwrapOnFailure, r.err))
	}
	return r.value
}

// Map transforms Result[T] to Result[U] on success and short-circuits on failure.
func Map[T any, U any](r Result[T], fn func(T) U) Result[U] {
	if r.IsFailure() {
		return Failure[U](r.Err())
	}
	return Success(fn(r.value))
}

// Bind (flatMap) chains operations returning Result.
func Bind[T any, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	if r.IsFailure() {
		return Failure[U](r.Err())
	}
	return fn(r.value)
}

// Tap executes side effect on success without changing value.
func Tap[T any](r Result[T], fn func(T)) Result[T] {
	if r.IsSuccess() {
		fn(r.value)
	}
	return r
}

// TapError executes side effect on failure without changing result.
func TapError[T any](r Result[T], fn func(error)) Result[T] {
	if r.IsFailure() {
		fn(r.err)
	}
	return r
}

// Recover converts failure to success via fallback function.
func Recover[T any](r Result[T], fallback func(error) T) Result[T] {
	if r.IsSuccess() {
		return r
	}
	return Success(fallback(r.err))
}

// From converts legacy (value, error) into Result.
func From[T any](v T, err error) Result[T] {
	if err != nil {
		return Failure[T](err)
	}
	return Success(v)
}

// WrapError enriches failure error context while keeping value type.
func WrapError[T any](r Result[T], format string, args ...any) Result[T] {
	if r.IsSuccess() {
		return r
	}
	msg := fmt.Sprintf(format, args...)
	return Failure[T](fmt.Errorf("%s: %w", msg, r.err))
}


