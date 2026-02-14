package interface_internals

import "testing"

func TestDemoOutput(t *testing.T) {
	t.Logf("describe nil => %s", DescribeInterface(nil))
	t.Logf("describe person => %s", DescribeInterface(Person{Name: "alice"}))

	isNil, desc := NilPitfall()
	t.Logf("nil pitfall isNil=%v desc=%s", isNil, desc)

	out, ok := TypeAssertDemo("hello")
	t.Logf("type assert string => out=%s ok=%v", out, ok)
}


