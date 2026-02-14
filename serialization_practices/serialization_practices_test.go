package serialization_practices

import (
	"bytes"
	"testing"
	"time"
)

func TestSerializeAndDeserializeStrict(t *testing.T) {
	now := CustomTime{Time: time.Date(2026, 2, 11, 12, 0, 0, 0, time.UTC)}
	u := UserDTO{
		ID:        1,
		Name:      "alice",
		Email:     "alice@example.com",
		CreatedAt: now,
	}

	b, err := SerializeJSON(u, false)
	if err != nil {
		t.Fatalf("serialize json: %v", err)
	}

	var out UserDTO
	if err := DeserializeJSONStrict(b, &out); err != nil {
		t.Fatalf("strict deserialize: %v", err)
	}
	if out.ID != 1 || out.Name != "alice" {
		t.Fatalf("unexpected out: %+v", out)
	}
}

func TestDeserializeStrictRejectUnknownField(t *testing.T) {
	payload := []byte(`{"id":1,"name":"bob","email":"b@x.com","created_at":"2026-02-11T12:00:00Z","extra":1}`)
	var out UserDTO
	if err := DeserializeJSONStrict(payload, &out); err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestJSONLOK(t *testing.T) {
	items := []UserDTO{
		{ID: 1, Name: "u1", Email: "u1@example.com", CreatedAt: CustomTime{Time: time.Now().UTC()}},
		{ID: 2, Name: "u2", Email: "u2@example.com", CreatedAt: CustomTime{Time: time.Now().UTC()}},
	}
	var buf bytes.Buffer
	if err := StreamEncodeJSONL(&buf, items); err != nil {
		t.Fatalf("encode jsonl: %v", err)
	}
	got, err := StreamDecodeJSONL(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("decode jsonl: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len=%d want=2", len(got))
	}
}

func TestGobRoundTrip(t *testing.T) {
	in := map[string]int{"a": 1, "b": 2}
	b, err := ToGob(in)
	if err != nil {
		t.Fatalf("to gob: %v", err)
	}
	var out map[string]int
	if err := FromGob(b, &out); err != nil {
		t.Fatalf("from gob: %v", err)
	}
	if out["a"] != 1 || out["b"] != 2 {
		t.Fatalf("unexpected out: %#v", out)
	}
}

func TestRedactEmail(t *testing.T) {
	got := RedactEmail("alice@example.com")
	if got != "a***@example.com" {
		t.Fatalf("redacted=%s", got)
	}
}


