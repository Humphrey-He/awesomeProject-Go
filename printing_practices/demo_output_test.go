package printing_practices

import (
	"errors"
	"testing"
	"time"
)

func TestDemoOutput(t *testing.T) {
	kv := FormatKV(map[string]any{"trace_id": "t-1", "status": 200, "cost": "12ms"})
	t.Logf("kv log: %s", kv)
	t.Logf("error log: %s", PrintError("query_db", errors.New("timeout")))

	now := time.Date(2026, 2, 12, 18, 30, 1, 123000000, time.FixedZone("CST", 8*3600))
	t.Logf("utc time: %s", FormatTimeUTC(now))
	t.Logf("local time: %s", FormatTimeInLocation(now, time.FixedZone("CST", 8*3600)))
	t.Logf("duration: %s", HumanDuration(1530*time.Millisecond))
}
