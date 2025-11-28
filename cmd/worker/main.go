 // Punto de entrada del nodo worker
 package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"runtime"
	"time"
	
	"github.com/dblancolascarez/mini-spark/internal/common"
	"github.com/dblancolascarez/mini-spark/internal/worker"
)

func main() {
	fmt.Println("Mini-Spark Worker Starting...")

	// Cargar configuración
	config := common.DefaultConfig()
	
	// Obtener ID del worker (de variable de entorno o generar)
	workerID := os.Getenv("WORKER_ID")
	if workerID == "" {
		workerID = fmt.Sprintf("worker-%d", time.Now().Unix())
	}

	// Obtener dirección del master (de variable de entorno o usar default)
	masterAddress := os.Getenv("MASTER_ADDRESS")
	if masterAddress == "" {
		masterAddress = config.Worker.MasterAddress
	}

	fmt.Printf("[Worker] ID: %s\n", workerID)
	fmt.Printf("[Worker] Master: %s\n", masterAddress)

	// Crear cliente para comunicarse con master
	client := worker.NewClient(workerID, masterAddress, "localhost", config.Worker.Port)

	// Registrarse con el master
	fmt.Println("[Worker] Registering with master...")
	if err := client.Register(config.Worker.TaskPoolSize); err != nil {
		log.Fatalf("[Worker] Failed to register: %v", err)
	}
	fmt.Println("[Worker] Successfully registered with master")

	// Iniciar envío de heartbeats
	go sendHeartbeats(client, config.Worker.HeartbeatInterval)

	fmt.Println("[Worker] Worker is running")
	fmt.Println("Press Ctrl+C to shutdown...")

	// Esperar señal de terminación
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n[Worker] Shutting down...")
	fmt.Println("[Worker] Shutdown complete")
}

// sendHeartbeats envía heartbeats periódicos al master
func sendHeartbeats(client *worker.Client, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		// Obtener métricas básicas
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		memoryMB := int(m.Alloc / 1024 / 1024)

		// Enviar heartbeat
		if err := client.SendHeartbeat(0, 0.0, memoryMB); err != nil {
			fmt.Printf("[Worker] Failed to send heartbeat: %v\n", err)
		} else {
			fmt.Printf("[Worker] Heartbeat sent (memory: %dMB)\n", memoryMB)
		}
	}
}
