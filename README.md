# Ige In-Memory Circuit Breaker & Fallback Registry

A thread-safe, concurrent Go implementation of the **Circuit Breaker** stability pattern. Eliminates cascade failure in microservice architectures by detecting repetitive downstream faults, opening the circuit to reject further load, gracefully invoking fallback handlers, and probing downstream health via half-open state testing.

## State Transitions
1. **CLOSED**: Normal operation. Failures increment an internal counter.
2. **OPEN**: Tripped when failure threshold is met. Rejects upstream calls immediately without stressing downstream APIs.
3. **HALF-OPEN**: Tripped after a configurable cool-down timeout. Allows probe traffic; returns to CLOSED on success or flips back to OPEN on continued errors.

## Quick Start
```bash
# Run CLI demonstration
go run main.go

# Run unit tests
go test -v ./...
```
