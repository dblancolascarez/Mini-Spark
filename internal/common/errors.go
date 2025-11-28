// Errores personalizados
package common

import "errors"

// Errores comunes del sistema

var (
	ErrWorkerNotFound     = errors.New("worker not found")
	ErrWorkerAlreadyExists = errors.New("worker already exists")
	ErrNoActiveWorkers    = errors.New("no active workers available")
	ErrJobNotFound        = errors.New("job not found")
	ErrInvalidDAG         = errors.New("invalid DAG structure")
	ErrTaskFailed         = errors.New("task execution failed")
)
