package websocket_practices

import (
	"bytes"
	"testing"
)

func TestDemoOutput(t *testing.T) {
	f := Frame{Fin: true, Opcode: 0x1, Payload: []byte("hi")}
	encoded, err := EncodeFrame(f)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := DecodeFrame(bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	accept := ComputeAcceptKey("demo")
	t.Logf("accept=%s opcode=%d payload=%s", accept, decoded.Opcode, string(decoded.Payload))
}
