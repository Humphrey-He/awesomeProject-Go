package unsafe_core

import "testing"

func TestDemoOutput(t *testing.T) {
	info := StructLayoutDemo()
	t.Logf("layout size=%d align=%d offsetA=%d offsetB=%d offsetC=%d",
		info.Size, info.Align, info.OffsetA, info.OffsetB, info.OffsetC)

	b := []byte("abc")
	s := BytesToStringNoCopy(b)
	t.Logf("before b=%q s=%q", b, s)
	b[0] = 'z'
	t.Logf("after  b=%q s=%q", b, s)

	t.Logf("little_endian=%v", IsLittleEndian())
}
