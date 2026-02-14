package polymorphism_inheritance

import "testing"

func TestMethodOverrideByEmbedding(t *testing.T) {
	d := NewDog("buddy", "golden")
	if got := d.Speak(); got != "woof" {
		t.Fatalf("dog speak=%s", got)
	}
	if got := d.Move(); got != "buddy moves with 4 legs" {
		t.Fatalf("dog move=%s", got)
	}

	b := NewBird("kiwi", true)
	if got := b.Move(); got != "kiwi flies" {
		t.Fatalf("bird move=%s", got)
	}
}

func TestPolymorphism_Dispatch(t *testing.T) {
	var animals []Animal
	animals = append(animals, NewDog("d1", "mix"))
	animals = append(animals, NewBird("b1", false))

	got := []string{animals[0].Speak(), animals[1].Speak()}
	if got[0] != "woof" || got[1] != "chirp" {
		t.Fatalf("unexpected polymorphic results: %v", got)
	}
}

func TestZoo_Perform(t *testing.T) {
	var z Zoo
	z.Add(NewDog("max", "husky"))
	z.Add(NewBird("mimi", true))
	res := z.Perform()
	if len(res) != 2 {
		t.Fatalf("res length=%d", len(res))
	}
}


