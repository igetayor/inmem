package main

import (
	"errors"
	"fmt"
	"time"

	breaker "ige-inmemory-circuit-breaker/breaker"
)

func main() {
	fmt.Println("============================================================")
	fmt.Println("🔌 Ige In-Memory Circuit Breaker & Fallback Registry")
	fmt.Println("============================================================")

	// Trip after 2 failures, 500ms reset timeout
	cb := breaker.NewCircuitBreaker(2, 500*time.Millisecond)

	fallbackFn := func(err error) error {
		fmt.Printf("   -> [FALLBACK ACTIVATED] Gracefully serving degraded response: %v\n", err)
		return nil
	}

	flakyService := func(shouldFail bool) func() error {
		return func() error {
			if shouldFail {
				return errors.New("upstream HTTP 503 Service Unavailable")
			}
			return nil
		}
	}

	fmt.Println("\n[1] Executing healthy request in CLOSED state:")
	_ = cb.Execute(flakyService(false), fallbackFn)
	fmt.Printf("   State: %s\n", cb.GetState())

	fmt.Println("\n[2] Triggering first failure:")
	_ = cb.Execute(flakyService(true), fallbackFn)
	fmt.Printf("   State: %s\n", cb.GetState())

	fmt.Println("\n[3] Triggering second failure (TRIPPING BREAKER):")
	_ = cb.Execute(flakyService(true), fallbackFn)
	fmt.Printf("   State: %s\n", cb.GetState())

	fmt.Println("\n[4] Attempting request while OPEN (Traffic rejected immediately without touching downstream):")
	_ = cb.Execute(flakyService(false), fallbackFn)
	fmt.Printf("   State: %s\n", cb.GetState())

	fmt.Println("\n Sleeping 600ms to allow cool-down into HALF-OPEN state...")
	time.Sleep(600 * time.Millisecond)
	fmt.Printf("   State after cool-down: %s\n", cb.GetState())

	fmt.Println("\n[5] Executing successful probe in HALF-OPEN state (HEALING CIRCUIT):")
	_ = cb.Execute(flakyService(false), fallbackFn)
	fmt.Printf("   State after successful probe: %s\n", cb.GetState())

	fmt.Println("\n✅ Circuit Breaker demonstrated state transitions: CLOSED -> OPEN -> HALF-OPEN -> CLOSED.")
}
