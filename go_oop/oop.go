package go_oop

import (
	"errors"
	"fmt"
	"sync"
)

// User demonstrates encapsulation:
// fields are private, state is updated through methods.
type User struct {
	id      int64
	name    string
	email   string
	balance int64
	mu      sync.Mutex
}

func NewUser(id int64, name, email string) (*User, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	if name == "" {
		return nil, errors.New("name is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	return &User{id: id, name: name, email: email}, nil
}

func (u *User) ID() int64     { return u.id }
func (u *User) Name() string  { return u.name }
func (u *User) Email() string { return u.email }

func (u *User) SetEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	u.email = email
	return nil
}

func (u *User) Deposit(amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	u.balance += amount
	return nil
}

func (u *User) Withdraw(amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	if amount > u.balance {
		return errors.New("insufficient balance")
	}
	u.balance -= amount
	return nil
}

func (u *User) Balance() int64 {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.balance
}

// Entity is a behavior-oriented abstraction.
type Entity interface {
	GetID() int64
	DisplayName() string
}

func (u *User) GetID() int64         { return u.ID() }
func (u *User) DisplayName() string  { return u.Name() }

// Profile is composed into Customer to simulate reusable capabilities.
type Profile struct {
	Address string
	Phone   string
}

func (p Profile) ContactInfo() string {
	return fmt.Sprintf("phone=%s,address=%s", p.Phone, p.Address)
}

// Customer demonstrates composition over inheritance.
type Customer struct {
	*User
	Profile
	Level string
}

func NewCustomer(id int64, name, email, level string) (*Customer, error) {
	u, err := NewUser(id, name, email)
	if err != nil {
		return nil, err
	}
	if level == "" {
		level = "standard"
	}
	return &Customer{
		User:  u,
		Level: level,
	}, nil
}

func (c *Customer) DisplayName() string {
	return fmt.Sprintf("%s[%s]", c.Name(), c.Level)
}

// Service showcases dependency inversion via interfaces.
type Service struct {
	store Store
}

type Store interface {
	Save(Entity) error
	FindByID(id int64) (Entity, error)
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) Register(entity Entity) error {
	if entity == nil {
		return errors.New("entity is nil")
	}
	return s.store.Save(entity)
}

func (s *Service) GetDisplayName(id int64) (string, error) {
	entity, err := s.store.FindByID(id)
	if err != nil {
		return "", err
	}
	return entity.DisplayName(), nil
}


