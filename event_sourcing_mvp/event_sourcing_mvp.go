package event_sourcing_mvp

import (
	"errors"
)

type Event interface {
	Type() string
	Version() int
}

type Deposited struct {
	Amount int64
	Ver    int
}

func (e Deposited) Type() string { return "Deposited" }
func (e Deposited) Version() int { return e.Ver }

type Withdrawn struct {
	Amount int64
	Ver    int
}

func (e Withdrawn) Type() string { return "Withdrawn" }
func (e Withdrawn) Version() int { return e.Ver }

type Account struct {
	ID      string
	Balance int64
	Version int
}

func (a *Account) Apply(e Event) {
	switch evt := e.(type) {
	case Deposited:
		a.Balance += evt.Amount
		a.Version = evt.Ver
	case Withdrawn:
		a.Balance -= evt.Amount
		a.Version = evt.Ver
	}
}

func (a *Account) Deposit(amount int64) (Event, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	return Deposited{Amount: amount, Ver: a.Version + 1}, nil
}

func (a *Account) Withdraw(amount int64) (Event, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if a.Balance < amount {
		return nil, errors.New("insufficient funds")
	}
	return Withdrawn{Amount: amount, Ver: a.Version + 1}, nil
}

type EventStore interface {
	Append(streamID string, expectedVersion int, events ...Event) error
	Load(streamID string) ([]Event, error)
}

type InMemoryEventStore struct {
	streams map[string][]Event
}

func NewInMemoryEventStore() *InMemoryEventStore {
	return &InMemoryEventStore{streams: map[string][]Event{}}
}

func (s *InMemoryEventStore) Append(streamID string, expectedVersion int, events ...Event) error {
	current := s.streams[streamID]
	if expectedVersion != len(current) {
		return errors.New("version conflict")
	}
	s.streams[streamID] = append(current, events...)
	return nil
}

func (s *InMemoryEventStore) Load(streamID string) ([]Event, error) {
	return append([]Event(nil), s.streams[streamID]...), nil
}

func RehydrateAccount(id string, events []Event) *Account {
	a := &Account{ID: id}
	for _, e := range events {
		a.Apply(e)
	}
	return a
}
