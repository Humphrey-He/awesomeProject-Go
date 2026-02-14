package result_wrapper

import (
	"errors"
	"testing"
)

func TestDemoOutput(t *testing.T) {
	step1 := Success(2)
	step2 := Map(step1, func(v int) int { return v * 10 })
	step3 := Bind(step2, func(v int) Result[float64] {
		if v == 0 {
			return Failure[float64](errors.New("divide by zero"))
		}
		return Success(100.0 / float64(v))
	})

	v, err := step3.Unpack()
	t.Logf("step1=%v step2=%v", step1.ValueOrZero(), step2.ValueOrZero())
	t.Logf("final value=%.2f err=%v", v, err)

	failed := WrapError(Failure[int](errors.New("db down")), "query user")
	t.Logf("wrapped error=%v", failed.Err())
}
