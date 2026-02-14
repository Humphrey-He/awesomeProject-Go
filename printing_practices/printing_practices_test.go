package printing_practices

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestFormatKV(t *testing.T) {
	out := FormatKV(map[string]any{
		"b": 2,
		"a": 1,
	})
	if out != "a=1 b=2" {
		t.Fatalf("unexpected kv: %s", out)
	}
}

func TestPrintError(t *testing.T) {
	err := errors.New("timeout")
	out := PrintError("query_db", err)
	if !strings.Contains(out, "op=query_db") || !strings.Contains(out, "timeout") {
		t.Fatalf("unexpected error print: %s", out)
	}
}

func TestNewLogger(t *testing.T) {
	l, buf := NewLogger("[svc] ")
	l.Println("hello")
	if !strings.Contains(buf.String(), "hello") {
		t.Fatalf("logger output missing message: %s", buf.String())
	}
}

func TestTimeFormatting(t *testing.T) {
	base := time.Date(2026, 2, 11, 14, 5, 6, 123456789, time.FixedZone("UTC+8", 8*3600))
	utc := FormatTimeUTC(base)
	if !strings.HasSuffix(utc, "Z") {
		t.Fatalf("utc format should end with Z, got %s", utc)
	}
	local := FormatTimeInLocation(base, time.FixedZone("CST", 8*3600))
	if !strings.Contains(local, "CST") {
		t.Fatalf("location format missing zone: %s", local)
	}
}

func TestHumanDuration(t *testing.T) {
	if got := HumanDuration(500 * time.Microsecond); got != "500us" {
		t.Fatalf("unexpected micro duration: %s", got)
	}
	if got := HumanDuration(12 * time.Millisecond); !strings.Contains(got, "ms") {
		t.Fatalf("unexpected milli duration: %s", got)
	}
	if got := HumanDuration(2 * time.Second); got != "2.00s" {
		t.Fatalf("unexpected second duration: %s", got)
	}
}


