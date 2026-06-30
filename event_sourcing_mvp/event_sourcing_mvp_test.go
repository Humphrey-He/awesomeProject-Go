package event_sourcing_mvp

import "testing"

func TestEventSourcingFlow(t *testing.T) {
	store := NewInMemoryEventStore()
	account := &Account{ID: "acc-1"}
	e1, _ := account.Deposit(100)
	account.Apply(e1)
	if err := store.Append(account.ID, 0, e1); err != nil {
		t.Fatalf("append: %v", err)
	}
	e2, _ := account.Withdraw(40)
	account.Apply(e2)
	if err := store.Append(account.ID, 1, e2); err != nil {
		t.Fatalf("append: %v", err)
	}
	if account.Balance != 60 {
		t.Fatalf("balance mismatch")
	}

	events, _ := store.Load(account.ID)
	loaded := RehydrateAccount(account.ID, events)
	if loaded.Balance != 60 || loaded.Version != 2 {
		t.Fatalf("rehydrate mismatch")
	}
}

func TestVersionConflict(t *testing.T) {
	store := NewInMemoryEventStore()
	if err := store.Append("x", 1, Deposited{Amount: 10, Ver: 1}); err == nil {
		t.Fatalf("expected conflict")
	}
}
