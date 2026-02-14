package serialization_practices

import (
	"bytes"
	"testing"
	"time"
)

func TestDemoOutput(t *testing.T) {
	u := UserDTO{
		ID:        1,
		Name:      "alice",
		Email:     "alice@example.com",
		CreatedAt: CustomTime{Time: time.Date(2026, 2, 12, 10, 0, 0, 0, time.UTC)},
	}

	b, err := SerializeJSON(u, true)
	if err != nil {
		t.Fatalf("serialize err: %v", err)
	}
	t.Logf("pretty json:\n%s", string(b))

	var out UserDTO
	if err := DeserializeJSONStrict(b, &out); err != nil {
		t.Fatalf("deserialize err: %v", err)
	}
	t.Logf("decoded user: %+v", out)

	var buf bytes.Buffer
	if err := StreamEncodeJSONL(&buf, []UserDTO{u, out}); err != nil {
		t.Fatalf("jsonl encode err: %v", err)
	}
	t.Logf("jsonl payload:\n%s", buf.String())
}
