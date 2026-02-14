package ring_buffer

import "testing"

func TestDemoOutput(t *testing.T) {
	rb := NewRingBuffer(3)
	_ = rb.Write(1)
	_ = rb.Write(2)
	_ = rb.Write(3)
	t.Logf("full=%v len=%d slice=%v", rb.IsFull(), rb.Len(), rb.ToSlice())

	v, _ := rb.Read()
	t.Logf("read=%v len=%d", v, rb.Len())
	rb.WriteOverwrite(4)
	t.Logf("after overwrite slice=%v", rb.ToSlice())
}
