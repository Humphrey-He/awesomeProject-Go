package event_sourcing_mvp

import "testing"

func TestDemoOutput(t *testing.T) {
	store := NewInMemoryEventStore()
	acc := &Account{ID: "demo"}
	e1, _ := acc.Deposit(20)
	acc.Apply(e1)
	_ = store.Append(acc.ID, 0, e1)
	events, _ := store.Load(acc.ID)
	rebuilt := RehydrateAccount(acc.ID, events)
	t.Logf("balance=%d version=%d", rebuilt.Balance, rebuilt.Version)
}
