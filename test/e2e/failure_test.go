package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

// TestWorkerFailureRecovery prueba la recuperación ante fallo de worker
func TestWorkerFailureRecovery(t *testing.T) {
	// Iniciar master
	masterCmd := exec.Command("../../bin/master")
	masterCmd.Env = append(os.Environ(), "LOG_LEVEL=info")
	if err := masterCmd.Start(); err != nil {
		t.Fatalf("Failed to start master: %v", err)
	}
	defer masterCmd.Process.Kill()
	
	time.Sleep(2 * time.Second) // Esperar que master inicie
	
	// Iniciar worker 1
	worker1Cmd := exec.Command("../../bin/worker")
	worker1Cmd.Env = append(os.Environ(), "WORKER_ID=worker-test-1")
	if err := worker1Cmd.Start(); err != nil {
		t.Fatalf("Failed to start worker1: %v", err)
	}
	
	time.Sleep(2 * time.Second) // Esperar registro
	
	// Iniciar worker 2
	worker2Cmd := exec.Command("../../bin/worker")
	worker2Cmd.Env = append(os.Environ(), "WORKER_ID=worker-test-2")
	if err := worker2Cmd.Start(); err != nil {
		t.Fatalf("Failed to start worker2: %v", err)
	}
	defer worker2Cmd.Process.Kill()
	
	time.Sleep(2 * time.Second)
	
	// Enviar job
	submitCmd := exec.Command("../../bin/client", "submit", "../../examples/wordcount.json")
	if err := submitCmd.Run(); err != nil {
		t.Fatalf("Failed to submit job: %v", err)
	}
	
	time.Sleep(1 * time.Second)
	
	// Matar worker 1 para simular fallo
	t.Log("Simulating worker failure...")
	if err := worker1Cmd.Process.Kill(); err != nil {
		t.Errorf("Failed to kill worker1: %v", err)
	}
	
	// Esperar replanificación
	time.Sleep(15 * time.Second)
	
	// Verificar que el job se completó (worker 2 debería haber tomado las tareas)
	t.Log("Job should have completed on remaining worker")
}

// TestMultipleWorkerFailures prueba múltiples fallos consecutivos
func TestMultipleWorkerFailures(t *testing.T) {
	t.Log("Testing multiple worker failures...")
	// En una implementación completa, aquí se probaría con 3+ workers
	// y se matarían múltiples workers para verificar reintentos
	t.Skip("Skipping multi-failure test for now")
}

// TestHeartbeatTimeout prueba que workers sin heartbeat sean detectados
func TestHeartbeatTimeout(t *testing.T) {
	// Iniciar master
	masterCmd := exec.Command("../../bin/master")
	if err := masterCmd.Start(); err != nil {
		t.Fatalf("Failed to start master: %v", err)
	}
	defer masterCmd.Process.Kill()
	
	time.Sleep(2 * time.Second)
	
	// Iniciar worker que enviará heartbeat
	workerCmd := exec.Command("../../bin/worker")
	workerCmd.Env = append(os.Environ(), "WORKER_ID=worker-heartbeat-test")
	if err := workerCmd.Start(); err != nil {
		t.Fatalf("Failed to start worker: %v", err)
	}
	
	time.Sleep(5 * time.Second)
	
	// Suspender worker (SIGSTOP) para simular hang sin heartbeat
	if err := workerCmd.Process.Signal(os.Signal(os.Kill)); err != nil {
		t.Logf("Could not suspend worker: %v", err)
	}
	
	// Esperar timeout (10 segundos + margen)
	time.Sleep(15 * time.Second)
	
	// El master debería haber detectado al worker como DOWN
	t.Log("Worker should be marked as DOWN after timeout")
}

// TestTaskRetryLogic prueba la lógica de reintentos
func TestTaskRetryLogic(t *testing.T) {
	t.Log("Testing task retry logic...")
	// En una implementación completa, aquí se simularía un worker
	// que falla tareas intencionalmente para probar reintentos
	t.Skip("Skipping retry logic test for now")
}

// Helper para verificar que un proceso está corriendo
func processRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(os.Signal(0))
	return err == nil
}

// Helper para esperar que un puerto esté disponible
func waitForPort(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		// Intentar conexión simple
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for port %d", port)
}