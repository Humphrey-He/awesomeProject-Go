package json_zero_ambiguity

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// Optional distinguishes 3 states for JSON field:
// 1) absent:   Set=false
// 2) null:     Set=true, Null=true
// 3) value:    Set=true, Null=false, Value=...
type Optional[T any] struct {
	Set   bool
	Null  bool
	Value T
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		o.Null = true
		var zero T
		o.Value = zero
		return nil
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("unmarshal optional value: %w", err)
	}
	o.Null = false
	o.Value = v
	return nil
}

func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if o.Null {
		return []byte("null"), nil
	}
	return json.Marshal(o.Value)
}

type User struct {
	Name   string
	Age    int
	Active bool
	Bio    string
}

type PatchUserRequest struct {
	Name   Optional[string] `json:"name"`
	Age    Optional[int]    `json:"age"`
	Active Optional[bool]   `json:"active"`
	Bio    Optional[string] `json:"bio"`
}

// ApplyPatch applies only "Set=true" fields.
// For string fields, null clears to empty string.
func ApplyPatch(u *User, req PatchUserRequest) error {
	if u == nil {
		return errors.New("user is nil")
	}

	if req.Name.Set {
		if req.Name.Null {
			u.Name = ""
		} else {
			u.Name = req.Name.Value
		}
	}
	if req.Age.Set {
		if req.Age.Null {
			u.Age = 0
		} else {
			u.Age = req.Age.Value
		}
	}
	if req.Active.Set {
		if req.Active.Null {
			u.Active = false
		} else {
			u.Active = req.Active.Value
		}
	}
	if req.Bio.Set {
		if req.Bio.Null {
			u.Bio = ""
		} else {
			u.Bio = req.Bio.Value
		}
	}
	return nil
}

func DecodePatchStrict(data []byte) (PatchUserRequest, error) {
	var req PatchUserRequest
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return PatchUserRequest{}, fmt.Errorf("decode patch strict: %w", err)
	}
	return req, nil
}


