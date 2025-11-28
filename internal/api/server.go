// Servidor HTTP
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dblancolascarez/mini-spark/internal/master"
	"github.com/dblancolascarez/mini-spark/internal/protocol"
)

// Server maneja las peticiones HTTP
type Server struct {
	coordinator *master.Coordinator
	scheduler   *master.Scheduler
	port        int
}

// NewServer crea un nuevo servidor API
func NewServer(coordinator *master.Coordinator, scheduler *master.Scheduler, port int) *Server {
	return &Server{
		coordinator: coordinator,
		scheduler:   scheduler,
		port:        port,
	}
}

// Start inicia el servidor HTTP
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Endpoints de workers
	mux.HandleFunc("/api/v1/workers/register", s.handleRegisterWorker)
	mux.HandleFunc("/api/v1/workers/heartbeat", s.handleHeartbeat)
	mux.HandleFunc("/api/v1/workers", s.handleListWorkers)

	// Endpoints de jobs
	mux.HandleFunc("/api/v1/jobs", s.handleSubmitJob)
	mux.HandleFunc("/api/v1/jobs/", s.handleJobStatus)

	// Health check
	mux.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("[API] Server listening on %s\n", addr)

	return http.ListenAndServe(addr, s.loggingMiddleware(mux))
}

// handleRegisterWorker maneja el registro de workers
func (s *Server) handleRegisterWorker(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.WorkerRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validar campos requeridos
	if req.WorkerID == "" || req.Host == "" || req.Port == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Registrar worker
	if err := s.coordinator.RegisterWorker(req); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	resp := protocol.WorkerRegisterResponse{
		Success: true,
		Message: fmt.Sprintf("Worker %s registered successfully", req.WorkerID),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleHeartbeat maneja los heartbeats de workers
func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.coordinator.UpdateHeartbeat(req); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp := protocol.HeartbeatResponse{
		Acknowledged: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleListWorkers lista todos los workers
func (s *Server) handleListWorkers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	workers := s.coordinator.GetActiveWorkers()
	stats := s.coordinator.GetWorkerStats()

	response := map[string]interface{}{
		"workers": workers,
		"stats":   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSubmitJob maneja el envío de jobs
func (s *Server) handleSubmitJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.JobSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implementar en Semana 2
	resp := protocol.JobSubmitResponse{
		JobID:   fmt.Sprintf("job-%d", time.Now().Unix()),
		Status:  "ACCEPTED",
		Message: "Job accepted (execution not implemented yet)",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

// handleJobStatus maneja consultas de estado de jobs
func (s *Server) handleJobStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implementar en Semana 2
	response := map[string]interface{}{
		"status":  "NOT_IMPLEMENTED",
		"message": "Job status tracking will be implemented in Week 2",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealth endpoint de salud
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	active, total := s.coordinator.GetWorkerCount()
	
	health := map[string]interface{}{
		"status":         "healthy",
		"timestamp":      time.Now(),
		"active_workers": active,
		"total_workers":  total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// loggingMiddleware registra todas las peticiones
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fmt.Printf("[API] %s %s from %s\n", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		fmt.Printf("[API] %s %s completed in %v\n", r.Method, r.URL.Path, time.Since(start))
	})
}
