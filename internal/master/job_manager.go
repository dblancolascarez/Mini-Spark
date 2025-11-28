// Gestión del estado de jobs
package master

import (
	"fmt"
	"sync"
	"time"
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
	JobID       string
	Name        string
	Status      JobStatus
	SubmittedAt time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	Progress    float64
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
func (jm *JobManager) CreateJob(jobID, name string) *JobInfo {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job := &JobInfo{
		JobID:       jobID,
		Name:        name,
		Status:      JobStatusAccepted,
		SubmittedAt: time.Now(),
		Progress:    0.0,
	}

	jm.jobs[jobID] = job
	fmt.Printf("[JobManager] Job created: %s (%s)\n", jobID, name)
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