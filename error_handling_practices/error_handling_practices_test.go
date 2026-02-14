package error_handling_practices

import (
	"errors"
	"testing"
)

func TestParseUserID(t *testing.T) {
	_, err := ParseUserID("")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input err, got %v", err)
	}
	id, err := ParseUserID("12")
	if err != nil || id != 12 {
		t.Fatalf("id=%d err=%v", id, err)
	}
}

func TestFindUserName(t *testing.T) {
	store := map[int]string{1: "alice"}
	_, err := FindUserName(store, 2)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	var opErr *OpError
	if !errors.As(err, &opErr) {
		t.Fatalf("expected OpError wrapper, got %T", err)
	}
}

func TestRunWithCleanup(t *testing.T) {
	err := RunWithCleanup(func() error { return nil }, func() error { return nil })
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	err = RunWithCleanup(func() error { return errors.New("biz fail") }, func() error { return nil })
	if err == nil {
		t.Fatal("expected run error")
	}

	err = RunWithCleanup(func() error { return nil }, func() error { return errors.New("close fail") })
	if err == nil || err.Error() == "close fail" {
		t.Fatalf("expected wrapped cleanup error, got %v", err)
	}
}


