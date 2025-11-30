package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dblancolascarez/mini-spark/internal/dag"
	"github.com/dblancolascarez/mini-spark/internal/master"
	"github.com/dblancolascarez/mini-spark/internal/protocol"
)

// Server maneja las peticiones HTTP
type Server struct {
	coordinator *master.Coordinator
	scheduler   *master.Scheduler
	jobManager  *master.JobManager
	jobExecutor *master.JobExecutor
	port        int
}

// NewServer crea un nuevo servidor API
func NewServer(coordinator *master.Coordinator, scheduler *master.Scheduler, jobManager *master.JobManager, port int) *Server {
	jobExecutor := master.NewJobExecutor(coordinator, scheduler, jobManager)
	
	return &Server{
		coordinator: coordinator,
		scheduler:   scheduler,
		jobManager:  jobManager,
		jobExecutor: jobExecutor,
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
	mux.HandleFunc("/api/v1/jobs", s.handleJobs)
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

	if req.WorkerID == "" || req.Host == "" || req.Port == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

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

// handleJobs maneja GET (listar) y POST (enviar) jobs
func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListJobs(w, r)
	case http.MethodPost:
		s.handleSubmitJob(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSubmitJob maneja el envío de jobs
func (s *Server) handleSubmitJob(w http.ResponseWriter, r *http.Request) {
	var req protocol.JobSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validar request
	if req.Name == "" {
		http.Error(w, "Job name is required", http.StatusBadRequest)
		return
	}

	// Parser y validar DAG
	parser := dag.NewParser()
	jobDAG, err := parser.BuildDAG(&req.DAG)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse DAG: %v", err), http.StatusBadRequest)
		return
	}

	validator := dag.NewValidator()
	if err := validator.Validate(jobDAG); err != nil {
		http.Error(w, fmt.Sprintf("DAG validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Generar ID de job
	jobID := fmt.Sprintf("job-%d", time.Now().Unix())

	// Crear job en JobManager
	job := s.jobManager.CreateJob(jobID, req.Name, jobDAG)

	// Ejecutar job en background
	go func() {
		if err := s.jobExecutor.ExecuteJob(job); err != nil {
			fmt.Printf("[API] Job %s execution failed: %v\n", jobID, err)
			s.jobManager.UpdateJobStatus(jobID, master.JobStatusFailed)
		}
	}()

	resp := protocol.JobSubmitResponse{
		JobID:   jobID,
		Status:  string(master.JobStatusAccepted),
		Message: "Job accepted and queued for execution",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

// handleListJobs lista todos los jobs
func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	jobs := s.jobManager.ListJobs()

	jobsResponse := make([]map[string]interface{}, len(jobs))
	for i, job := range jobs {
		jobsResponse[i] = map[string]interface{}{
			"job_id":   job.JobID,
			"name":     job.Name,
			"status":   job.Status,
			"progress": job.Progress,
		}
	}

	response := map[string]interface{}{
		"jobs":  jobsResponse,
		"count": len(jobs),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleJobStatus maneja consultas de estado de jobs
func (s *Server) handleJobStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extraer job ID de la URL
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/jobs/")
	jobID := strings.TrimSuffix(path, "/")

	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	statusResp, err := s.jobManager.GetJobStatusResponse(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statusResp)
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