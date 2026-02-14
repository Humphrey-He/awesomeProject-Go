package bit_operations

import "testing"

func TestDemoOutput(t *testing.T) {
	var n uint64 = 0
	n = SetBit(n, 1)
	n = SetBit(n, 3)
	t.Logf("after set bit1,bit3 => %08b", n)

	n = ToggleBit(n, 1)
	t.Logf("after toggle bit1 => %08b", n)
	t.Logf("has bit3=%v countBits=%d", HasBit(n, 3), CountBits(n))
	t.Logf("next power of two for 13 => %d", NextPowerOfTwo(13))
}
