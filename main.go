package main

import (
	"errors"
	"flag"
	"fmt"
	"time"

	breaker "ige-inmemory-circuit-breaker/breaker"
)

func runDemo() {
	fmt.Println("============================================================")
	fmt.Println("[*] Ige In-Memory Circuit Breaker & Fallback Registry")
	fmt.Println("============================================================")

	// Trip after 2 failures, 600ms reset timeout
	cb := breaker.NewCircuitBreaker(2, 600*time.Millisecond)

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
	fmt.Printf("   State: %s\n", cb.GetTelemetry().State)

	fmt.Println("\n[2] Triggering first failure:")
	_ = cb.Execute(flakyService(true), fallbackFn)
	fmt.Printf("   State: %s (Failures: %d/%d)\n", cb.GetTelemetry().State, cb.GetTelemetry().FailureCount, cb.GetTelemetry().MaxFailures)

	fmt.Println("\n[3] Triggering second failure (TRIPPING BREAKER):")
	_ = cb.Execute(flakyService(true), fallbackFn)
	fmt.Printf("   State: %s\n", cb.GetTelemetry().State)

	fmt.Println("\n[4] Attempting request while OPEN (Traffic rejected immediately without touching downstream):")
	_ = cb.Execute(flakyService(false), fallbackFn)
	fmt.Printf("   State: %s\n", cb.GetTelemetry().State)

	fmt.Println("\n[*] Sleeping 700ms to allow cool-down into HALF-OPEN state...")
	time.Sleep(700 * time.Millisecond)
	fmt.Printf("   State after cool-down: %s\n", cb.GetTelemetry().State)

	fmt.Println("\n[5] Executing successful probe in HALF-OPEN state (HEALING CIRCUIT):")
	_ = cb.Execute(flakyService(false), fallbackFn)
	_ = cb.Execute(flakyService(false), fallbackFn)
	fmt.Printf("   State after successful probes: %s\n", cb.GetTelemetry().State)

	fmt.Println("\n[+] Circuit Breaker demonstrated state transitions: CLOSED -> OPEN -> HALF-OPEN -> CLOSED.")
	fmt.Println("    To launch the interactive visualizer, run: go run . -serve")
}

func main() {
	serve := flag.Bool("serve", false, "Start interactive web server on port 8092")
	port := flag.String("port", "8092", "Port for web server")
	flag.Parse()

	if *serve {
		startServer(*port)
	} else {
		runDemo()
	}
}
