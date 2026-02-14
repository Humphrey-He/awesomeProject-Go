package defer_practices

import "testing"

// TestDemoOutput is intentionally used to show stdout in `go test -v`.
func TestDemoOutput(t *testing.T) {
	t.Log("=== DeferExecutionOrder ===")
	DeferExecutionOrder()

	t.Log("=== DeferArgumentEvaluation ===")
	DeferArgumentEvaluation()

	t.Log("=== DeferWithClosure ===")
	DeferWithClosure()

	t.Log("=== DeferWithPointer ===")
	DeferWithPointer()

	t.Logf("DeferReturnValue1=%d", DeferReturnValue1())
	t.Logf("DeferReturnValue2=%d", DeferReturnValue2())
}
