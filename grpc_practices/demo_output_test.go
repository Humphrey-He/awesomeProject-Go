package grpc_practices

import (
	"bytes"
	"testing"
)

func TestDemoOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	_ = WriteMessage(buf, false, []byte("hello"))
	compressed, payload, _ := ReadMessage(bytes.NewReader(buf.Bytes()), 1024)
	t.Logf("compressed=%v payload=%s", compressed, string(payload))
}
