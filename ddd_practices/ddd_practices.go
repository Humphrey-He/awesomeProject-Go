package ddd_practices

import (
	"errors"
)

type Money struct {
	Currency string
	Amount   int64
}

func (m Money) Add(o Money) (Money, error) {
	if m.Currency != o.Currency {
		return Money{}, errors.New("currency mismatch")
	}
	return Money{Currency: m.Currency, Amount: m.Amount + o.Amount}, nil
}

func (m Money) Mul(qty int) Money {
	return Money{Currency: m.Currency, Amount: m.Amount * int64(qty)}
}

type OrderStatus string

const (
	OrderDraft  OrderStatus = "DRAFT"
	OrderPaid   OrderStatus = "PAID"
	OrderCancel OrderStatus = "CANCELLED"
)

type OrderItem struct {
	SKU   string
	Qty   int
	Price Money
}

type DomainEvent interface {
	Name() string
}

type OrderCreated struct{ ID string }

func (e OrderCreated) Name() string { return "OrderCreated" }

type OrderPaid struct{ ID string }

func (e OrderPaid) Name() string { return "OrderPaid" }

type Order struct {
	ID     string
	Status OrderStatus
	Items  []OrderItem
	events []DomainEvent
}

func NewOrder(id string) *Order {
	o := &Order{ID: id, Status: OrderDraft}
	o.events = append(o.events, OrderCreated{ID: id})
	return o
}

func (o *Order) AddItem(item OrderItem) error {
	if o.Status != OrderDraft {
		return errors.New("order not editable")
	}
	if item.Qty <= 0 {
		return errors.New("qty must be positive")
	}
	if item.Price.Amount <= 0 {
		return errors.New("price must be positive")
	}
	if len(o.Items) > 0 && o.Items[0].Price.Currency != item.Price.Currency {
		return errors.New("currency mismatch")
	}
	o.Items = append(o.Items, item)
	return nil
}

func (o *Order) Total() (Money, error) {
	if len(o.Items) == 0 {
		return Money{}, errors.New("empty order")
	}
	total := Money{Currency: o.Items[0].Price.Currency}
	for _, item := range o.Items {
		line := item.Price.Mul(item.Qty)
		sum, err := total.Add(line)
		if err != nil {
			return Money{}, err
		}
		total = sum
	}
	return total, nil
}

func (o *Order) Pay() error {
	if o.Status != OrderDraft {
		return errors.New("invalid state")
	}
	if _, err := o.Total(); err != nil {
		return err
	}
	o.Status = OrderPaid
	o.events = append(o.events, OrderPaid{ID: o.ID})
	return nil
}

func (o *Order) Events() []DomainEvent {
	return append([]DomainEvent(nil), o.events...)
}

type OrderRepository interface {
	Save(order *Order) error
	FindByID(id string) (*Order, error)
}

type InMemoryOrderRepository struct {
	data map[string]*Order
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{data: map[string]*Order{}}
}

func (r *InMemoryOrderRepository) Save(order *Order) error {
	r.data[order.ID] = order
	return nil
}

func (r *InMemoryOrderRepository) FindByID(id string) (*Order, error) {
	if v, ok := r.data[id]; ok {
		return v, nil
	}
	return nil, errors.New("not found")
}
