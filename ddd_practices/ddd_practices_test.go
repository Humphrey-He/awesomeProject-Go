package ddd_practices

import "testing"

func TestOrderLifecycle(t *testing.T) {
	o := NewOrder("o-1")
	if err := o.AddItem(OrderItem{SKU: "sku-1", Qty: 2, Price: Money{Currency: "USD", Amount: 100}}); err != nil {
		t.Fatalf("add item: %v", err)
	}
	if err := o.Pay(); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if o.Status != OrderPaid {
		t.Fatalf("status mismatch")
	}
	if len(o.Events()) != 2 {
		t.Fatalf("events mismatch")
	}
}

func TestOrderRepository(t *testing.T) {
	repo := NewInMemoryOrderRepository()
	o := NewOrder("o-2")
	_ = repo.Save(o)
	if _, err := repo.FindByID("o-2"); err != nil {
		t.Fatalf("expected order")
	}
}
