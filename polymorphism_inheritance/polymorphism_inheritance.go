package polymorphism_inheritance

import "fmt"

// Animal is the polymorphic interface.
type Animal interface {
	Name() string
	Speak() string
	Move() string
}

// BaseAnimal acts as reusable shared behavior.
type BaseAnimal struct {
	AnimalName string
	Legs       int
}

func (b BaseAnimal) Name() string {
	return b.AnimalName
}

func (b BaseAnimal) Move() string {
	return fmt.Sprintf("%s moves with %d legs", b.AnimalName, b.Legs)
}

// Dog "inherits" BaseAnimal by embedding and overrides Speak.
type Dog struct {
	BaseAnimal
	Breed string
}

func NewDog(name, breed string) Dog {
	return Dog{
		BaseAnimal: BaseAnimal{AnimalName: name, Legs: 4},
		Breed:      breed,
	}
}

func (d Dog) Speak() string {
	return "woof"
}

// Bird also embeds BaseAnimal and overrides Move/Speak.
type Bird struct {
	BaseAnimal
	CanFly bool
}

func NewBird(name string, canFly bool) Bird {
	return Bird{
		BaseAnimal: BaseAnimal{AnimalName: name, Legs: 2},
		CanFly:     canFly,
	}
}

func (b Bird) Speak() string {
	return "chirp"
}

func (b Bird) Move() string {
	if b.CanFly {
		return fmt.Sprintf("%s flies", b.AnimalName)
	}
	return b.BaseAnimal.Move()
}

// Zoo shows runtime polymorphism.
type Zoo struct {
	animals []Animal
}

func (z *Zoo) Add(a Animal) {
	z.animals = append(z.animals, a)
}

func (z *Zoo) Perform() []string {
	out := make([]string, 0, len(z.animals))
	for _, a := range z.animals {
		out = append(out, fmt.Sprintf("%s says %s", a.Name(), a.Speak()))
	}
	return out
}


