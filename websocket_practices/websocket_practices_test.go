package websocket_practices

import (
	"bytes"
	"testing"
)

func TestComputeAcceptKey(t *testing.T) {
	key := "dGhlIHNhbXBsZSBub25jZQ=="
	want := "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="
	if got := ComputeAcceptKey(key); got != want {
		t.Fatalf("accept mismatch: %s", got)
	}
}

func TestFrameRoundTrip(t *testing.T) {
	f := Frame{
		Fin:     true,
		Opcode:  0x1,
		Masked:  true,
		MaskKey: [4]byte{1, 2, 3, 4},
		Payload: []byte("hello"),
	}
	encoded, err := EncodeFrame(f)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := DecodeFrame(bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if string(decoded.Payload) != "hello" {
		t.Fatalf("payload mismatch: %s", string(decoded.Payload))
	}
}
