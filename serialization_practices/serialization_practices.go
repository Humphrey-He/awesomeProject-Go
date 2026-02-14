package serialization_practices

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// UserDTO shows practical JSON tags:
// - stable field names
// - omitempty for optional pointer field
// - RFC3339 timestamps
type UserDTO struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	NickName  *string    `json:"nick_name,omitempty"`
	CreatedAt CustomTime `json:"created_at"`
}

// CustomTime demonstrates explicit time format control for APIs.
type CustomTime struct {
	time.Time
}

func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.Time.UTC().Format(time.RFC3339))
}

func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("unmarshal custom time as string: %w", err)
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fmt.Errorf("parse custom time: %w", err)
	}
	ct.Time = t
	return nil
}

// SerializeJSON writes pretty or compact JSON depending on indent flag.
func SerializeJSON(v any, indent bool) ([]byte, error) {
	if v == nil {
		return nil, errors.New("input is nil")
	}
	if indent {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshal indent json: %w", err)
		}
		return b, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}
	return b, nil
}

// DeserializeJSONStrict rejects unknown fields to avoid silent contract drift.
func DeserializeJSONStrict(data []byte, out any) error {
	if len(bytes.TrimSpace(data)) == 0 {
		return errors.New("empty payload")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode strict json: %w", err)
	}
	if dec.More() {
		return errors.New("multiple json values in payload")
	}
	return nil
}

// StreamEncodeJSONL serializes many values into JSON Lines.
func StreamEncodeJSONL(w io.Writer, values []UserDTO) error {
	enc := json.NewEncoder(w)
	for i := range values {
		if err := enc.Encode(values[i]); err != nil {
			return fmt.Errorf("encode json line idx=%d: %w", i, err)
		}
	}
	return nil
}

// StreamDecodeJSONL deserializes JSON Lines from reader.
func StreamDecodeJSONL(r io.Reader) ([]UserDTO, error) {
	dec := json.NewDecoder(r)
	var out []UserDTO
	for {
		var item UserDTO
		err := dec.Decode(&item)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode json line: %w", err)
		}
		out = append(out, item)
	}
	return out, nil
}

// ToGob / FromGob demonstrate binary serialization for internal transport.
func ToGob(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("encode gob: %w", err)
	}
	return buf.Bytes(), nil
}

func FromGob(data []byte, out any) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode gob: %w", err)
	}
	return nil
}

// RedactEmail masks local-part for safe logs before serialization.
func RedactEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 || len(parts[0]) <= 1 {
		return "***"
	}
	return parts[0][:1] + "***@" + parts[1]
}


