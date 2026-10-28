# Senior Engineering Walkthrough: Ige In-Memory Circuit Breaker

> **Target Audience:** Principal Infrastructure Engineers, Distributed Systems Architects, Staff SREs.  
> **Key Architecture Highlights:** Zero External Dependencies (pure Go `sync` and `net/http`), Atomic State Transitions (Closed / Open / Half-Open), Fail-Fast Microsecond Short-Circuiting, Configurable Cooldown & Probe Rates.

---

## 1. Executive Summary & Problem Space

In microservice architectures, when a downstream dependency (e.g. database cluster, payment partner, authentication service) experiences degradation or outage, upstream services repeatedly hammering it leads to **cascading resource exhaustion**: thread pools saturate, HTTP connection queues fill up, and the entire system collapses.

The **Circuit Breaker Pattern** provides automatic fault isolation:
1. **`CLOSED`**: Traffic flows uninterrupted. Failures are tracked.
2. **`OPEN`**: Once the failure threshold is hit, the breaker trips. All subsequent calls are rejected or diverted to a fallback in sub-microsecond time without touching the downstream network.
3. **`HALF-OPEN`**: After a cooldown period, a small number of canary probe requests are permitted through. If they succeed, the circuit self-heals back to `CLOSED`. If any probe fails, it trips back to `OPEN`.

---

## 2. File & Function Code Architecture

### `breaker/breaker.go` — Core Thread-Safe State Machine
- `type State string`: Enumerated types `StateClosed`, `StateOpen`, `StateHalfOpen`.
- `type Telemetry struct`: Snapshot of failure counts, remaining cooldown, total requests, and tripped count.
- `type CircuitBreaker struct`: Guarded by `sync.RWMutex` to prevent race conditions during high concurrent traffic.
- `NewCircuitBreaker(maxFailures int, resetTimeout time.Duration)`: Factory constructor initializing counters and probe thresholds.
- `(cb *CircuitBreaker) Execute(op func() error, fallback func(err error) error) error`:
  - Acquires lock to evaluate whether cooldown has elapsed.
  - If `OPEN`, immediately releases lock and invokes `fallback(ErrCircuitOpen)` without executing `op()`.
  - Executes `op()`. If error returned, increments failure counter and trips circuit if threshold exceeded.
  - If success in `HALF-OPEN`, increments consecutive success counter until probe target is met, transitioning circuit back to `CLOSED`.
- `(cb *CircuitBreaker) GetTelemetry() Telemetry`: Returns atomic copy of runtime counters and remaining cooldown milliseconds.
- `(cb *CircuitBreaker) Reset()`: Administrative override forcing circuit back to `CLOSED`.

### `server.go` — Embedded Zero-Dependency HTTP Gateway (Port 8092)
- `handleStatus(w, r)`: Returns live JSON telemetry and recent request history.
- `handleExecute(w, r)`: Executes simulated downstream service calls with configurable failure flags.
- `handleReset(w, r)`: Administrative endpoint resetting circuit state.
- `startServer(port)`: Spawns the HTTP listener with CORS and static file routing.

### `main.go` — Dual CLI Runner & Server Bootstrap
- `runDemo()`: Terminal execution demonstrating state transitions and fallback handling.
- `main()`: Parses `-serve` and `-port` flags.

### `static/` — Modular Frontend Visualizer
- `static/css/styles.css`: Dark mode, glowing state ring indicator, and real-time transition animations.
- `static/js/api.js`: REST client module.
- `static/js/components/StateDial.js`: Large central state dial with live cooldown countdown.
- `static/js/components/MetricsCards.js`: Failure counters, trip totals, and error rate percentages.
- `static/js/components/RequestStream.js`: Real-time request log with color-coded status badges.
- `static/js/app.js`: Connects components and polls every 500ms.

---

## 3. How to Run the Live Demo

### Option A: Interactive Web Mission Control (Port 8092)
```powershell
go run . -serve
# or run the precompiled binary:
.\ige-inmemory-circuit-breaker.exe -serve
```
Open your browser to: **http://127.0.0.1:8092**

**Demo Narrative Walkthrough:**
1. Point to the glowing green **`CLOSED`** state dial. Click **Send Healthy Request** several times; observe instant 200 OK responses.
2. Click **Inject Downstream Failure** 3 times.
3. On the 3rd failure, watch the state dial pulse red and snap to **`OPEN`**.
4. Try clicking **Send Healthy Request** while OPEN: show the audience that it is rejected instantly as **SHORT_CIRCUITED**, protecting the downstream service from load.
5. Watch the cooldown timer count down from 3.0s to 0.0s. The dial turns yellow to **`HALF-OPEN`**.
6. Send 2 consecutive healthy probe requests; observe the dial transition back to green **`CLOSED`** (self-healed).

### Option B: Terminal CLI Execution
```powershell
.\ige-inmemory-circuit-breaker.exe
```
Prints the state machine sequence in pure ASCII without dependencies.
