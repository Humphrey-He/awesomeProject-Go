package json_zero_ambiguity

import "testing"

func TestOptionalStates(t *testing.T) {
	req, err := DecodePatchStrict([]byte(`{"name":"alice","age":0,"bio":null}`))
	if err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if !req.Name.Set || req.Name.Null || req.Name.Value != "alice" {
		t.Fatalf("name state invalid: %+v", req.Name)
	}
	if !req.Age.Set || req.Age.Null || req.Age.Value != 0 {
		t.Fatalf("age state invalid: %+v", req.Age)
	}
	if !req.Bio.Set || !req.Bio.Null {
		t.Fatalf("bio state invalid: %+v", req.Bio)
	}
	if req.Active.Set {
		t.Fatalf("active should be absent")
	}
}

func TestApplyPatch(t *testing.T) {
	u := &User{Name: "old", Age: 18, Active: true, Bio: "hi"}
	req, _ := DecodePatchStrict([]byte(`{"name":"new","bio":null}`))
	if err := ApplyPatch(u, req); err != nil {
		t.Fatalf("apply patch err: %v", err)
	}
	if u.Name != "new" || u.Bio != "" || u.Age != 18 || !u.Active {
		t.Fatalf("unexpected user: %+v", *u)
	}
}

func TestDecodePatchStrictUnknownField(t *testing.T) {
	_, err := DecodePatchStrict([]byte(`{"x":1}`))
	if err == nil {
		t.Fatal("want unknown field error")
	}
}


