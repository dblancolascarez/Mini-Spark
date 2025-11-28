// Coordinación y registro de workers
package master

import (
	"fmt"
	"sync"
	"time"

	"github.com/dblancolascarez/mini-spark/internal/protocol"
)

// WorkerStatus representa el estado de un worker
type WorkerStatus string

const (
	WorkerStatusActive WorkerStatus = "ACTIVE"
	WorkerStatusDown   WorkerStatus = "DOWN"
)

// WorkerInfo contiene información de un worker registrado
type WorkerInfo struct {
	ID            string
	Host          string
	Port          int
	Capacity      int
	Status        WorkerStatus
	ActiveTasks   int
	LastHeartbeat time.Time
	RegisteredAt  time.Time
	CPUUsage      float64
	MemoryUsageMB int
}

// Coordinator maneja el registro y estado de workers
type Coordinator struct {
	workers          map[string]*WorkerInfo
	mu               sync.RWMutex
	heartbeatTimeout time.Duration
}

// NewCoordinator crea un nuevo coordinador
func NewCoordinator(heartbeatTimeout time.Duration) *Coordinator {
	return &Coordinator{
		workers:          make(map[string]*WorkerInfo),
		heartbeatTimeout: heartbeatTimeout,
	}
}

// RegisterWorker registra un nuevo worker
func (c *Coordinator) RegisterWorker(req protocol.WorkerRegisterRequest) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Verificar si ya existe
	if existing, exists := c.workers[req.WorkerID]; exists {
		if existing.Status == WorkerStatusActive {
			return fmt.Errorf("worker %s already registered", req.WorkerID)
		}
		// Si estaba DOWN, reactivar
		existing.Status = WorkerStatusActive
		existing.LastHeartbeat = time.Now()
		fmt.Printf("[Coordinator] Worker %s reactivated\n", req.WorkerID)
		return nil
	}

	// Registrar nuevo worker
	worker := &WorkerInfo{
		ID:            req.WorkerID,
		Host:          req.Host,
		Port:          req.Port,
		Capacity:      req.Capacity,
		Status:        WorkerStatusActive,
		ActiveTasks:   0,
		LastHeartbeat: time.Now(),
		RegisteredAt:  req.RegisterAt,
	}

	c.workers[req.WorkerID] = worker
	fmt.Printf("[Coordinator] Worker registered: %s (%s:%d)\n", req.WorkerID, req.Host, req.Port)
	return nil
}

// UpdateHeartbeat actualiza el heartbeat de un worker
func (c *Coordinator) UpdateHeartbeat(req protocol.HeartbeatRequest) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	worker, exists := c.workers[req.WorkerID]
	if !exists {
		return fmt.Errorf("worker %s not registered", req.WorkerID)
	}

	worker.LastHeartbeat = req.Timestamp
	worker.ActiveTasks = req.ActiveTasks
	worker.CPUUsage = req.CPUUsage
	worker.MemoryUsageMB = req.MemoryUsageMB
	worker.Status = WorkerStatusActive

	return nil
}

// GetActiveWorkers retorna lista de workers activos
func (c *Coordinator) GetActiveWorkers() []*WorkerInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var active []*WorkerInfo
	for _, worker := range c.workers {
		if worker.Status == WorkerStatusActive {
			active = append(active, worker)
		}
	}
	return active
}

// GetWorker retorna información de un worker específico
func (c *Coordinator) GetWorker(workerID string) (*WorkerInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	worker, exists := c.workers[workerID]
	if !exists {
		return nil, fmt.Errorf("worker %s not found", workerID)
	}
	return worker, nil
}

// CheckDeadWorkers verifica workers que no han enviado heartbeat
func (c *Coordinator) CheckDeadWorkers() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	var deadWorkers []string
	now := time.Now()

	for id, worker := range c.workers {
		if worker.Status == WorkerStatusActive {
			timeSinceHeartbeat := now.Sub(worker.LastHeartbeat)
			if timeSinceHeartbeat > c.heartbeatTimeout {
				worker.Status = WorkerStatusDown
				deadWorkers = append(deadWorkers, id)
				fmt.Printf("[Coordinator] Worker %s marked as DOWN (no heartbeat for %v)\n", 
					id, timeSinceHeartbeat)
			}
		}
	}

	return deadWorkers
}

// GetWorkerCount retorna el número total de workers
func (c *Coordinator) GetWorkerCount() (active int, total int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total = len(c.workers)
	for _, worker := range c.workers {
		if worker.Status == WorkerStatusActive {
			active++
		}
	}
	return active, total
}

// GetWorkerStats retorna estadísticas de todos los workers
func (c *Coordinator) GetWorkerStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_workers"] = len(c.workers)
	
	activeCount := 0
	totalTasks := 0
	totalCPU := 0.0
	totalMemory := 0

	for _, worker := range c.workers {
		if worker.Status == WorkerStatusActive {
			activeCount++
			totalTasks += worker.ActiveTasks
			totalCPU += worker.CPUUsage
			totalMemory += worker.MemoryUsageMB
		}
	}

	stats["active_workers"] = activeCount
	stats["total_active_tasks"] = totalTasks
	stats["avg_cpu_usage"] = 0.0
	stats["total_memory_mb"] = totalMemory

	if activeCount > 0 {
		stats["avg_cpu_usage"] = totalCPU / float64(activeCount)
	}

	return stats
}
 