// Punto de entrada del nodo master

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
	fmt.Println("Mini-Spark Master starting...")

	// Cargar configuración
	config := common.DefaultConfig()

	// Crear coordinador
	coordinator := master.NewCoordinator(config.Master.HeartbeatTimeout)
	fmt.Printf("[Master] Coordinator initialized (heartbeat timeout: %v)\n", 
		config.Master.HeartbeatTimeout)

	
	// Crear planificador
	scheduler := master.NewScheduler(coordinator, master.PolicyRoundRobin)
	fmt.Printf("[Master] Scheduler initialized (policy: %s)\n", config.Master.SchedulerPolicy)

	// Crear servidor API
	server := api.NewServer(coordinator, scheduler, config.Master.Port)

	// Iniciar monitoreo de workers en goroutine
	go monitorWorkers(coordinator)

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
	// TODO: Implementar shutdown graceful (guardar estado, etc.)
	fmt.Println("[Master] Shutdown complete")



	fmt.Println("Master running on :8080")
	fmt.Println("Press Ctrl+C to shutdown...")

	// Esperar señal de terminación
	<-sigChan
	log.Println("Shutting down master...")
}

// monitorWorkers verifica periódicamente el estado de los workers
func monitorWorkers(coordinator *master.Coordinator) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		deadWorkers := coordinator.CheckDeadWorkers()
		if len(deadWorkers) > 0 {
			fmt.Printf("[Monitor] Detected %d dead worker(s)\n", len(deadWorkers))
			// TODO: Replanificar tareas de workers muertos
		}

		active, total := coordinator.GetWorkerCount()
		if active > 0 {
			fmt.Printf("[Monitor] Workers: %d active / %d total\n", active, total)
		}
	}
}
