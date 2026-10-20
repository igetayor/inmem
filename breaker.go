package breaker

import (
	"errors"
	"sync"
	"time"
)

type State string

const (
	StateClosed   State = "CLOSED"    // Normal operation; traffic flows freely
	StateOpen     State = "OPEN"      // Trip threshold reached; all traffic rejected/fallback executed
	StateHalfOpen State = "HALF-OPEN" // Cool-down elapsed; probing downstream health with test requests
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is OPEN; request rejected to prevent cascade failure")
)

type CircuitBreaker struct {
	mu           sync.Mutex
	state        State
	failureCount int
	maxFailures  int
	resetTimeout time.Duration
	lastFailure  time.Time
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        StateClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

// Execute wraps an operation with circuit-breaker protection and optional fallback.
func (cb *CircuitBreaker) Execute(op func() error, fallback func(err error) error) error {
	cb.mu.Lock()

	// Check if state should transition from OPEN to HALF-OPEN
	if cb.state == StateOpen && time.Since(cb.lastFailure) > cb.resetTimeout {
		cb.state = StateHalfOpen
	}

	if cb.state == StateOpen {
		cb.mu.Unlock()
		if fallback != nil {
			return fallback(ErrCircuitOpen)
		}
		return ErrCircuitOpen
	}

	cb.mu.Unlock()

	// Execute operation
	err := op()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailure = time.Now()

		if cb.failureCount >= cb.maxFailures || cb.state == StateHalfOpen {
			cb.state = StateOpen
		}

		if fallback != nil {
			return fallback(err)
		}
		return err
	}

	// Success: if we were HALF-OPEN or CLOSED, reset counter
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
	}
	cb.failureCount = 0
	return nil
}

// GetState returns current circuit breaker state.
func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state == StateOpen && time.Since(cb.lastFailure) > cb.resetTimeout {
		return StateHalfOpen
	}
	return cb.state
}
