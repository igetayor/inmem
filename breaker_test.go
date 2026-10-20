package breaker

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreakerTripping(t *testing.T) {
	cb := NewCircuitBreaker(2, 200*time.Millisecond)

	failOp := func() error {
		return errors.New("mock failure")
	}
	successOp := func() error {
		return nil
	}

	// 1st failure
	_ = cb.Execute(failOp, nil)
	if cb.GetState() != StateClosed {
		t.Errorf("Expected state CLOSED after 1 failure, got %s", cb.GetState())
	}

	// 2nd failure -> Trips into OPEN
	_ = cb.Execute(failOp, nil)
	if cb.GetState() != StateOpen {
		t.Errorf("Expected state OPEN after 2 failures, got %s", cb.GetState())
	}

	// Immediate next request rejected with ErrCircuitOpen
	err := cb.Execute(successOp, nil)
	if err != ErrCircuitOpen {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}

	// Wait for reset timeout -> HALF-OPEN
	time.Sleep(250 * time.Millisecond)
	if cb.GetState() != StateHalfOpen {
		t.Errorf("Expected state HALF-OPEN after cool-down, got %s", cb.GetState())
	}

	// Successful probe in HALF-OPEN closes circuit
	err = cb.Execute(successOp, nil)
	if err != nil {
		t.Errorf("Expected success, got %v", err)
	}
	if cb.GetState() != StateClosed {
		t.Errorf("Expected state CLOSED after successful probe, got %s", cb.GetState())
	}
}

func TestCircuitBreakerFallback(t *testing.T) {
	cb := NewCircuitBreaker(1, 100*time.Millisecond)
	fallbackRan := false

	_ = cb.Execute(func() error {
		return errors.New("downstream error")
	}, func(err error) error {
		fallbackRan = true
		return nil
	})

	if !fallbackRan {
		t.Errorf("Expected fallback to execute on error")
	}
}
