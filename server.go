package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	breaker "ige-inmemory-circuit-breaker/breaker"
)

const defaultPort = "8092"

var (
	cb        = breaker.NewCircuitBreaker(3, 3000*time.Millisecond)
	logMu     sync.Mutex
	recentLog []RequestLogItem
)

type RequestLogItem struct {
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"` // SUCCESS, FAILED, SHORT_CIRCUITED
	Message   string `json:"message"`
	State     string `json:"state"`
}

type ExecuteRequest struct {
	SimulateFailure bool   `json:"simulate_failure"`
	ServiceName     string `json:"service_name"`
}

func logRequest(status, msg, state string) {
	logMu.Lock()
	defer logMu.Unlock()
	entry := RequestLogItem{
		Timestamp: time.Now().Format("15:04:05.000"),
		Status:    status,
		Message:   msg,
		State:     state,
	}
	recentLog = append(recentLog, entry)
	if len(recentLog) > 30 {
		recentLog = recentLog[1:]
	}
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	telem := cb.GetTelemetry()
	logMu.Lock()
	history := make([]RequestLogItem, len(recentLog))
	copy(history, recentLog)
	logMu.Unlock()

	resp := map[string]interface{}{
		"telemetry": telem,
		"history":   history,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleExecute(w http.ResponseWriter, r *http.Request) {
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
			logRequest("SHORT_CIRCUITED", fmt.Sprintf("Rejected %s (Fail-Fast)", req.ServiceName), "OPEN")
			return err
		}
		logRequest("FALLBACK", fmt.Sprintf("Served degraded response for %s: %v", req.ServiceName, err), string(cb.GetTelemetry().State))
		return nil
	}

	err := cb.Execute(mockDownstream, fallbackHandler)
	telem := cb.GetTelemetry()

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
		logRequest("SUCCESS", fmt.Sprintf("OK from %s", req.ServiceName), string(telem.State))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	cb.Reset()
	logRequest("INFO", "Circuit breaker manually RESET to CLOSED", "CLOSED")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func startServer(port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/execute", handleExecute)
	mux.HandleFunc("/api/reset", handleReset)

	// Static file handler
	workDir, _ := os.Getwd()
	staticDir := filepath.Join(workDir, "static")
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		staticDir = workDir
	}

	fileServer := http.FileServer(http.Dir(staticDir))
	mux.Handle("/", fileServer)

	addr := ":" + port
	fmt.Printf("[*] In-Memory Circuit Breaker HTTP Server running at http://127.0.0.1%s\n", addr)
	fmt.Println("    Endpoints: GET /api/status | POST /api/execute | POST /api/reset")

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
