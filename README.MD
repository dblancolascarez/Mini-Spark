# Mini-Spark-Go

Motor de Procesamiento Distribuido Batch (Ruta A) implementado en Go.

## Descripción

Sistema distribuido tipo Spark que ejecuta jobs definidos como DAGs (Directed Acyclic Graphs) 
sobre datos en archivos, con operadores como map, filter, reduce, join y aggregate.

## Arquitectura

- **Master/Coordinator**: Coordina workers, planifica tareas, gestiona estado
- **Workers**: Ejecutan tareas, manejan particiones, reportan métricas
- **Cliente**: CLI para enviar jobs y consultar estado


## Instalación y Uso

### Con Docker (Recomendado)

#### Levantar entorno de desarrollo**
`docker-compose -f docker-compose.dev.yml run --rm dev`

#### Compilar
`make build`

#### En diferentes terminales:
**Utilizar Docker en terminales**
`docker exec -it <CONTAINER_ID> /bin/bash`

**Importante: instalar curl**
`apk add curl jq`

**Terminal 1: Master**
`./bin/master`

**Terminal 2: Worker 1**
```bash
export WORKER_ID=worker-1
./bin/worker
```
**Terminal 3: Worker 2**
``` bash
export WORKER_ID=worker-2
./bin/worker
```

**Terminal 4: Pruebas**
``` bash
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/workers
```

### Pruebas

#### Ejecutar suite de pruebas
`./scripts/test-cluster.sh`

## API Endpoints

`GET /health` - Estado del sistema
`GET /api/v1/workers` - Lista workers activos
`POST /api/v1/workers/register` - Registrar worker
`POST /api/v1/workers/heartbeat` - Heartbeat de worker

## Desarrollo

``` bash
make build # Compilar
make clean # Limpiar binarios
make run-master # Ejecutar master
make run-worker # Ejecutar worker
```

### Otros

**Inicializar master**
`./bin/master`

**Inicializar worker**
``` bash
cd /app
export WORKER_ID=worker-1
./bin/worker
```

**Enviar job**
`./bin/client submit examples/wordcount.json`

**Ver job status**
`curl -s http://localhost:8080/api/v1/jobs/$JOB_ID | jq`

**Ver todos los jobs**
`./bin/client list`

**Ver estadisticas de jobs**
`curl -s http://localhost:8080/api/v1/workers | jq '.stats'`

**Ejemplo: Enviar 5 jobs consecutivos**
``` bash
for i in {1..5}; do
  echo "Submitting job $i..."
  ./bin/client submit examples/filter_map.json
  sleep 0.5
done
```

## TODO
En Internal/master/job_executor.go
``` go
// AssignTaskToWorker asigna una tarea a un worker específico (para Semana 3)
func (je *JobExecutor) AssignTaskToWorker(workerID string, task *protocol.TaskAssignment) error {
	// TODO: Implementar en Semana 3
	return fmt.Errorf("remote task execution not yet implemented")
}
```
En Internal/operators/agregate.go
``` go
// AssignTaskToWorker asigna una tarea a un worker específico (para Semana 3)
func (je *JobExecutor) AssignTaskToWorker(workerID string, task *protocol.TaskAssignment) error {
	// TODO: Implementar en Semana 3
	return fmt.Errorf("remote task execution not yet implemented")
}
```

En Internal/operator/shuffle.go
``` go
// Execute ejecuta el shuffle
func (s *ShuffleOperator) Execute(input []Record) ([]Record, error) {
	// TODO: Implementar en Semana 3
	return nil, fmt.Errorf("shuffle operator not yet implemented")
}
```