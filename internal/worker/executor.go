package worker

import (
	"fmt"

	"github.com/dblancolascarez/mini-spark/internal/dag"
	"github.com/dblancolascarez/mini-spark/internal/operators"
	"github.com/dblancolascarez/mini-spark/internal/protocol"
)

// Executor ejecuta tareas en el worker
type Executor struct {
	workerID string
	factory  *operators.Factory
}

// NewExecutor crea un nuevo executor
func NewExecutor(workerID string) *Executor {
	return &Executor{
		workerID: workerID,
		factory:  operators.NewFactory(),
	}
}

// ExecuteTask ejecuta una tarea específica
func (e *Executor) ExecuteTask(task *protocol.TaskAssignment) (*protocol.TaskResult, error) {
	fmt.Printf("[Executor:%s] Executing task %s (operator: %s)\n",
		e.workerID, task.TaskID, task.Operator)

	// Crear el operador
	node := &dag.Node{
		ID:         task.TaskID,
		Operator:   task.Operator,
		Parameters: task.Parameters,
	}

	op, err := e.factory.CreateOperator(node)
	if err != nil {
		return nil, fmt.Errorf("failed to create operator: %w", err)
	}

	// Ejecutar operador
	// Convertir Input de la tarea a []operators.Record
	input := make([]operators.Record, len(task.Input))
	for i, data := range task.Input {
		input[i] = operators.Record(data)
	}

	output, err := op.Execute(input)
	if err != nil {
		return &protocol.TaskResult{
			TaskID:  task.TaskID,
			Status:  "FAILED",
			Error:   err.Error(),
			Records: 0,
		}, err
	}

	fmt.Printf("[Executor:%s] Task %s completed (%d records)\n",
		e.workerID, task.TaskID, len(output))

	// Convertir output a []map[string]interface{}
	data := make([]map[string]interface{}, len(output))
	for i, record := range output {
		data[i] = map[string]interface{}(record)
	}

	return &protocol.TaskResult{
		TaskID:     task.TaskID,
		Status:     "COMPLETED",
		Records:    len(output),
		Data:       data,
		OutputPath: task.OutputPath,
	}, nil
}

// ExecuteDAG ejecuta un DAG completo (para testing local)
func (e *Executor) ExecuteDAG(dag *dag.DAG) error {
	// Obtener orden topológico
	sorted, err := dag.TopologicalSort()
	if err != nil {
		return fmt.Errorf("failed to sort DAG: %w", err)
	}

	// Mapa para almacenar outputs intermedios
	outputs := make(map[string][]operators.Record)

	// Ejecutar cada nodo en orden
	for _, node := range sorted {
		fmt.Printf("[Executor] Executing node: %s (%s)\n", node.ID, node.Operator)

		// Crear operador
		op, err := e.factory.CreateOperator(node)
		if err != nil {
			return fmt.Errorf("failed to create operator for %s: %w", node.ID, err)
		}

		// Obtener input de nodos padres
		var input []operators.Record
		if len(node.Parents) > 0 {
			// Combinar outputs de todos los padres
			for _, parent := range node.Parents {
				if parentOutput, exists := outputs[parent.ID]; exists {
					input = append(input, parentOutput...)
				}
			}
		}

		// Ejecutar operador
		output, err := op.Execute(input)
		if err != nil {
			return fmt.Errorf("operator %s failed: %w", node.ID, err)
		}

		// Guardar output
		outputs[node.ID] = output
		fmt.Printf("[Executor] Node %s produced %d records\n", node.ID, len(output))
	}

	fmt.Println("[Executor] DAG execution completed successfully")
	return nil
}
