package master

import (
	"fmt"

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
		
		// Seleccionar worker
		worker, err := je.scheduler.SelectWorker()
		workerID := "local"
		if err == nil {
			workerID = worker.ID
			fmt.Printf("[JobExecutor] Assigned task %s to worker %s\n", taskID, workerID)
		} else {
			fmt.Printf("[JobExecutor] No workers available, executing locally\n")
		}

		// Marcar tarea como RUNNING
		je.jobManager.UpdateTaskStatus(taskID, "RUNNING", workerID)

		// Ejecutar tarea
		output, err := je.executeTask(node, outputs)
		if err != nil {
			je.jobManager.UpdateTaskStatus(taskID, "FAILED", workerID)
			je.jobManager.UpdateJobStatus(job.JobID, JobStatusFailed)
			return fmt.Errorf("task %s failed: %w", taskID, err)
		}

		// Guardar output
		outputs[node.ID] = output

		// Marcar tarea como completada
		je.jobManager.UpdateTaskStatus(taskID, "COMPLETED", workerID)
		fmt.Printf("[JobExecutor] Task %s completed (%d records)\n", taskID, len(output))
	}

	fmt.Printf("[JobExecutor] Job %s completed successfully\n", job.JobID)
	return nil
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

// AssignTaskToWorker asigna una tarea a un worker específico (para Semana 3)
func (je *JobExecutor) AssignTaskToWorker(workerID string, task *protocol.TaskAssignment) error {
	// TODO: Implementar en Semana 3
	return fmt.Errorf("remote task execution not yet implemented")
}