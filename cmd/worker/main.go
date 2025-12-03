 // Punto de entrada del nodo worker
 package main

import (
	"fmt"
	"log"
	"net"
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

	// Asignar puerto dinámicamente
	// Si se especifica WORKER_PORT, usar ese; sino buscar puerto disponible
	workerPort := 0 // 0 = el sistema asigna automáticamente
	if portEnv := os.Getenv("WORKER_PORT"); portEnv != "" {
		fmt.Sscanf(portEnv, "%d", &workerPort)
	}
	
	if workerPort == 0 {
		// Buscar puerto disponible automáticamente
		workerPort = findAvailablePort(8081, 8100)
	}

	fmt.Printf("[Worker] ID: %s\n", workerID)
	fmt.Printf("[Worker] Master: %s\n", masterAddress)
	fmt.Printf("[Worker] Port: %d\n", workerPort)

	// Crear servidor HTTP para recibir tareas
	server := worker.NewServer(workerID, workerPort)
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("[Worker] Server failed: %v", err)
		}
	}()

	// Crear cliente para comunicarse con master
	client := worker.NewClient(workerID, masterAddress, "localhost", workerPort)

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

// findAvailablePort busca un puerto disponible en el rango especificado
func findAvailablePort(start, end int) int {
	for port := start; port <= end; port++ {
		if isPortAvailable(port) {
			return port
		}
	}
	// Si no encuentra puerto disponible, retornar el default
	return 8081
}

// isPortAvailable verifica si un puerto está disponible
func isPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}
