package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dblancolascarez/mini-spark/internal/api"
	"github.com/dblancolascarez/mini-spark/internal/common"
	"github.com/dblancolascarez/mini-spark/internal/master"
)

func main() {
	fmt.Println("=== Mini-Spark Master Starting ===")

	// Cargar configuración
	config := common.DefaultConfig()
	
	// Crear coordinador
	coordinator := master.NewCoordinator(config.Master.HeartbeatTimeout)
	fmt.Printf("[Master] Coordinator initialized (heartbeat timeout: %v)\n", 
		config.Master.HeartbeatTimeout)

	// Crear planificador
	scheduler := master.NewScheduler(coordinator, master.PolicyRoundRobin)
	fmt.Printf("[Master] Scheduler initialized (policy: %s)\n", config.Master.SchedulerPolicy)

	// Crear job manager
	jobManager := master.NewJobManager()
	fmt.Printf("[Master] Job Manager initialized\n")

	// Crear servidor API
	server := api.NewServer(coordinator, scheduler, jobManager, config.Master.Port)

	// Iniciar monitoreo de workers en goroutine
	go monitorWorkers(coordinator, jobManager)

	// Iniciar servidor API en goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("[Master] Failed to start API server: %v", err)
		}
	}()

	fmt.Printf("[Master] API server running on :%d\n", config.Master.Port)
	fmt.Println("[Master] Ready to accept workers and jobs")
	fmt.Println("Press Ctrl+C to shutdown...")

	// Esperar señal de terminación
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n[Master] Shutting down gracefully...")
	fmt.Println("[Master] Shutdown complete")
}

// monitorWorkers verifica periódicamente el estado de los workers
func monitorWorkers(coordinator *master.Coordinator, jobManager *master.JobManager) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		deadWorkers := coordinator.CheckDeadWorkers()
		if len(deadWorkers) > 0 {
			fmt.Printf("[Monitor] Detected %d dead worker(s)\n", len(deadWorkers))
			
			// Replanificar tareas de workers caídos
			for _, workerID := range deadWorkers {
				rescheduled := jobManager.RescheduleTasks(workerID)
				if rescheduled > 0 {
					fmt.Printf("[Monitor] Rescheduled %d tasks from worker %s\n", rescheduled, workerID)
				}
			}
		}

		active, total := coordinator.GetWorkerCount()
		if total > 0 {
			fmt.Printf("[Monitor] Workers: %d active / %d total\n", active, total)
		}
	}
}