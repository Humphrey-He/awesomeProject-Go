package bit_operations

import "testing"

func TestBasicBitOps(t *testing.T) {
	var n uint64
	n = SetBit(n, 1)
	if !HasBit(n, 1) {
		t.Fatal("bit1 should be set")
	}
	n = ToggleBit(n, 1)
	if HasBit(n, 1) {
		t.Fatal("bit1 should be toggled off")
	}
	n = SetBit(n, 3)
	n = ClearBit(n, 3)
	if HasBit(n, 3) {
		t.Fatal("bit3 should be cleared")
	}
}

func TestCountAndPowerOfTwo(t *testing.T) {
	if c := CountBits(0b101101); c != 4 {
		t.Fatalf("count=%d want=4", c)
	}
	if !IsPowerOfTwo(1024) {
		t.Fatal("1024 should be power of two")
	}
	if IsPowerOfTwo(1023) {
		t.Fatal("1023 should not be power of two")
	}
}

func TestNextPowerOfTwo(t *testing.T) {
	cases := map[uint64]uint64{
		1:    1,
		2:    2,
		3:    4,
		5:    8,
		1025: 2048,
	}
	for in, want := range cases {
		if got := NextPowerOfTwo(in); got != want {
			t.Fatalf("NextPowerOfTwo(%d)=%d want=%d", in, got, want)
		}
	}
}

func TestPermMask(t *testing.T) {
	var m uint8
	m = GrantPerm(m, PermRead)
	m = GrantPerm(m, PermWrite)
	if !HasPerm(m, PermRead) || !HasPerm(m, PermWrite) {
		t.Fatalf("permission missing: %08b", m)
	}
	m = RevokePerm(m, PermWrite)
	if HasPerm(m, PermWrite) {
		t.Fatalf("write should be revoked: %08b", m)
	}
}
