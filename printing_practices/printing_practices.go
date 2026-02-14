package printing_practices

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

// FormatKV prints stable key=value pairs sorted by key for deterministic logs.
func FormatKV(fields map[string]any) string {
	if len(fields) == 0 {
		return ""
	}
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, fields[k]))
	}
	return strings.Join(parts, " ")
}

// PrintError wraps operation and error for clear diagnostics.
func PrintError(op string, err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("op=%s err=%v", op, err)
}

// NewLogger creates a logger with microseconds and file line for troubleshooting.
func NewLogger(prefix string) (*log.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	l := log.New(&buf, prefix, log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	return l, &buf
}

// Time printing best practices.
func FormatTimeUTC(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func FormatTimeInLocation(t time.Time, loc *time.Location) string {
	if loc == nil {
		loc = time.Local
	}
	return t.In(loc).Format("2006-01-02 15:04:05 MST")
}

// HumanDuration prints duration with friendly unit selection.
func HumanDuration(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%dus", d.Microseconds())
	case d < time.Second:
		return fmt.Sprintf("%.2fms", float64(d.Microseconds())/1000.0)
	case d < time.Minute:
		return fmt.Sprintf("%.2fs", d.Seconds())
	default:
		return d.Round(time.Second).String()
	}
}

// ExampleBusinessLog composes a practical one-line log.
func ExampleBusinessLog(traceID string, cost time.Duration, status int) string {
	return FormatKV(map[string]any{
		"trace_id": traceID,
		"status":   status,
		"cost":     HumanDuration(cost),
		"time":     FormatTimeUTC(time.Now()),
	})
}


