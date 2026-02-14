package result_wrapper

import (
	"errors"
	"strings"
	"testing"
)

func parseIntPositive(s string) Result[int] {
	switch s {
	case "":
		return Failure[int](errors.New("empty"))
	case "1":
		return Success(1)
	case "2":
		return Success(2)
	default:
		return Failure[int](errors.New("not supported"))
	}
}

func reciprocal(v int) Result[float64] {
	if v == 0 {
		return Failure[float64](errors.New("divide by zero"))
	}
	return Success(1.0 / float64(v))
}

func TestSuccessFailureBasics(t *testing.T) {
	s := Success[string]("ok")
	if !s.IsSuccess() || s.IsFailure() || s.Err() != nil {
		t.Fatalf("unexpected success state")
	}

	f := Failure[string](errors.New("boom"))
	if !f.IsFailure() || f.IsSuccess() || f.Err() == nil {
		t.Fatalf("unexpected failure state")
	}
	if got := f.ValueOr("fallback"); got != "fallback" {
		t.Fatalf("fallback value mismatch: %s", got)
	}
}

func TestUnpackAndFrom(t *testing.T) {
	r := From(10, nil)
	v, err := r.Unpack()
	if err != nil || v != 10 {
		t.Fatalf("unexpected unpack: v=%d err=%v", v, err)
	}

	r2 := From(99, errors.New("x"))
	_, err = r2.Unpack()
	if err == nil {
		t.Fatalf("expected error from unpack")
	}
}

func TestFailureAutoZeroValueInference(t *testing.T) {
	r := Failure[map[string]int](errors.New("failed"))
	v, err := r.Unpack()
	if err == nil {
		t.Fatalf("expected failure")
	}
	if v != nil {
		t.Fatalf("zero value for map should be nil")
	}
}

func TestMapBindShortCircuit(t *testing.T) {
	// success path
	got := Bind(
		Map(parseIntPositive("2"), func(v int) int { return v * 10 }),
		func(v int) Result[float64] { return reciprocal(v) },
	)
	fv, err := got.Unpack()
	if err != nil || fv != 0.05 {
		t.Fatalf("unexpected chain result: %v %v", fv, err)
	}

	// failure short-circuit path
	called := false
	got2 := Bind(parseIntPositive(""), func(v int) Result[float64] {
		called = true
		return reciprocal(v)
	})
	if called {
		t.Fatalf("bind should short-circuit on failure")
	}
	if got2.Err() == nil {
		t.Fatalf("expected error")
	}
}

func TestTapTapErrorRecover(t *testing.T) {
	tapped := 0
	errTapped := ""

	r1 := Tap(Success(3), func(v int) { tapped = v })
	if r1.IsFailure() || tapped != 3 {
		t.Fatalf("tap on success failed")
	}

	r2 := TapError(Failure[int](errors.New("e1")), func(err error) { errTapped = err.Error() })
	if r2.IsSuccess() || errTapped != "e1" {
		t.Fatalf("tapError failed")
	}

	r3 := Recover(Failure[int](errors.New("broken")), func(err error) int { return 42 })
	if v, err := r3.Unpack(); err != nil || v != 42 {
		t.Fatalf("recover failed: %v %v", v, err)
	}
}

func TestWrapError(t *testing.T) {
	r := WrapError(Failure[int](errors.New("db down")), "get profile")
	if r.Err() == nil || !strings.Contains(r.Err().Error(), "get profile") {
		t.Fatalf("wrapped error missing context: %v", r.Err())
	}
}

func TestMustPanicsOnFailure(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("must should panic on failure")
		}
	}()
	_ = Failure[int](errors.New("x")).Must()
}


