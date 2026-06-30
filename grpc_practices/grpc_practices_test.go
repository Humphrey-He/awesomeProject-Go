package grpc_practices

import (
	"bytes"
	"context"
	"testing"
)

func TestFrameRoundTrip(t *testing.T) {
	buf := &bytes.Buffer{}
	if err := WriteMessage(buf, false, []byte("ping")); err != nil {
		t.Fatalf("write: %v", err)
	}
	compressed, payload, err := ReadMessage(bytes.NewReader(buf.Bytes()), 1024)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if compressed {
		t.Fatalf("unexpected compressed flag")
	}
	if string(payload) != "ping" {
		t.Fatalf("payload mismatch: %s", string(payload))
	}
}

func TestInterceptorOrder(t *testing.T) {
	order := make([]string, 0, 4)
	interceptorA := func(ctx context.Context, req []byte, next UnaryHandler) ([]byte, error) {
		order = append(order, "A-in")
		resp, err := next(ctx, req)
		order = append(order, "A-out")
		return resp, err
	}
	interceptorB := func(ctx context.Context, req []byte, next UnaryHandler) ([]byte, error) {
		order = append(order, "B-in")
		resp, err := next(ctx, req)
		order = append(order, "B-out")
		return resp, err
	}
	h := func(ctx context.Context, req []byte) ([]byte, error) {
		order = append(order, "H")
		return []byte("ok"), nil
	}
	wrapped := ChainUnaryInterceptors(h, interceptorA, interceptorB)
	_, _ = wrapped(context.Background(), []byte("req"))
	want := "A-in,B-in,H,B-out,A-out"
	got := stringsJoin(order)
	if got != want {
		t.Fatalf("order mismatch: %s", got)
	}
}

func stringsJoin(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ","
		}
		out += p
	}
	return out
}
