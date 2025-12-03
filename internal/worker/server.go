package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/dblancolascarez/mini-spark/internal/protocol"
)

// Server maneja requests HTTP del master
type Server struct {
	workerID string
	executor *Executor
	port     int
	mu       sync.Mutex
}

// NewServer crea un nuevo servidor worker
func NewServer(workerID string, port int) *Server {
	return &Server{
		workerID: workerID,
		executor: NewExecutor(workerID),
		port:     port,
	}
}

// Start inicia el servidor HTTP
func (s *Server) Start() error {
	http.HandleFunc("/api/v1/tasks/execute", s.handleExecuteTask)
	http.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("[Worker Server] Listening on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// handleExecuteTask ejecuta una tarea asignada por el master
func (s *Server) handleExecuteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var task protocol.TaskAssignment
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	fmt.Printf("[Worker Server] Received task: %s\n", task.TaskID)

	// Ejecutar tarea
	result, err := s.executor.ExecuteTask(&task)
	if err != nil {
		result = &protocol.TaskResult{
			TaskID:   task.TaskID,
			JobID:    task.JobID,
			WorkerID: s.workerID,
			Status:   "FAILED",
			Error:    err.Error(),
		}
	} else {
		result.WorkerID = s.workerID
		result.JobID = task.JobID
	}

	// Responder con resultado
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleHealth responde con estado del worker
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"worker_id": s.workerID,
	})
}
