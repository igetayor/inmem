package breaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the operational state of the circuit breaker.
type State string

const (
	StateClosed   State = "CLOSED"    // Normal operation; traffic flows freely
	StateOpen     State = "OPEN"      // Trip threshold reached; traffic rejected/diverted to fallback
	StateHalfOpen State = "HALF-OPEN" // Cool-down elapsed; probing downstream health with test requests
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is OPEN; request rejected to prevent cascade failure")
)

// Telemetry holds runtime statistics for health inspection and observability.
type Telemetry struct {
	State              State         `json:"state"`
	FailureCount       int           `json:"failure_count"`
	MaxFailures        int           `json:"max_failures"`
	ResetTimeoutMs     int64         `json:"reset_timeout_ms"`
	RemainingCoolDown  int64         `json:"remaining_cooldown_ms"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	TrippedCount       int           `json:"tripped_count"`
	LastStateChange    time.Time     `json:"last_state_change"`
}

// CircuitBreaker manages fault tolerance and failure isolation for downstream service calls.
type CircuitBreaker struct {
	mu                 sync.RWMutex
	state              State
	failureCount       int
	maxFailures        int
	consecutiveSuccess int
	halfOpenProbes     int
	resetTimeout       time.Duration
	lastFailure        time.Time
	lastStateChange    time.Time

	// Telemetry counters
	totalRequests      int64
	successfulRequests int64
	failedRequests     int64
	trippedCount       int
}

// NewCircuitBreaker initializes a thread-safe circuit breaker with specified failure threshold and reset cooldown.
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:           StateClosed,
		maxFailures:     maxFailures,
		halfOpenProbes:  2,
		resetTimeout:    resetTimeout,
		lastStateChange: time.Now(),
	}
}

// Execute wraps an operation with circuit-breaker protection and optional fallback.
func (cb *CircuitBreaker) Execute(op func() error, fallback func(err error) error) error {
	cb.mu.Lock()
	cb.totalRequests++

	// Check if state should transition from OPEN to HALF-OPEN due to elapsed cooldown
	if cb.state == StateOpen && time.Since(cb.lastFailure) > cb.resetTimeout {
		cb.state = StateHalfOpen
		cb.consecutiveSuccess = 0
		cb.lastStateChange = time.Now()
	}

	currentState := cb.state
	if currentState == StateOpen {
		cb.mu.Unlock()
		if fallback != nil {
			return fallback(ErrCircuitOpen)
		}
		return ErrCircuitOpen
	}
	cb.mu.Unlock()

	// Execute downstream operation
	err := op()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failedRequests++
		cb.failureCount++
		cb.lastFailure = time.Now()

		// In HALF-OPEN or exceeding threshold in CLOSED, trip the circuit
		if cb.failureCount >= cb.maxFailures || cb.state == StateHalfOpen {
			if cb.state != StateOpen {
				cb.trippedCount++
				cb.lastStateChange = time.Now()
			}
			cb.state = StateOpen
		}

		if fallback != nil {
			return fallback(err)
		}
		return err
	}

	// Success handling
	cb.successfulRequests++
	if cb.state == StateHalfOpen {
		cb.consecutiveSuccess++
		if cb.consecutiveSuccess >= cb.halfOpenProbes {
			cb.state = StateClosed
			cb.failureCount = 0
			cb.consecutiveSuccess = 0
			cb.lastStateChange = time.Now()
		}
	} else if cb.state == StateClosed {
		cb.failureCount = 0
	}

	return nil
}

// GetTelemetry returns an atomic snapshot of circuit breaker health and counters.
func (cb *CircuitBreaker) GetTelemetry() Telemetry {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check cool-down transition
	if cb.state == StateOpen && time.Since(cb.lastFailure) > cb.resetTimeout {
		cb.state = StateHalfOpen
		cb.consecutiveSuccess = 0
		cb.lastStateChange = time.Now()
	}

	var remainingMs int64 = 0
	if cb.state == StateOpen {
		elapsed := time.Since(cb.lastFailure)
		if elapsed < cb.resetTimeout {
			remainingMs = (cb.resetTimeout - elapsed).Milliseconds()
		}
	}

	return Telemetry{
		State:              cb.state,
		FailureCount:       cb.failureCount,
		MaxFailures:        cb.maxFailures,
		ResetTimeoutMs:     cb.resetTimeout.Milliseconds(),
		RemainingCoolDown:  remainingMs,
		TotalRequests:      cb.totalRequests,
		SuccessfulRequests: cb.successfulRequests,
		FailedRequests:     cb.failedRequests,
		TrippedCount:       cb.trippedCount,
		LastStateChange:    cb.lastStateChange,
	}
}

// Reset manually forces the circuit breaker back to CLOSED state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failureCount = 0
	cb.consecutiveSuccess = 0
	cb.lastStateChange = time.Now()
}
