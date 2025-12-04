package master

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dblancolascarez/mini-spark/internal/dag"
	"github.com/dblancolascarez/mini-spark/internal/operators"
	"github.com/dblancolascarez/mini-spark/internal/protocol"
)

// JobExecutor ejecuta jobs distribuyendo tareas a workers
type JobExecutor struct {
	coordinator *Coordinator
	scheduler   *Scheduler
	jobManager  *JobManager
	factory     *operators.Factory
}

// NewJobExecutor crea un nuevo executor de jobs
func NewJobExecutor(coordinator *Coordinator, scheduler *Scheduler, jobManager *JobManager) *JobExecutor {
	return &JobExecutor{
		coordinator: coordinator,
		scheduler:   scheduler,
		jobManager:  jobManager,
		factory:     operators.NewFactory(),
	}
}

// ExecuteJob ejecuta un job completo
func (je *JobExecutor) ExecuteJob(job *JobInfo) error {
	fmt.Printf("[JobExecutor] Starting job %s (%s)\n", job.JobID, job.Name)

	// Actualizar estado a RUNNING
	je.jobManager.UpdateJobStatus(job.JobID, JobStatusRunning)

	// Obtener orden de ejecución
	sorted, err := job.DAG.TopologicalSort()
	if err != nil {
		return fmt.Errorf("failed to sort DAG: %w", err)
	}

	// Almacenar outputs intermedios
	outputs := make(map[string][]operators.Record)

	// Ejecutar nodos en orden topológico
	for _, node := range sorted {
		fmt.Printf("[JobExecutor] Executing node %s (%s)\n", node.ID, node.Operator)

		taskID := fmt.Sprintf("%s-task-%s", job.JobID, node.ID)

		// Intentar ejecutar la tarea con reintentos
		const maxRetries = 3
		var output []operators.Record
		var lastErr error

		for attempt := 1; attempt <= maxRetries; attempt++ {
			// Seleccionar worker
			worker, err := je.scheduler.SelectWorker()
			workerID := "local"
			if err == nil && worker != nil {
				workerID = worker.ID
				fmt.Printf("[JobExecutor] Assigned task %s to worker %s (attempt %d/%d)\n",
					taskID, workerID, attempt, maxRetries)
			} else {
				fmt.Printf("[JobExecutor] No workers available, executing locally (attempt %d/%d)\n",
					attempt, maxRetries)
			}

			// Marcar tarea como RUNNING
			je.jobManager.UpdateTaskStatus(taskID, "RUNNING", workerID)

			// Ejecutar tarea
			output, lastErr = je.executeTaskWithRetry(node, outputs, workerID)

			if lastErr == nil {
				// Éxito - salir del loop de reintentos
				je.jobManager.UpdateTaskStatus(taskID, "COMPLETED", workerID)
				fmt.Printf("[JobExecutor] Task %s completed (%d records)\n", taskID, len(output))
				break
			}

			// Falló - marcar como fallido y reintentar si quedan intentos
			fmt.Printf("[JobExecutor] Task %s failed on attempt %d: %v\n", taskID, attempt, lastErr)
			je.jobManager.UpdateTaskStatus(taskID, "FAILED", workerID)

			if attempt == maxRetries {
				// Se agotaron los reintentos
				je.jobManager.UpdateJobStatus(job.JobID, JobStatusFailed)
				return fmt.Errorf("task %s failed after %d attempts: %w", taskID, maxRetries, lastErr)
			}

			fmt.Printf("[JobExecutor] Retrying task %s...\n", taskID)
		}

		// Guardar output
		outputs[node.ID] = output
	}

	fmt.Printf("[JobExecutor] Job %s completed successfully\n", job.JobID)
	return nil
}

// executeTaskWithRetry intenta ejecutar una tarea con soporte para workers remotos
func (je *JobExecutor) executeTaskWithRetry(node *dag.Node, outputs map[string][]operators.Record, workerID string) ([]operators.Record, error) {
	// Si tenemos un worker remoto disponible, intentar ejecución remota
	if workerID != "local" {
		// Obtener input de nodos padres
		var input []operators.Record
		if len(node.Parents) > 0 {
			for _, parent := range node.Parents {
				if parentOutput, exists := outputs[parent.ID]; exists {
					input = append(input, parentOutput...)
				}
			}
		}

		// Convertir input a []map[string]interface{}
		inputData := make([]map[string]interface{}, len(input))
		for i, record := range input {
			inputData[i] = map[string]interface{}(record)
		}

		task := &protocol.TaskAssignment{
			TaskID:     node.ID,
			Operator:   node.Operator,
			Parameters: node.Parameters,
			Input:      inputData,
		}

		// Intentar enviar al worker remoto
		records, err := je.AssignTaskToWorker(workerID, task, input)
		if err != nil {
			fmt.Printf("[JobExecutor] Remote execution failed, falling back to local: %v\n", err)
			// Continuar con ejecución local
		} else {
			// Éxito remoto - devolver datos del worker
			return records, nil
		}
	}

	// Ejecución local (fallback o cuando no hay workers)
	return je.executeTask(node, outputs)
}

// executeTask ejecuta una tarea con sus operadores
func (je *JobExecutor) executeTask(node *dag.Node, outputs map[string][]operators.Record) ([]operators.Record, error) {
	// Crear operador
	op, err := je.factory.CreateOperator(node)
	if err != nil {
		return nil, fmt.Errorf("failed to create operator: %w", err)
	}

	// Obtener input de nodos padres
	var input []operators.Record
	if len(node.Parents) > 0 {
		for _, parent := range node.Parents {
			if parentOutput, exists := outputs[parent.ID]; exists {
				input = append(input, parentOutput...)
			}
		}
	}

	// Ejecutar operador
	output, err := op.Execute(input)
	if err != nil {
		return nil, fmt.Errorf("operator %s failed: %w", node.Operator, err)
	}

	return output, nil
}

// AssignTaskToWorker asigna una tarea a un worker específico
func (je *JobExecutor) AssignTaskToWorker(workerID string, task *protocol.TaskAssignment, input []operators.Record) ([]operators.Record, error) {
	// Obtener información del worker
	worker, err := je.coordinator.GetWorker(workerID)
	if err != nil {
		return nil, fmt.Errorf("worker %s not found: %w", workerID, err)
	}

	// Construir URL del worker
	workerURL := fmt.Sprintf("http://%s:%d/api/v1/tasks/execute", worker.Host, worker.Port)

	// Serializar tarea
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task: %w", err)
	}

	// Enviar tarea al worker via HTTP POST
	resp, err := http.Post(workerURL, "application/json", bytes.NewBuffer(taskJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to send task to worker: %w", err)
	}
	defer resp.Body.Close()

	// Verificar respuesta
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("worker returned status %d", resp.StatusCode)
	}

	// Leer resultado
	var result protocol.TaskResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Verificar si la tarea falló
	if result.Status == "FAILED" {
		return nil, fmt.Errorf("task failed on worker: %s", result.Error)
	}

	fmt.Printf("[JobExecutor] Task %s completed on worker %s (%d records)\n",
		task.TaskID, workerID, result.Records)

	// Convertir Data a []operators.Record
	records := make([]operators.Record, len(result.Data))
	for i, data := range result.Data {
		records[i] = operators.Record(data)
	}

	return records, nil
}
