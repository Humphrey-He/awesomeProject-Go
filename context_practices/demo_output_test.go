package context_practices

import (
	"context"
	"testing"
	"time"
)

func TestDemoOutput(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req-demo-001")
	if id, ok := RequestIDFrom(ctx); ok {
		t.Logf("request_id=%s", id)
	}

	err := ProcessWithTimeout(ctx, 20*time.Millisecond, 50*time.Millisecond)
	t.Logf("ProcessWithTimeout err=%v", err)

	err = RunPipeline(context.Background())
	t.Logf("RunPipeline err=%v", err)
}
