package interface_internals

import "testing"

func TestDescribeInterface(t *testing.T) {
	if got := DescribeInterface(nil); got != "interface=nil" {
		t.Fatalf("unexpected describe nil: %s", got)
	}
	got := DescribeInterface(Person{Name: "alice"})
	if got == "" {
		t.Fatal("describe should not be empty")
	}
}

func TestNilPitfall(t *testing.T) {
	isNil, desc := NilPitfall()
	if isNil {
		t.Fatal("typed nil in interface should not be nil interface")
	}
	if desc == "" {
		t.Fatal("description should not be empty")
	}
}

func TestTypeAssertDemo(t *testing.T) {
	if out, ok := TypeAssertDemo(Person{Name: "bob"}); !ok || out != "person:bob" {
		t.Fatalf("unexpected person assert result: %s %v", out, ok)
	}
	if out, ok := TypeAssertDemo("x"); !ok || out != "string:x" {
		t.Fatalf("unexpected string assert result: %s %v", out, ok)
	}
	if _, ok := TypeAssertDemo(3.14); ok {
		t.Fatal("float should be unknown=false")
	}
}

func TestBuildModels(t *testing.T) {
	em := BuildEmptyModel(123)
	if em.Type == "" || em.Data == "" {
		t.Fatal("empty interface model fields should not be empty")
	}

	nm := BuildNonEmptyModel(Person{Name: "mike"})
	if nm.Interface != "Speaker" {
		t.Fatalf("unexpected interface name: %s", nm.Interface)
	}
	if len(nm.MethodSet) != 1 {
		t.Fatalf("unexpected method set: %v", nm.MethodSet)
	}
}
