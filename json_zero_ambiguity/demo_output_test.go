package json_zero_ambiguity

import "testing"

func TestDemoOutput(t *testing.T) {
	u := &User{Name: "old", Age: 20, Active: true, Bio: "hello"}
	req, err := DecodePatchStrict([]byte(`{"name":"new","age":0,"bio":null}`))
	if err != nil {
		t.Fatalf("decode err: %v", err)
	}
	t.Logf("before user=%+v", *u)
	t.Logf("name set=%v null=%v val=%q", req.Name.Set, req.Name.Null, req.Name.Value)
	t.Logf("age set=%v null=%v val=%d", req.Age.Set, req.Age.Null, req.Age.Value)
	t.Logf("bio set=%v null=%v", req.Bio.Set, req.Bio.Null)

	if err := ApplyPatch(u, req); err != nil {
		t.Fatalf("apply err: %v", err)
	}
	t.Logf("after user=%+v", *u)
}
