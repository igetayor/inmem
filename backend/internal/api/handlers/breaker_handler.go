// Package handlers contains the HTTP handlers (controllers) for the
// Circuit Breaker REST API. Each handler is a thin transport adapter:
// parse request → call domain → write typed JSON response.
//
// Routes registered by cmd/server/main.go:
//   GET  /api/status   → Status
//   POST /api/execute  → Execute
//   POST /api/reset    → Reset
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"ige-inmemory-circuit-breaker/backend/internal/domain/breaker"
)

// RequestLogItem records a single request's outcome for the live activity feed.
type RequestLogItem struct {
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"` // SUCCESS | FAILED | SHORT_CIRCUITED | FALLBACK
	Message   string `json:"message"`
	State     string `json:"state"`
}

// ExecuteRequest is the inbound DTO for POST /api/execute.
type ExecuteRequest struct {
	SimulateFailure bool   `json:"simulate_failure"`
	ServiceName     string `json:"service_name"`
}

// Handler holds domain dependencies injected by the server bootstrap.
type Handler struct {
	cb        *breaker.CircuitBreaker
	logMu     sync.Mutex
	recentLog []RequestLogItem
}

// New creates a Handler with the shared CircuitBreaker instance.
func New(cb *breaker.CircuitBreaker) *Handler {
	return &Handler{cb: cb}
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) appendLog(status, msg, state string) {
	h.logMu.Lock()
	defer h.logMu.Unlock()
	h.recentLog = append(h.recentLog, RequestLogItem{
		Timestamp: time.Now().Format("15:04:05.000"),
		Status:    status,
		Message:   msg,
		State:     state,
	})
	if len(h.recentLog) > 30 {
		h.recentLog = h.recentLog[1:]
	}
}

// Status returns telemetry snapshot + recent request history.
//
//	GET /api/status
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	telem := h.cb.GetTelemetry()
	h.logMu.Lock()
	history := make([]RequestLogItem, len(h.recentLog))
	copy(history, h.recentLog)
	h.logMu.Unlock()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"telemetry": telem,
		"history":   history,
	})
}

// Execute runs a (possibly simulated) downstream call through the circuit breaker.
//
//	POST /api/execute
func (h *Handler) Execute(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req ExecuteRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.ServiceName == "" {
		req.ServiceName = "Payment Gateway"
	}

	mockDownstream := func() error {
		if req.SimulateFailure {
			return errors.New("HTTP 503 Service Unavailable (Downstream Pod Unresponsive)")
		}
		return nil
	}

	fallbackHandler := func(err error) error {
		if errors.Is(err, breaker.ErrCircuitOpen) {
			h.appendLog("SHORT_CIRCUITED", fmt.Sprintf("Rejected %s (Fail-Fast)", req.ServiceName), "OPEN")
			return err
		}
		h.appendLog("FALLBACK", fmt.Sprintf("Served degraded response for %s: %v", req.ServiceName, err), string(h.cb.GetTelemetry().State))
		return nil
	}

	err := h.cb.Execute(mockDownstream, fallbackHandler)
	telem := h.cb.GetTelemetry()
	response := map[string]interface{}{
		"circuit_state": telem.State,
		"success":       err == nil,
	}
	if err != nil {
		if errors.Is(err, breaker.ErrCircuitOpen) {
			response["error"] = "Circuit breaker is OPEN. Traffic rejected immediately."
			response["short_circuited"] = true
		} else {
			response["error"] = err.Error()
			response["short_circuited"] = false
		}
	} else {
		response["message"] = fmt.Sprintf("Successfully invoked %s", req.ServiceName)
		h.appendLog("SUCCESS", fmt.Sprintf("OK from %s", req.ServiceName), string(telem.State))
	}
	writeJSON(w, http.StatusOK, response)
}

// Reset manually closes the circuit breaker.
//
//	POST /api/reset
func (h *Handler) Reset(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	h.cb.Reset()
	h.appendLog("INFO", "Circuit breaker manually RESET to CLOSED", "CLOSED")
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}
