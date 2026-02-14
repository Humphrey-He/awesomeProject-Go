package context_practices

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRequestID(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req-1")
	id, ok := RequestIDFrom(ctx)
	if !ok || id != "req-1" {
		t.Fatalf("id=%s ok=%v", id, ok)
	}
}

func TestProcessWithTimeout(t *testing.T) {
	err := ProcessWithTimeout(context.Background(), 20*time.Millisecond, 80*time.Millisecond)
	if err == nil {
		t.Fatal("want timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestRunPipeline(t *testing.T) {
	if err := RunPipeline(context.Background()); err != nil {
		t.Fatalf("pipeline err: %v", err)
	}
}


