// Definición de mensajes IPC
package protocol

import "time"

// Mensajes para comunicación entre componentes

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
	Name        string          `json:"name"`
	DAG         DAGDefinition   `json:"dag"`
	Parallelism int             `json:"parallelism"`
	Partitions  int             `json:"partitions"`
}

// DAGDefinition - Definición del DAG
type DAGDefinition struct {
	Nodes []NodeDefinition `json:"nodes"`
	Edges []EdgeDefinition `json:"edges"`
}

// NodeDefinition - Nodo del DAG
type NodeDefinition struct {
	ID         string                 `json:"id"`
	Operator   string                 `json:"op"`
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
