package polymorphism_inheritance

import "testing"

func TestDemoOutput(t *testing.T) {
	d := NewDog("buddy", "golden")
	b := NewBird("kiwi", true)
	t.Logf("dog: name=%s speak=%s move=%s", d.Name(), d.Speak(), d.Move())
	t.Logf("bird: name=%s speak=%s move=%s", b.Name(), b.Speak(), b.Move())

	var z Zoo
	z.Add(d)
	z.Add(b)
	for _, line := range z.Perform() {
		t.Logf("zoo: %s", line)
	}
}


