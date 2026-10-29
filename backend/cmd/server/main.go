// Package main is the server entrypoint for the In-Memory Circuit Breaker service.
// Wires domain (breaker) → handlers → HTTP mux → server.
//
// Usage:
//
//	go run ./backend/cmd/server           (CLI demo)
//	go run ./backend/cmd/server -serve    (HTTP server, port 8092)
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"ige-inmemory-circuit-breaker/backend/internal/api/handlers"
	"ige-inmemory-circuit-breaker/backend/internal/domain/breaker"
)

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

func startServer(port string) {
	cb := breaker.NewCircuitBreaker(3, 3000*time.Millisecond)
	h := handlers.New(cb)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", h.Status)
	mux.HandleFunc("/api/execute", h.Execute)
	mux.HandleFunc("/api/reset", h.Reset)

	workDir, _ := os.Getwd()
	frontendDir := filepath.Join(workDir, "frontend")
	if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
		frontendDir = filepath.Join(workDir, "static") // graceful fallback
	}
	mux.Handle("/", http.FileServer(http.Dir(frontendDir)))

	addr := ":" + port
	fmt.Printf("[*] In-Memory Circuit Breaker Server  →  http://127.0.0.1%s\n", addr)
	fmt.Println("    API: GET /api/status | POST /api/execute | POST /api/reset")

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server error: %v", err)
	}
}

func runDemo() {
	fmt.Println("============================================================")
	fmt.Println("[*] Ige In-Memory Circuit Breaker & Fallback Registry")
	fmt.Println("============================================================")

	cb := breaker.NewCircuitBreaker(2, 600*time.Millisecond)
	fallbackFn := func(err error) error {
		fmt.Printf("   -> [FALLBACK ACTIVATED] Gracefully serving degraded response: %v\n", err)
		return nil
	}
	flaky := func(shouldFail bool) func() error {
		return func() error {
			if shouldFail {
				return errors.New("upstream HTTP 503 Service Unavailable")
			}
			return nil
		}
	}

	fmt.Println("\n[1] Executing healthy request in CLOSED state:")
	_ = cb.Execute(flaky(false), fallbackFn)
	fmt.Printf("   State: %s\n", cb.GetTelemetry().State)

	fmt.Println("\n[2] Triggering first failure:")
	_ = cb.Execute(flaky(true), fallbackFn)
	fmt.Printf("   State: %s (Failures: %d/%d)\n", cb.GetTelemetry().State, cb.GetTelemetry().FailureCount, cb.GetTelemetry().MaxFailures)

	fmt.Println("\n[3] Triggering second failure (TRIPPING BREAKER):")
	_ = cb.Execute(flaky(true), fallbackFn)
	fmt.Printf("   State: %s\n", cb.GetTelemetry().State)

	fmt.Println("\n[4] Attempting request while OPEN (traffic rejected without touching downstream):")
	_ = cb.Execute(flaky(false), fallbackFn)
	fmt.Printf("   State: %s\n", cb.GetTelemetry().State)

	fmt.Println("\n[*] Sleeping 700ms to allow cool-down into HALF-OPEN...")
	time.Sleep(700 * time.Millisecond)
	fmt.Printf("   State after cool-down: %s\n", cb.GetTelemetry().State)

	fmt.Println("\n[5] Successful probe in HALF-OPEN (HEALING CIRCUIT):")
	_ = cb.Execute(flaky(false), fallbackFn)
	_ = cb.Execute(flaky(false), fallbackFn)
	fmt.Printf("   State after successful probes: %s\n", cb.GetTelemetry().State)

	fmt.Println("\n[+] CLOSED -> OPEN -> HALF-OPEN -> CLOSED demonstrated.")
	fmt.Println("    To launch the interactive visualizer: go run ./backend/cmd/server -serve")
}
