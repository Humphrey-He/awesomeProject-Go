package context_practices

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type requestIDKey struct{}

// WithRequestID uses typed key to avoid key collisions.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func RequestIDFrom(ctx context.Context) (string, bool) {
	v := ctx.Value(requestIDKey{})
	s, ok := v.(string)
	return s, ok
}

// DoWork simulates downstream call and always takes context as first arg.
func DoWork(ctx context.Context, cost time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(cost):
		return nil
	}
}

// ProcessWithTimeout demonstrates explicit timeout boundary.
func ProcessWithTimeout(parent context.Context, timeout time.Duration, cost time.Duration) error {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	return DoWork(ctx, cost)
}

// PipelineStep shows context propagation style in call chain.
func PipelineStep(ctx context.Context, stepName string, cost time.Duration) error {
	if stepName == "" {
		return errors.New("step name is empty")
	}
	if err := DoWork(ctx, cost); err != nil {
		return fmt.Errorf("step=%s: %w", stepName, err)
	}
	return nil
}

// RunPipeline demonstrates structured propagation and fast-fail cancel.
func RunPipeline(parent context.Context) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	if err := PipelineStep(ctx, "fetch", 20*time.Millisecond); err != nil {
		return err
	}
	if err := PipelineStep(ctx, "transform", 20*time.Millisecond); err != nil {
		return err
	}
	if err := PipelineStep(ctx, "save", 20*time.Millisecond); err != nil {
		return err
	}
	return nil
}


