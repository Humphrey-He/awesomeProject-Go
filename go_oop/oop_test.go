package go_oop

import (
	"errors"
	"testing"
)

type memoryStore struct {
	data map[int64]Entity
}

func newMemoryStore() *memoryStore {
	return &memoryStore{data: make(map[int64]Entity)}
}

func (m *memoryStore) Save(e Entity) error {
	m.data[e.GetID()] = e
	return nil
}

func (m *memoryStore) FindByID(id int64) (Entity, error) {
	e, ok := m.data[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return e, nil
}

func TestUser_Encapsulation(t *testing.T) {
	u, err := NewUser(1, "alice", "a@example.com")
	if err != nil {
		t.Fatalf("new user error: %v", err)
	}
	if err := u.Deposit(100); err != nil {
		t.Fatalf("deposit error: %v", err)
	}
	if err := u.Withdraw(30); err != nil {
		t.Fatalf("withdraw error: %v", err)
	}
	if got := u.Balance(); got != 70 {
		t.Fatalf("balance=%d want=70", got)
	}
}

func TestCustomer_CompositionAndOverride(t *testing.T) {
	c, err := NewCustomer(2, "bob", "b@example.com", "vip")
	if err != nil {
		t.Fatalf("new customer error: %v", err)
	}
	c.Profile = Profile{Phone: "1880000", Address: "shanghai"}

	if got := c.ContactInfo(); got == "" {
		t.Fatal("contact info should not be empty")
	}
	if got := c.DisplayName(); got != "bob[vip]" {
		t.Fatalf("display name=%s", got)
	}
}

func TestService_DependencyInversion(t *testing.T) {
	svc := NewService(newMemoryStore())
	u, _ := NewUser(3, "charlie", "c@example.com")
	if err := svc.Register(u); err != nil {
		t.Fatalf("register error: %v", err)
	}
	name, err := svc.GetDisplayName(3)
	if err != nil {
		t.Fatalf("get name error: %v", err)
	}
	if name != "charlie" {
		t.Fatalf("name=%s want=charlie", name)
	}
}


