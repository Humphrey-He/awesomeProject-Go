package error_handling_practices

import "testing"

func TestDemoOutput(t *testing.T) {
	id, err := ParseUserID("12")
	t.Logf("ParseUserID(12) => id=%d err=%v", id, err)

	_, err = ParseUserID("")
	t.Logf("ParseUserID(\"\") => err=%v", err)

	name, err := FindUserName(map[int]string{1: "alice"}, 1)
	t.Logf("FindUserName(1) => name=%s err=%v", name, err)

	_, err = FindUserName(map[int]string{1: "alice"}, 2)
	t.Logf("FindUserName(2) => err=%v", err)
}
