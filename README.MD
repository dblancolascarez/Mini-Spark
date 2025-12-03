# Mini-Spark-Go

Motor de Procesamiento Distribuido Batch (Ruta A) implementado en Go.

## Descripción

Sistema distribuido tipo Spark que ejecuta jobs definidos como DAGs (Directed Acyclic Graphs) 
sobre datos en archivos, con operadores como map, filter, reduce, join y aggregate.

**✅ Proyecto Completo - Semanas 1-4 Implementadas**

## Características Implementadas

### ✅ Semana 1-2: Arquitectura Base
- Master/Coordinator con registro y heartbeats de workers
- Planificador (Round-robin, Least-loaded)
- API REST completa para jobs y workers
- Cliente CLI para envío de jobs
- Operadores básicos: map, filter, flat_map, reader, writer

### ✅ Semana 3: Operadores Avanzados y Tolerancia a Fallos
- **Operadores avanzados**: join, aggregate (sum/avg/max/min), shuffle, reduce completo
- **Ejecución distribuida**: Workers con servidor HTTP para recibir tareas
- **Tolerancia a fallos**: 
  - Sistema de reintentos (3 intentos por tarea)
  - Detección de workers caídos
  - Replanificación automática de tareas

### ✅ Semana 4: Observabilidad y Pruebas
- **Logging estructurado**: Niveles DEBUG/INFO/WARN/ERROR, formato JSON/texto
- **Sistema de métricas**: CPU, memoria, latencia, throughput por nodo y job
- **Tests E2E**: Pruebas de fallo y benchmarks
- **Demo completo**: Script automatizado con casos de uso

## Arquitectura

```
┌─────────────────────────────────────────────────────────┐
│                    MASTER NODE                          │
│  ┌──────────┐  ┌───────────┐  ┌──────────────┐        │
│  │   API    │  │Coordinator│  │  Scheduler   │        │
│  │  Server  │  │           │  │              │        │
│  └────┬─────┘  └─────┬─────┘  └──────┬───────┘        │
│       │              │                │                 │
│       └──────────────┴────────────────┘                 │
│                      │                                   │
└──────────────────────┼───────────────────────────────────┘
                       │ HTTP/JSON
         ┌─────────────┼─────────────┐
         │             │             │
    ┌────▼────┐   ┌────▼────┐   ┌───▼─────┐
    │ Worker1 │   │ Worker2 │   │ Worker3 │
    │         │   │         │   │         │
    │ Executor│   │ Executor│   │ Executor│
    └─────────┘   └─────────┘   └─────────┘
```

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
# 1. Compilar binarios
make build

# O manualmente:
go build -o bin/master ./cmd/master
go build -o bin/worker ./cmd/worker
go build -o bin/client ./cmd/client

# 2. Terminal 1 - Iniciar Master
./bin/master

# 3. Terminal 2 - Iniciar Worker 1
export WORKER_ID=worker-1
./bin/worker
```
**Terminal 3: Worker 2**
``` bash
export WORKER_ID=worker-2
./bin/worker

# 5. Terminal 4 - Enviar job
./bin/client submit examples/wordcount.json
./bin/client list

# 6. Ejecutar demo completo
./scripts/demo.sh
```

### Opción 2: Con Docker

```bash
# Levantar entorno
docker-compose -f docker-compose.dev.yml run --rm dev

# Dentro del contenedor
make build
./bin/master
```

## API Endpoints

### Sistema
- `GET /health` - Estado del sistema y workers activos
- `GET /api/v1/workers` - Lista de workers con estadísticas
- `POST /api/v1/workers/register` - Registrar nuevo worker
- `POST /api/v1/workers/heartbeat` - Heartbeat de worker

### Jobs
- `POST /api/v1/jobs` - Enviar nuevo job
- `GET /api/v1/jobs` - Listar todos los jobs
- `GET /api/v1/jobs/{id}` - Estado de un job específico
- `GET /api/v1/jobs/{id}/results` - Resultados de un job

### Métricas
- `GET /api/v1/metrics` - Métricas del sistema (CPU, memoria, throughput)

## Operadores Disponibles

| Operador | Descripción | Parámetros |
|----------|-------------|------------|
| `read_csv` | Lee datos de archivos CSV | `path`, `partitions` |
| `read_jsonl` | Lee datos de archivos JSONL | `path`, `partitions` |
| `map` | Transforma cada registro | `fn` (función) |
| `filter` | Filtra registros | `fn` (condición) |
| `flat_map` | Map que retorna múltiples registros | `fn` |
| `reduce_by_key` | Reduce por clave | `key`, `fn` (count/sum/avg/max/min) |
| `join` | Join entre datasets | `type` (inner/left/right), `left_key`, `right_key` |
| `aggregate` | Agregaciones | `fn` (count/sum/avg/max/min/group_by) |
| `shuffle` | Redistribuye particiones | `partitions`, `key` |
| `write_csv` | Escribe resultados a CSV | `path` |

## Ejemplos de Jobs

### WordCount
```json
{
  "name": "wordcount",
  "dag": {
    "nodes": [
      {"id": "read", "operator": "read_csv", "params": {"path": "data/input/words.csv"}},
      {"id": "flat", "operator": "flat_map", "params": {"fn": "tokenize"}},
      {"id": "reduce", "operator": "reduce_by_key", "params": {"key": "word", "fn": "count"}},
      {"id": "write", "operator": "write_csv", "params": {"path": "data/output/result.csv"}}
    ],
    "edges": [["read","flat"], ["flat","reduce"], ["reduce","write"]]
  }
}
```

### Aggregate con Estadísticas
```json
{
  "name": "sales_stats",
  "dag": {
    "nodes": [
      {"id": "read", "operator": "read_csv", "params": {"path": "data/input/sales.csv"}},
      {"id": "agg", "operator": "aggregate", "params": {"fn": "sum", "field": "amount"}},
      {"id": "write", "operator": "write_csv", "params": {"path": "data/output/stats.csv"}}
    ],
    "edges": [["read","agg"], ["agg","write"]]
  }
}
```

## Comandos Útiles

```bash
# Ver estado del cluster
curl -s http://localhost:8080/health | jq

# Ver workers activos
curl -s http://localhost:8080/api/v1/workers | jq

# Ver todos los jobs
curl -s http://localhost:8080/api/v1/jobs | jq

# Ver estado de un job específico
curl -s http://localhost:8080/api/v1/jobs/$JOB_ID | jq

# Ver métricas del sistema
curl -s http://localhost:8080/api/v1/metrics | jq

# Enviar múltiples jobs
for i in {1..5}; do
  ./bin/client submit examples/wordcount.json
  sleep 0.5
done

# Ejecutar tests
go test ./test/e2e/... -v
go test ./test/integration/... -v
go test ./test/unit/... -v

# Ejecutar demo completo
./scripts/demo.sh
```

## Tolerancia a Fallos

El sistema implementa varios mecanismos de tolerancia a fallos:

1. **Heartbeat Monitoring**: Workers envían pulsos cada 3 segundos
2. **Detección de Fallos**: Master marca workers como DOWN si no recibe heartbeat en 10s
3. **Reintentos**: Hasta 3 intentos por tarea fallida
4. **Replanificación**: Tareas de workers caídos se reasignan automáticamente
5. **Fallback Local**: Si no hay workers disponibles, ejecución local en master

### Simular Fallo de Worker

```bash
# Terminal 1: Master corriendo
./bin/master

# Terminal 2: Worker 1
export WORKER_ID=worker-1
./bin/worker

# Terminal 3: Worker 2  
export WORKER_ID=worker-2
./bin/worker

# Terminal 4: Enviar job
./bin/client submit examples/wordcount.json

# Terminal 2: Matar Worker 1 (Ctrl+C)
# El master detectará el fallo y replanificará las tareas al Worker 2
```

## Logging y Métricas

### Configurar Logging
```go
import "github.com/dblancolascarez/mini-spark/internal/monitoring"

// Crear logger
logger := monitoring.NewLogger("my-component", monitoring.INFO, "json")

// Usar logger
logger.Info("Task started", map[string]interface{}{"task_id": "task-123"})
logger.Error("Task failed", err, map[string]interface{}{"task_id": "task-123"})
```

### Métricas Disponibles
- CPU usage (%)
- Memoria usage (MB)
- Tareas activas/completadas/fallidas
- Latencia promedio (ms)
- Throughput (records/sec)
- Reintentos totales

## Tests

```bash
# Tests unitarios
go test ./test/unit/... -v

# Tests de integración
go test ./test/integration/... -v

# Tests E2E (requiere cluster corriendo)
go test ./test/e2e/... -v

# Benchmarks
go test ./test/e2e/benchmark_test.go -v -timeout 30m

# Todos los tests
make test
```

## Estructura del Proyecto

```
Mini-Spark/
├── cmd/                    # Puntos de entrada
│   ├── master/            # Master node
│   ├── worker/            # Worker node
│   └── client/            # CLI cliente
├── internal/              # Código interno
│   ├── api/               # API REST
│   ├── master/            # Lógica del master
│   ├── worker/            # Lógica del worker
│   ├── operators/         # Operadores DAG
│   ├── dag/               # Parser y validador DAG
│   ├── protocol/          # Mensajes y tipos
│   ├── monitoring/        # Logging y métricas
│   └── storage/           # Sistema de archivos
├── test/                  # Tests
│   ├── unit/             # Tests unitarios
│   ├── integration/      # Tests de integración
│   └── e2e/              # Tests end-to-end
├── examples/              # Jobs de ejemplo
├── scripts/               # Scripts de utilidad
│   ├── demo.sh           # Demo completo
│   ├── test-cluster.sh   # Tests del cluster
│   └── run-cluster.sh    # Iniciar cluster
└── data/                  # Datos de entrada/salida
    ├── input/
    └── output/
```

## Desarrollo

```bash
# Compilar
make build

# Limpiar
make clean

# Ejecutar master en desarrollo
make run-master

# Ejecutar worker en desarrollo
make run-worker

# Ver ayuda
make help
```

## Contribución

Este es un proyecto académico para el curso de Sistemas Operativos.

## Autores

- Equipo Mini-Spark
- Curso: Principios de Sistemas Operativos
- Profesor: Kenneth Obando Rodríguez

## Licencia

Proyecto académico - TEC 2025