package master

import (
	"fmt"
	"sync"
	"time"

	"github.com/dblancolascarez/mini-spark/internal/dag"
	"github.com/dblancolascarez/mini-spark/internal/protocol"
)

// JobStatus representa el estado de un job
type JobStatus string

const (
	JobStatusAccepted  JobStatus = "ACCEPTED"
	JobStatusRunning   JobStatus = "RUNNING"
	JobStatusFailed    JobStatus = "FAILED"
	JobStatusSucceeded JobStatus = "SUCCEEDED"
)

// JobInfo contiene información de un job
type JobInfo struct {
	JobID          string
	Name           string
	Status         JobStatus
	DAG            *dag.DAG
	SubmittedAt    time.Time
	StartedAt      *time.Time
	CompletedAt    *time.Time
	Progress       float64
	TotalTasks     int
	CompletedTasks int
	FailedTasks    int
	Tasks          map[string]*TaskInfo
}

// TaskInfo contiene información de una tarea
type TaskInfo struct {
	TaskID      string
	JobID       string
	NodeID      string
	Operator    string
	Status      string // PENDING, RUNNING, COMPLETED, FAILED
	WorkerID    string
	StartedAt   *time.Time
	CompletedAt *time.Time
	Retries     int
	Error       string
}

// JobManager gestiona el ciclo de vida de los jobs
type JobManager struct {
	jobs map[string]*JobInfo
	mu   sync.RWMutex
}

// NewJobManager crea un nuevo gestor de jobs
func NewJobManager() *JobManager {
	return &JobManager{
		jobs: make(map[string]*JobInfo),
	}
}

// CreateJob crea un nuevo job
func (jm *JobManager) CreateJob(jobID, name string, jobDAG *dag.DAG) *JobInfo {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job := &JobInfo{
		JobID:          jobID,
		Name:           name,
		Status:         JobStatusAccepted,
		DAG:            jobDAG,
		SubmittedAt:    time.Now(),
		Progress:       0.0,
		TotalTasks:     len(jobDAG.Nodes),
		CompletedTasks: 0,
		FailedTasks:    0,
		Tasks:          make(map[string]*TaskInfo),
	}

	// Crear TaskInfo para cada nodo del DAG
	for nodeID, node := range jobDAG.Nodes {
		taskID := fmt.Sprintf("%s-task-%s", jobID, nodeID)
		job.Tasks[taskID] = &TaskInfo{
			TaskID:   taskID,
			JobID:    jobID,
			NodeID:   nodeID,
			Operator: node.Operator,
			Status:   "PENDING",
			Retries:  0,
		}
	}

	jm.jobs[jobID] = job
	fmt.Printf("[JobManager] Job created: %s (%s) with %d tasks\n", 
		jobID, name, job.TotalTasks)
	return job
}

// GetJob obtiene información de un job
func (jm *JobManager) GetJob(jobID string) (*JobInfo, error) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job %s not found", jobID)
	}
	return job, nil
}

// UpdateJobStatus actualiza el estado de un job
func (jm *JobManager) UpdateJobStatus(jobID string, status JobStatus) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}

	job.Status = status
	
	if status == JobStatusRunning && job.StartedAt == nil {
		now := time.Now()
		job.StartedAt = &now
	}
	
	if status == JobStatusSucceeded || status == JobStatusFailed {
		now := time.Now()
		job.CompletedAt = &now
	}

	fmt.Printf("[JobManager] Job %s status updated: %s\n", jobID, status)
	return nil
}

// UpdateTaskStatus actualiza el estado de una tarea
func (jm *JobManager) UpdateTaskStatus(taskID, status, workerID string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	// Buscar el job que contiene esta tarea
	var job *JobInfo
	var task *TaskInfo
	
	for _, j := range jm.jobs {
		if t, exists := j.Tasks[taskID]; exists {
			job = j
			task = t
			break
		}
	}

	if task == nil {
		return fmt.Errorf("task %s not found", taskID)
	}

	task.Status = status
	task.WorkerID = workerID

	now := time.Now()
	if status == "RUNNING" && task.StartedAt == nil {
		task.StartedAt = &now
	}

	if status == "COMPLETED" {
		task.CompletedAt = &now
		job.CompletedTasks++
	}

	if status == "FAILED" {
		task.CompletedAt = &now
		job.FailedTasks++
	}

	// Actualizar progreso del job
	job.Progress = float64(job.CompletedTasks) / float64(job.TotalTasks) * 100

	// Si todas las tareas completaron, marcar job como completado
	if job.CompletedTasks == job.TotalTasks {
		job.Status = JobStatusSucceeded
		completedAt := time.Now()
		job.CompletedAt = &completedAt
		fmt.Printf("[JobManager] Job %s COMPLETED (100%%)\n", job.JobID)
	}

	// Si alguna tarea falló, marcar job como fallido
	if job.FailedTasks > 0 && status == "FAILED" {
		job.Status = JobStatusFailed
		failedAt := time.Now()
		job.CompletedAt = &failedAt
		fmt.Printf("[JobManager] Job %s FAILED\n", job.JobID)
	}

	return nil
}

// ListJobs retorna todos los jobs
func (jm *JobManager) ListJobs() []*JobInfo {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	jobs := make([]*JobInfo, 0, len(jm.jobs))
	for _, job := range jm.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// GetJobStatusResponse convierte JobInfo a protocolo
func (jm *JobManager) GetJobStatusResponse(jobID string) (*protocol.JobStatusResponse, error) {
	job, err := jm.GetJob(jobID)
	if err != nil {
		return nil, err
	}

	resp := &protocol.JobStatusResponse{
		JobID:          job.JobID,
		Name:           job.Name,
		Status:         string(job.Status),
		Progress:       job.Progress,
		TotalTasks:     job.TotalTasks,
		CompletedTasks: job.CompletedTasks,
		FailedTasks:    job.FailedTasks,
	}

	if job.StartedAt != nil {
		resp.StartedAt = job.StartedAt.Format(time.RFC3339)
	}

	if job.CompletedAt != nil {
		resp.CompletedAt = job.CompletedAt.Format(time.RFC3339)
	}

	return resp, nil
}