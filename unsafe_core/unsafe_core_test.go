package unsafe_core

import (
	"testing"
)

func TestStructLayoutDemo(t *testing.T) {
	info := StructLayoutDemo()
	if info.Size == 0 || info.Align == 0 {
		t.Fatalf("invalid layout info: %+v", info)
	}
	if !(info.OffsetA <= info.OffsetB && info.OffsetB <= info.OffsetC) {
		t.Fatalf("unexpected offsets: %+v", info)
	}
}

func TestBytesToStringNoCopy(t *testing.T) {
	b := []byte("abc")
	s := BytesToStringNoCopy(b)
	b[0] = 'z'
	if s != "zbc" {
		t.Fatalf("string should share same backing bytes, got %s", s)
	}
}

func TestStringToBytesNoCopy(t *testing.T) {
	s := "hello"
	b := StringToBytesNoCopy(s)
	if string(b) != "hello" {
		t.Fatalf("unexpected bytes view: %s", string(b))
	}
}

func TestUnsafeFloatBits(t *testing.T) {
	if !EqualFloatBits(3.1415926) {
		t.Fatal("unsafe float bits mismatch")
	}
}

func TestBytesToUint32NativeUnsafe(t *testing.T) {
	v, err := BytesToUint32NativeUnsafe([]byte{1, 0, 0, 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if IsLittleEndian() && v != 1 {
		t.Fatalf("little-endian expected 1, got %d", v)
	}
	if !IsLittleEndian() && v != 16777216 {
		t.Fatalf("big-endian expected 16777216, got %d", v)
	}
}
