package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrRateLimited     = errors.New("rate limited")
	ErrCircuitOpen     = errors.New("circuit breaker open")
	ErrTooManyHalfOpen = errors.New("too many requests in half-open")
)

// RateLimiter is a simple token-bucket limiter.
type RateLimiter struct {
	tokens chan struct{}
	stopCh chan struct{}
	once   sync.Once
}

// NewRateLimiter creates limiter with fixed refill rate and burst capacity.
func NewRateLimiter(ratePerSec int, burst int) *RateLimiter {
	if ratePerSec <= 0 {
		ratePerSec = 1
	}
	if burst <= 0 {
		burst = 1
	}
	rl := &RateLimiter{
		tokens: make(chan struct{}, burst),
		stopCh: make(chan struct{}),
	}
	for i := 0; i < burst; i++ {
		rl.tokens <- struct{}{}
	}

	interval := time.Second / time.Duration(ratePerSec)
	if interval <= 0 {
		interval = time.Nanosecond
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-rl.stopCh:
				return
			case <-ticker.C:
				select {
				case rl.tokens <- struct{}{}:
				default:
				}
			}
		}
	}()
	return rl
}

// Allow returns immediately.
func (r *RateLimiter) Allow() bool {
	select {
	case <-r.tokens:
		return true
	default:
		return false
	}
}

// Wait blocks until token available or context canceled.
func (r *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.tokens:
		return nil
	}
}

func (r *RateLimiter) Stop() {
	r.once.Do(func() {
		close(r.stopCh)
	})
}

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker protects downstream dependency.
type CircuitBreaker struct {
	mu sync.Mutex

	failureThreshold int
	openTimeout      time.Duration
	halfOpenMax      int

	state State

	failures int
	openedAt time.Time
	inFlight int
}

func NewCircuitBreaker(failureThreshold int, openTimeout time.Duration, halfOpenMax int) *CircuitBreaker {
	if failureThreshold <= 0 {
		failureThreshold = 3
	}
	if openTimeout <= 0 {
		openTimeout = 200 * time.Millisecond
	}
	if halfOpenMax <= 0 {
		halfOpenMax = 1
	}
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		openTimeout:      openTimeout,
		halfOpenMax:      halfOpenMax,
		state:            StateClosed,
	}
}

func (c *CircuitBreaker) State() State {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.advanceStateLocked()
	return c.state
}

// Execute runs fn under circuit protection.
func (c *CircuitBreaker) Execute(fn func() error) error {
	if err := c.beforeRequest(); err != nil {
		return err
	}

	err := fn()
	c.afterRequest(err)
	return err
}

func (c *CircuitBreaker) beforeRequest() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.advanceStateLocked()
	switch c.state {
	case StateOpen:
		return ErrCircuitOpen
	case StateHalfOpen:
		if c.inFlight >= c.halfOpenMax {
			return ErrTooManyHalfOpen
		}
		c.inFlight++
	}
	return nil
}

func (c *CircuitBreaker) afterRequest(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == StateHalfOpen && c.inFlight > 0 {
		c.inFlight--
	}

	if err == nil {
		switch c.state {
		case StateHalfOpen:
			c.state = StateClosed
			c.failures = 0
		case StateClosed:
			c.failures = 0
		}
		return
	}

	switch c.state {
	case StateClosed:
		c.failures++
		if c.failures >= c.failureThreshold {
			c.state = StateOpen
			c.openedAt = time.Now()
		}
	case StateHalfOpen:
		c.state = StateOpen
		c.openedAt = time.Now()
	}
}

func (c *CircuitBreaker) advanceStateLocked() {
	if c.state == StateOpen && time.Since(c.openedAt) >= c.openTimeout {
		c.state = StateHalfOpen
		c.inFlight = 0
	}
}


