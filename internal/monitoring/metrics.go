package monitoring

import (
	"runtime"
	"sync"
	"time"
)

// Metrics almacena métricas del sistema
type Metrics struct {
	mu sync.RWMutex

	// Métricas por nodo
	NodeID           string
	CPUUsagePercent  float64
	MemoryUsageMB    int64
	ActiveTasks      int
	CompletedTasks   int
	FailedTasks      int
	AverageLatencyMs float64

	// Métricas por job
	JobMetrics map[string]*JobMetrics

	// Timestamps
	StartTime time.Time
	LastUpdate time.Time

	// Contadores acumulados
	TotalTasksExecuted int64
	TotalBytesProcessed int64
	TotalRetries int
}

// JobMetrics contiene métricas específicas de un job
type JobMetrics struct {
	JobID           string
	StartTime       time.Time
	EndTime         *time.Time
	DurationMs      int64
	TasksTotal      int
	TasksCompleted  int
	TasksFailed     int
	RecordsProcessed int64
	ThroughputRPS   float64 // Records per second
	Stages          map[string]*StageMetrics
}

// StageMetrics contiene métricas de una etapa del DAG
type StageMetrics struct {
	StageID         string
	StartTime       time.Time
	EndTime         *time.Time
	DurationMs      int64
	RecordsInput    int64
	RecordsOutput   int64
	Retries         int
}

// NewMetrics crea un nuevo sistema de métricas
func NewMetrics(nodeID string) *Metrics {
	return &Metrics{
		NodeID:     nodeID,
		JobMetrics: make(map[string]*JobMetrics),
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
	}
}

// UpdateSystemMetrics actualiza métricas del sistema
func (m *Metrics) UpdateSystemMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Obtener estadísticas de memoria
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.MemoryUsageMB = int64(memStats.Alloc / 1024 / 1024)

	// CPU usage (aproximado basado en goroutines)
	numGoroutines := runtime.NumGoroutine()
	numCPU := runtime.NumCPU()
	m.CPUUsagePercent = float64(numGoroutines) / float64(numCPU*10) * 100
	if m.CPUUsagePercent > 100 {
		m.CPUUsagePercent = 100
	}

	m.LastUpdate = time.Now()
}

// RecordTaskStart registra el inicio de una tarea
func (m *Metrics) RecordTaskStart() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveTasks++
}

// RecordTaskComplete registra la completación de una tarea
func (m *Metrics) RecordTaskComplete(durationMs int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.ActiveTasks--
	m.CompletedTasks++
	m.TotalTasksExecuted++

	// Actualizar latencia promedio
	if m.CompletedTasks == 1 {
		m.AverageLatencyMs = float64(durationMs)
	} else {
		m.AverageLatencyMs = (m.AverageLatencyMs*float64(m.CompletedTasks-1) + float64(durationMs)) / float64(m.CompletedTasks)
	}
}

// RecordTaskFailed registra una tarea fallida
func (m *Metrics) RecordTaskFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.ActiveTasks--
	m.FailedTasks++
}

// RecordRetry registra un reintento
func (m *Metrics) RecordRetry() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalRetries++
}

// StartJob inicia el seguimiento de un job
func (m *Metrics) StartJob(jobID string, totalTasks int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.JobMetrics[jobID] = &JobMetrics{
		JobID:      jobID,
		StartTime:  time.Now(),
		TasksTotal: totalTasks,
		Stages:     make(map[string]*StageMetrics),
	}
}

// CompleteJob marca un job como completado
func (m *Metrics) CompleteJob(jobID string, recordsProcessed int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if jobMetrics, exists := m.JobMetrics[jobID]; exists {
		now := time.Now()
		jobMetrics.EndTime = &now
		jobMetrics.DurationMs = now.Sub(jobMetrics.StartTime).Milliseconds()
		jobMetrics.RecordsProcessed = recordsProcessed

		// Calcular throughput
		if jobMetrics.DurationMs > 0 {
			jobMetrics.ThroughputRPS = float64(recordsProcessed) / (float64(jobMetrics.DurationMs) / 1000.0)
		}
	}
}

// RecordJobTask registra una tarea de job
func (m *Metrics) RecordJobTask(jobID string, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if jobMetrics, exists := m.JobMetrics[jobID]; exists {
		if success {
			jobMetrics.TasksCompleted++
		} else {
			jobMetrics.TasksFailed++
		}
	}
}

// StartStage inicia el seguimiento de una etapa
func (m *Metrics) StartStage(jobID, stageID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if jobMetrics, exists := m.JobMetrics[jobID]; exists {
		jobMetrics.Stages[stageID] = &StageMetrics{
			StageID:   stageID,
			StartTime: time.Now(),
		}
	}
}

// CompleteStage marca una etapa como completada
func (m *Metrics) CompleteStage(jobID, stageID string, recordsIn, recordsOut int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if jobMetrics, exists := m.JobMetrics[jobID]; exists {
		if stageMetrics, exists := jobMetrics.Stages[stageID]; exists {
			now := time.Now()
			stageMetrics.EndTime = &now
			stageMetrics.DurationMs = now.Sub(stageMetrics.StartTime).Milliseconds()
			stageMetrics.RecordsInput = recordsIn
			stageMetrics.RecordsOutput = recordsOut
		}
	}
}

// GetSnapshot obtiene una snapshot de las métricas
func (m *Metrics) GetSnapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.StartTime).Seconds()

	snapshot := map[string]interface{}{
		"node_id":              m.NodeID,
		"uptime_seconds":       uptime,
		"cpu_usage_percent":    m.CPUUsagePercent,
		"memory_usage_mb":      m.MemoryUsageMB,
		"active_tasks":         m.ActiveTasks,
		"completed_tasks":      m.CompletedTasks,
		"failed_tasks":         m.FailedTasks,
		"average_latency_ms":   m.AverageLatencyMs,
		"total_tasks_executed": m.TotalTasksExecuted,
		"total_retries":        m.TotalRetries,
		"last_update":          m.LastUpdate.Format(time.RFC3339),
	}

	// Agregar métricas de jobs
	jobs := make([]map[string]interface{}, 0)
	for _, jobMetrics := range m.JobMetrics {
		jobSnapshot := map[string]interface{}{
			"job_id":           jobMetrics.JobID,
			"tasks_total":      jobMetrics.TasksTotal,
			"tasks_completed":  jobMetrics.TasksCompleted,
			"tasks_failed":     jobMetrics.TasksFailed,
			"records_processed": jobMetrics.RecordsProcessed,
			"throughput_rps":   jobMetrics.ThroughputRPS,
		}

		if jobMetrics.EndTime != nil {
			jobSnapshot["duration_ms"] = jobMetrics.DurationMs
			jobSnapshot["completed_at"] = jobMetrics.EndTime.Format(time.RFC3339)
		}

		jobs = append(jobs, jobSnapshot)
	}
	snapshot["jobs"] = jobs

	return snapshot
}

// GetJobMetrics obtiene métricas de un job específico
func (m *Metrics) GetJobMetrics(jobID string) *JobMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.JobMetrics[jobID]
}

// Reset reinicia las métricas
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ActiveTasks = 0
	m.CompletedTasks = 0
	m.FailedTasks = 0
	m.AverageLatencyMs = 0
	m.TotalTasksExecuted = 0
	m.TotalBytesProcessed = 0
	m.TotalRetries = 0
	m.JobMetrics = make(map[string]*JobMetrics)
	m.StartTime = time.Now()
}

// GlobalMetrics es la instancia global de métricas
var GlobalMetrics *Metrics

func init() {
	GlobalMetrics = NewMetrics("master")
}