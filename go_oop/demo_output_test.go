package go_oop

import (
	"errors"
	"testing"
)

type demoStore struct {
	data map[int64]Entity
}

func (d *demoStore) Save(e Entity) error {
	d.data[e.GetID()] = e
	return nil
}

func (d *demoStore) FindByID(id int64) (Entity, error) {
	e, ok := d.data[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return e, nil
}

func TestDemoOutput(t *testing.T) {
	c, _ := NewCustomer(1, "alice", "a@example.com", "vip")
	c.Profile = Profile{Phone: "1880000", Address: "beijing"}
	_ = c.Deposit(100)
	_ = c.Withdraw(30)
	t.Logf("customer display=%s contact=%s balance=%d", c.DisplayName(), c.ContactInfo(), c.Balance())

	svc := NewService(&demoStore{data: map[int64]Entity{}})
	_ = svc.Register(c)
	name, _ := svc.GetDisplayName(1)
	t.Logf("service display name=%s", name)
}


