package protocol

import "time"

// WorkerRegisterRequest - Worker se registra con Master
type WorkerRegisterRequest struct {
	WorkerID   string    `json:"worker_id"`
	Host       string    `json:"host"`
	Port       int       `json:"port"`
	Capacity   int       `json:"capacity"`
	RegisterAt time.Time `json:"register_at"`
}

// WorkerRegisterResponse - Respuesta del Master
type WorkerRegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// HeartbeatRequest - Heartbeat desde Worker
type HeartbeatRequest struct {
	WorkerID      string    `json:"worker_id"`
	Timestamp     time.Time `json:"timestamp"`
	ActiveTasks   int       `json:"active_tasks"`
	CPUUsage      float64   `json:"cpu_usage"`
	MemoryUsageMB int       `json:"memory_usage_mb"`
}

// HeartbeatResponse - Respuesta a heartbeat
type HeartbeatResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

// JobSubmitRequest - Cliente envía job
type JobSubmitRequest struct {
	Name        string        `json:"name"`
	DAG         DAGDefinition `json:"dag"`
	Parallelism int           `json:"parallelism"`
	Partitions  int           `json:"partitions"`
}

// DAGDefinition - Definición del DAG
type DAGDefinition struct {
	Nodes []NodeDefinition `json:"nodes"`
	Edges []EdgeDefinition `json:"edges"`
}

// NodeDefinition - Nodo del DAG
type NodeDefinition struct {
	ID         string                 `json:"id"`
	Operator   string                 `json:"operator"`
	Parameters map[string]interface{} `json:"params,omitempty"`
}

// EdgeDefinition - Arista del DAG
type EdgeDefinition struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// JobSubmitResponse - Respuesta al envío de job
type JobSubmitResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// TaskAssignment - Master asigna tarea a Worker
type TaskAssignment struct {
	TaskID     string                   `json:"task_id"`
	JobID      string                   `json:"job_id"`
	Operator   string                   `json:"operator"`
	Parameters map[string]interface{}   `json:"parameters"`
	Input      []map[string]interface{} `json:"input,omitempty"`
	InputPaths []string                 `json:"input_paths,omitempty"`
	OutputPath string                   `json:"output_path"`
}

// TaskResult - Worker reporta resultado de tarea
type TaskResult struct {
	TaskID     string                   `json:"task_id"`
	JobID      string                   `json:"job_id"`
	WorkerID   string                   `json:"worker_id"`
	Status     string                   `json:"status"`
	Records    int                      `json:"records"`
	Data       []map[string]interface{} `json:"data,omitempty"`
	OutputPath string                   `json:"output_path,omitempty"`
	Error      string                   `json:"error,omitempty"`
}

// JobStatusRequest - Cliente consulta estado de job
type JobStatusRequest struct {
	JobID string `json:"job_id"`
}

// JobStatusResponse - Respuesta con estado de job
type JobStatusResponse struct {
	JobID          string  `json:"job_id"`
	Name           string  `json:"name"`
	Status         string  `json:"status"`
	Progress       float64 `json:"progress"`
	TotalTasks     int     `json:"total_tasks"`
	CompletedTasks int     `json:"completed_tasks"`
	FailedTasks    int     `json:"failed_tasks"`
	StartedAt      string  `json:"started_at,omitempty"`
	CompletedAt    string  `json:"completed_at,omitempty"`
	Error          string  `json:"error,omitempty"`
}
