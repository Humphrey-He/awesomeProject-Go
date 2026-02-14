//go:build !cgo
// +build !cgo

package cgo_practice

import "testing"

func TestCGORequired(t *testing.T) {
	t.Skip("cgo_practice requires CGO_ENABLED=1 and a working C toolchain")
}
