package ddd_practices

import "testing"

func TestDemoOutput(t *testing.T) {
	o := NewOrder("demo")
	_ = o.AddItem(OrderItem{SKU: "sku", Qty: 1, Price: Money{Currency: "USD", Amount: 50}})
	_ = o.Pay()
	total, _ := o.Total()
	t.Logf("status=%s total=%d %s events=%d", o.Status, total.Amount, total.Currency, len(o.Events()))
}
