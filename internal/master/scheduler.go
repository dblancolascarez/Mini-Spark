// Planificador de tareas (round-robin + carga)
package master

import (
	"fmt"
	"sync"
)

// SchedulerPolicy define la política de planificación
type SchedulerPolicy string

const (
	PolicyRoundRobin  SchedulerPolicy = "round-robin"
	PolicyLeastLoaded SchedulerPolicy = "least-loaded"
)

// Scheduler asigna tareas a workers
type Scheduler struct {
	coordinator *Coordinator
	policy      SchedulerPolicy
	mu          sync.Mutex
	roundRobinIndex int
}

// NewScheduler crea un nuevo planificador
func NewScheduler(coordinator *Coordinator, policy SchedulerPolicy) *Scheduler {
	return &Scheduler{
		coordinator:     coordinator,
		policy:          policy,
		roundRobinIndex: 0,
	}
}

// SelectWorker selecciona un worker según la política configurada
func (s *Scheduler) SelectWorker() (*WorkerInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	workers := s.coordinator.GetActiveWorkers()
	if len(workers) == 0 {
		return nil, fmt.Errorf("no active workers available")
	}

	switch s.policy {
	case PolicyRoundRobin:
		return s.selectRoundRobin(workers), nil
	case PolicyLeastLoaded:
		return s.selectLeastLoaded(workers), nil
	default:
		return s.selectRoundRobin(workers), nil
	}
}

// selectRoundRobin selecciona worker en modo round-robin
func (s *Scheduler) selectRoundRobin(workers []*WorkerInfo) *WorkerInfo {
	if len(workers) == 0 {
		return nil
	}

	selected := workers[s.roundRobinIndex%len(workers)]
	s.roundRobinIndex++
	
	fmt.Printf("[Scheduler] Selected worker %s (round-robin)\n", selected.ID)
	return selected
}

// selectLeastLoaded selecciona el worker con menos carga
func (s *Scheduler) selectLeastLoaded(workers []*WorkerInfo) *WorkerInfo {
	if len(workers) == 0 {
		return nil
	}

	var selected *WorkerInfo
	minLoad := int(^uint(0) >> 1) // Max int

	for _, worker := range workers {
		load := worker.ActiveTasks
		if load < minLoad {
			minLoad = load
			selected = worker
		}
	}

	fmt.Printf("[Scheduler] Selected worker %s (least-loaded, tasks=%d)\n", 
		selected.ID, selected.ActiveTasks)
	return selected
}
