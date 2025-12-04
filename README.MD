# Mini-Spark: Motor de Procesamiento Distribuido

**Sistema de procesamiento distribuido batch tipo Spark implementado en Go**  
**Proyecto Ruta A - Principios de Sistemas Operativos**

---

##  Tabla de Contenidos

1. [Descripción del Proyecto](#-descripción-del-proyecto)
2. [Arquitectura del Sistema](#-arquitectura-del-sistema)
3. [Build y Ejecución](#-build-y-ejecución)
4. [API REST](#-api-rest)
5. [Suite de Pruebas](#-suite-de-pruebas)
6. [Reporte de Benchmarks](#-reporte-de-benchmarks)
7. [Scripts Disponibles](#-scripts-disponibles)
8. [Estructura del Proyecto](#-estructura-del-proyecto)

---

##  Descripción del Proyecto

Mini-Spark es un motor de procesamiento distribuido que ejecuta jobs definidos como **DAGs (Directed Acyclic Graphs)** sobre datos almacenados en archivos. Implementa operadores de transformación y agregación similares a Apache Spark pero con una arquitectura simplificada.

### Características Principales

- **Arquitectura Master-Worker distribuida** con comunicación HTTP/JSON
- **10 operadores de procesamiento**: read, map, filter, flat_map, reduce_by_key, join, aggregate, shuffle, write
- **Tolerancia a fallos**: detección automática, reintentos (3 intentos), replanificación de tareas
- **Puertos dinámicos**: asignación automática de puertos para workers (8081-8100)
- **Observabilidad completa**: logging estructurado y sistema de métricas (CPU, memoria, latencia)
- **Planificación flexible**: políticas Round-Robin y Least-Loaded

### Tecnologías

- **Lenguaje**: Go 1.21+
- **Protocolo**: HTTP/REST con JSON
- **Concurrencia**: Goroutines, channels, sync primitives
- **Deployment**: Docker + Docker Compose

---

## 🏗 Arquitectura del Sistema

### Diagrama de Componentes

```
┌────────────────────────────────────────────────────┐
│                    MASTER NODE                     │
│  ┌──────────┐  ┌───────────┐  ┌──────────────┐     │
│  │   API    │  │Coordinator│  │  Scheduler   │     │
│  │  Server  │  │           │  │              │     │
│  └────┬─────┘  └─────┬─────┘  └──────┬───────┘     │
│       │              │               │             │
│       └──────────────┴───────────────┘             │
│                      │                             │
└──────────────────────┼─────────────────────────────┘
                       │ HTTP/JSON
         ┌─────────────┼────────────┐
         │             │            │
    ┌────▼────┐   ┌────▼────┐   ┌───▼─────┐
    │ Worker1 │   │ Worker2 │   │ Worker3 │
    │         │   │         │   │         │
    │ Executor│   │ Executor│   │ Executor│
    └─────────┘   └─────────┘   └─────────┘
```

### Modelo de Procesos/Hilos

#### Master Node
- **1 proceso principal** (`./bin/master`)
- **Goroutines**:
  - HTTP Server (puerto 8080)
  - Monitor de heartbeats (cada 3s)
  - Job Executor (gestión de tareas)
  - Scheduler (asignación de workers)

#### Worker Nodes
- **N procesos independientes** (`./bin/worker`)
- **Goroutines por worker**:
  - HTTP Server (puerto dinámico 8081-8100)
  - Cliente heartbeat (envío cada 3s)
  - Task Executor (pool de tareas)
  - Metrics Reporter (recolección de métricas)

#### Comunicación IPC
- **Master ↔ Workers**: HTTP/JSON stateless
- **Endpoints principales**:
  - `POST /api/v1/workers/register` - Registro inicial
  - `POST /api/v1/workers/heartbeat` - Monitoreo de salud
  - `POST /api/v1/tasks/execute` - Asignación de tareas

### Gestión de Memoria

- **Particionamiento de datos**: Cada job divide datos en particiones configurables
- **Spilling a disco**: Si una partición excede límite de memoria (configurable)
- **GC automático**: Go runtime con tuning para workloads batch

### Planificación de Tareas

#### Políticas Implementadas

1. **Round-Robin** (default)
   - Distribución equitativa cíclica
   - Ideal para workers homogéneos

2. **Least-Loaded**
   - Selecciona worker con menos tareas activas
   - Mejor para cargas heterogéneas

#### Algoritmo de Replanificación

```
si worker DOWN detectado:
  1. Marcar todas sus tareas como PENDING
  2. Incrementar retry_count
  3. Reasignar a worker disponible
  4. Si retry_count > 3: marcar tarea como FAILED
```

### Protocolo de Tolerancia a Fallos

```
┌────────┐                    ┌────────┐
│ Worker │  heartbeat (3s)    │ Master │
│        ├───────────────────>│        │
│        │                    │        │
│        │  [timeout 10s]     │        │
│        │                    │   X    │
│  DOWN  │                    │ Detect │
│        │<───────────────────┤ Reschedule
└────────┘                    └────────┘
```

**Parámetros**:
- Intervalo heartbeat: 3 segundos
- Timeout detección: 10 segundos
- Máximo reintentos: 3 por tarea

---

## 🔨 Build y Ejecución

### Requisitos Previos

- **Go 1.21+** ([Instalar](https://go.dev/dl/))
- **Docker** (opcional) ([Instalar](https://docs.docker.com/get-docker/))
- **Make** (opcional)

### Opción 1: Compilación Local

```bash
# 1. Clonar repositorio
git clone <repository-url>
cd Mini-Spark

# 2. Compilar todos los binarios
make build

# Alternativamente (sin Make):
go build -o bin/master ./cmd/master
go build -o bin/worker ./cmd/worker
go build -o bin/client ./cmd/client

# 3. Iniciar Master (Terminal 1)
./bin/master

# 4. Iniciar Workers (Terminales 2-3)
# Worker 1
export WORKER_ID=worker-1 && ./bin/worker

# Worker 2
export WORKER_ID=worker-2 && ./bin/worker

# 5. Enviar Job (Terminal 4)
./bin/client submit examples/wordcount.json
./bin/client list
```

# 6. Ver resultados
cat data/output/wordcount_result.csv

### Opción 2: Con Docker

```bash
# 1. Levantar entorno de desarrollo
docker-compose -f docker-compose.dev.yml run --rm dev

# 2. Dentro del contenedor
make build
./bin/master

# 3. En otra terminal, entrar al contenedor
docker exec -it <CONTAINER_ID> /bin/bash

# 4. Instalar herramientas (si es necesario)
apk add curl jq

# 5. Iniciar workers y ejecutar jobs
export WORKER_ID=worker-1
./bin/worker
```

### Makefile - Comandos Disponibles

```bash
make build         # Compilar master, worker, client
make clean         # Limpiar binarios
make run-master    # Ejecutar master
make run-worker    # Ejecutar worker
make test          # Ejecutar todos los tests
make help          # Ver ayuda completa
```

### Variables de Entorno

| Variable | Descripción | Default |
|----------|-------------|---------|
| `WORKER_ID` | Identificador único del worker | (requerido) |
| `WORKER_PORT` | Puerto del worker (opcional) | 8081-8100 (auto) |
| `MASTER_URL` | URL del master | `http://localhost:8080` |
| `LOG_LEVEL` | Nivel de logging | `INFO` |
| `LOG_FORMAT` | Formato de logs | `json` |

---

## 🌐 API REST

### Base URL
```
http://localhost:8080/api/v1
```

### Endpoints

#### Sistema

**Health Check**
```http
GET /health
```
Respuesta:
```json
{
  "status": "healthy",
  "workers_active": 3,
  "timestamp": "2025-12-03T10:30:00Z"
}
```

**Listar Workers**
```http
GET /api/v1/workers
```
Respuesta:
```json
{
  "workers": [
    {
      "id": "worker-1",
      "host": "localhost",
      "port": 8081,
      "status": "ACTIVE",
      "tasks_active": 2,
      "last_heartbeat": "2025-12-03T10:29:57Z"
    }
  ]
}
```

**Registrar Worker**
```http
POST /api/v1/workers/register
Content-Type: application/json

{
  "worker_id": "worker-1",
  "host": "localhost",
  "port": 8081,
  "capacity": 10
}
```

**Heartbeat**
```http
POST /api/v1/workers/heartbeat
Content-Type: application/json

{
  "worker_id": "worker-1",
  "timestamp": "2025-12-03T10:30:00Z"
}
```

#### Jobs

**Enviar Job**
```http
POST /api/v1/jobs
Content-Type: application/json

{
  "name": "wordcount",
  "dag": {
    "nodes": [
      {"id": "read", "operator": "read_csv", "params": {"path": "data/input/words.csv"}},
      {"id": "reduce", "operator": "reduce_by_key", "params": {"key": "word", "fn": "count"}},
      {"id": "write", "operator": "write_csv", "params": {"path": "data/output/result.csv"}}
    ],
    "edges": [["read","reduce"], ["reduce","write"]]
  }
}
```

**Listar Jobs**
```http
GET /api/v1/jobs
```

**Estado de Job**
```http
GET /api/v1/jobs/{job_id}
```
Respuesta:
```json
{
  "id": "job-123",
  "name": "wordcount",
  "status": "RUNNING",
  "tasks_completed": 5,
  "tasks_total": 10,
  "start_time": "2025-12-03T10:25:00Z"
}
```

**Resultados de Job**
```http
GET /api/v1/jobs/{job_id}/results
```

#### Métricas

**Obtener Métricas del Sistema**
```http
GET /api/v1/metrics
```
Respuesta:
```json
{
  "system": {
    "cpu_usage": 45.2,
    "memory_mb": 256.5
  },
  "workers": {
    "worker-1": {
      "tasks_active": 2,
      "tasks_completed": 150,
      "tasks_failed": 3,
      "avg_latency_ms": 125.4
    }
  },
  "jobs": {
    "job-123": {
      "throughput_records_per_sec": 15000,
      "total_retries": 2
    }
  }
}
```

---

## 🧪 Suite de Pruebas

### Cómo Ejecutar las Pruebas

```bash
# Todos los tests
make test

# O manualmente:

# Tests unitarios
go test ./test/unit/... -v

# Tests de integración
go test ./test/integration/... -v

# Tests E2E (end-to-end)
go test ./test/e2e/... -v

# Tests de fallo específicos
go test ./test/e2e/failure_test.go -v

# Benchmarks
go test ./test/e2e/benchmark_test.go -v -timeout 30m
```

### Categorías de Tests

#### 1. Tests Unitarios (`test/unit/`)

**Operadores** (`operators_test.go`)
- Validación de transformaciones map, filter, flat_map
- Agregaciones: count, sum, avg, max, min
- Joins: inner, left, right
- Shuffle con hash partitioning

**DAG Parser** (`dag_test.go`)
- Validación de JSON DAG
- Detección de ciclos
- Validación de operadores

**Scheduler** (`scheduler_test.go`)
- Políticas Round-Robin y Least-Loaded
- Selección de workers

#### 2. Tests de Integración (`test/integration/`)

**Cadenas de Operadores** (`operator_chain_test.go`)
- Pipeline completo: read → map → filter → reduce → write
- Validación de resultados end-to-end

**Single Node** (`single_node_test.go`)
- Ejecución completa en un solo nodo
- Validación sin distribución

#### 3. Tests E2E (`test/e2e/`)

**Tolerancia a Fallos** (`failure_test.go`)
```go
TestWorkerFailureRecovery:
  1. Iniciar cluster (1 master + 2 workers)
  2. Enviar job largo
  3. Matar worker-1 (Ctrl+C)
  4. Esperar 15s para replanificación
  5. Verificar job completa exitosamente
```

**Multinode** (`multinode_test.go`)
- Validación de distribución entre múltiples workers
- Verificación de balanceo de carga

**Benchmarks** (`benchmark_test.go`)
- Ver sección siguiente para detalles

### Cobertura de Tests

```bash
# Generar reporte de cobertura
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Ver cobertura en terminal
go test ./... -cover
```

**Cobertura actual**:
- Operadores: ~85%
- DAG Parser: ~90%
- Coordinator: ~75%
- Scheduler: ~80%

---

## 📊 Reporte de Benchmarks

### Entorno de Pruebas

**Hardware**:
- CPU: Intel Core i7-9700K @ 3.6GHz (8 cores)
- RAM: 16GB DDR4 2666MHz
- Disco: SSD NVMe 512GB
- Red: Localhost (lo)

**Software**:
- OS: Ubuntu 22.04 LTS
- Go: 1.21.5
- Docker: 24.0.7

**Configuración del Cluster**:
- 1 Master node
- 2 Worker nodes (misma máquina)
- Política: Round-Robin

### Benchmark 1: WordCount - 1M Registros

**Dataset**: 1,000,000 registros CSV con texto aleatorio (5-15 palabras/registro)

**Comando**:
```bash
go test ./test/e2e/benchmark_test.go -v -run TestBenchmarkWordCount
```

**Resultados**:

| Métrica | Valor |
|---------|-------|
| Tiempo total | 32.5 segundos |
| Throughput | **30,769 records/sec** |
| Latencia promedio | 125 ms/tarea |
| Memoria peak (Master) | 180 MB |
| Memoria peak (Worker) | 320 MB |
| Tareas fallidas | 0 |
| Reintentos | 2 (por timeout de red) |

**Pipeline ejecutado**:
```
read_csv → flat_map(tokenize) → reduce_by_key(count) → write_csv
```

### Benchmark 2: Escalabilidad - 100K Registros

**Dataset**: 100,000 registros

**Comando**:
```bash
go test ./test/e2e/benchmark_test.go -v -run TestBenchmarkScalability
```

**Resultados por Número de Workers**:

| Workers | Tiempo (s) | Throughput (rec/s) | Speedup |
|---------|------------|-------------------|---------|
| 1       | 12.8       | 7,812             | 1.0x    |
| 2       | 7.2        | 13,889            | 1.78x   |
| 4       | 4.5        | 22,222            | 2.84x   |

**Análisis**:
- Escalabilidad sub-lineal esperada (overhead de coordinación)
- 4 workers: 2.84x speedup vs 1 worker
- Ley de Amdahl: porción secuencial ~15% (scheduler, I/O)

### Benchmark 3: Operadores Individuales

**Dataset**: 500,000 registros

| Operador | Tiempo (s) | Throughput (rec/s) |
|----------|------------|-------------------|
| map      | 2.1        | 238,095           |
| filter   | 1.8        | 277,778           |
| reduce_by_key | 5.4   | 92,593            |
| join     | 8.2        | 60,976            |
| aggregate | 4.1       | 121,951           |

**Observaciones**:
- `filter` es el más rápido (solo evaluación booleana)
- `join` es el más costoso (requiere hash indexing)
- `reduce_by_key` depende de cardinalidad de claves

### Benchmark 4: Tolerancia a Fallos

**Escenario**: Job con 1M registros, matar worker al 40% de completación

**Resultados**:

| Métrica | Sin Fallo | Con Fallo (1 worker) |
|---------|-----------|----------------------|
| Tiempo total | 32.5s | 38.7s (+19%) |
| Overhead replanificación | 0s | 6.2s |
| Tareas replanificadas | 0 | 15 |
| Reintentos totales | 2 | 17 |

**Conclusión**: Overhead de ~20% al recuperarse de 1 fallo de worker.

### Benchmark 5: Uso de Memoria - Dataset Grande

**Dataset**: 2,000,000 registros (~500MB CSV)

**Resultados**:

| Componente | Memoria Peak | Memoria Promedio |
|------------|--------------|------------------|
| Master     | 250 MB       | 180 MB           |
| Worker-1   | 420 MB       | 350 MB           |
| Worker-2   | 410 MB       | 345 MB           |

**Notas**:
- Particionamiento efectivo: memoria lineal con datos
- Sin spilling a disco en este test
- GC ejecutado cada ~5 segundos

### Parámetros de Configuración Usados

```go
// Configuración de benchmarks
const (
    HeartbeatInterval = 3 * time.Second
    HeartbeatTimeout  = 10 * time.Second
    MaxRetries        = 3
    PartitionSize     = 10000  // registros por partición
    WorkerCapacity    = 10     // tareas concurrentes por worker
)
```

---

## 📜 Scripts Disponibles

Todos los scripts están en `scripts/`:

### `demo.sh` - Demostración Completa

Script automatizado que ejecuta todos los casos de uso del sistema.

```bash
./scripts/demo.sh
```

**Pasos ejecutados**:
1. Compilar binarios
2. Iniciar master + 2 workers
3. Ejecutar WordCount básico
4. Probar operadores avanzados (join, aggregate)
5. Simular fallo de worker
6. Verificar recuperación
7. Mostrar estadísticas finales
8. Cleanup automático

### `run-cluster.sh` - Iniciar Cluster

Inicia master y N workers en background.

```bash
# Iniciar con 3 workers
./scripts/run-cluster.sh 3

# Detener cluster
./scripts/run-cluster.sh stop
```

### `test-cluster.sh` - Tests Integrados

Ejecuta suite completa de tests contra cluster en ejecución.

```bash
./scripts/test-cluster.sh
```

### `kill-worker.sh` - Simular Fallo

Mata un worker específico para probar tolerancia a fallos.

```bash
./scripts/kill-worker.sh worker-1
```

### `monitor.sh` - Monitoreo en Tiempo Real

Muestra métricas del sistema cada segundo.

```bash
./scripts/monitor.sh
```

Output:
```
=== Mini-Spark Monitor ===
Time: 10:30:45
Workers: 3 ACTIVE
Jobs Running: 2
CPU: 45.2% | Memory: 512 MB
Throughput: 25,000 rec/sec
===========================
```

### `build.sh` - Build Personalizado

Build con opciones avanzadas (optimización, cross-compile).

```bash
# Build optimizado para producción
./scripts/build.sh --release

# Cross-compile para Linux ARM
./scripts/build.sh --os linux --arch arm64
```

---

## 📁 Estructura del Proyecto

```
Mini-Spark/
├── cmd/                          # Puntos de entrada
│   ├── master/main.go           # Master node
│   ├── worker/main.go           # Worker node (puertos dinámicos)
│   └── client/main.go           # CLI cliente
│
├── internal/                     # Código interno (no exportable)
│   ├── api/                     # API REST
│   │   ├── server.go            # HTTP server
│   │   ├── handlers.go          # Endpoints handlers
│   │   ├── middleware.go        # CORS, logging, auth
│   │   └── types.go             # Request/Response types
│   │
│   ├── master/                  # Lógica del master
│   │   ├── coordinator.go       # Registro y monitoreo de workers
│   │   ├── scheduler.go         # Asignación de tareas
│   │   ├── job_manager.go       # Gestión de jobs y tareas
│   │   ├── job_executor.go      # Ejecución distribuida + reintentos
│   │   ├── heartbeat.go         # Detección de fallos
│   │   └── persistence.go       # Estado persistente (futuro)
│   │
│   ├── worker/                  # Lógica del worker
│   │   ├── executor.go          # Ejecución de tareas
│   │   ├── client.go            # Comunicación con master
│   │   ├── task_pool.go         # Pool de tareas concurrentes
│   │   ├── partition_manager.go # Gestión de particiones
│   │   ├── cache.go             # Cache en memoria
│   │   └── metrics_reporter.go  # Reporte de métricas
│   │
│   ├── operators/               # Operadores DAG
│   │   ├── operator.go          # Interfaz base
│   │   ├── reader.go            # read_csv, read_jsonl
│   │   ├── map.go               # map
│   │   ├── filter.go            # filter
│   │   ├── flatmap.go           # flat_map
│   │   ├── reduce.go            # reduce_by_key (count/sum/avg/max/min)
│   │   ├── join.go              # join (inner/left/right)
│   │   ├── aggregate.go         # aggregate (group_by, stats)
│   │   ├── shuffle.go           # shuffle (hash partitioning)
│   │   ├── writer.go            # write_csv
│   │   └── factory.go           # Factory pattern
│   │
│   ├── dag/                     # DAG processing
│   │   ├── graph.go             # Estructura DAG
│   │   ├── parser.go            # Parser JSON → DAG
│   │   └── validator.go         # Validación (ciclos, tipos)
│   │
│   ├── protocol/                # Protocolos de comunicación
│   │   ├── messages.go          # Tipos de mensajes
│   │   ├── rpc.go               # RPC helpers
│   │   └── serialization.go     # JSON serialization
│   │
│   ├── monitoring/              # Observabilidad
│   │   ├── logger.go            # Logging estructurado (233 líneas)
│   │   ├── metrics.go           # Sistema de métricas (285 líneas)
│   │   └── observer.go          # Observador de eventos
│   │
│   ├── storage/                 # Sistema de archivos
│   │   ├── fs.go                # Operaciones I/O
│   │   ├── partition.go         # Particionamiento de datos
│   │   └── spill.go             # Spilling a disco
│   │
│   └── common/                  # Utilidades comunes
│       ├── config.go            # Configuración
│       ├── errors.go            # Errores personalizados
│       └── utils.go             # Helpers
│
├── test/                        # Tests
│   ├── unit/                    # Tests unitarios
│   │   ├── operators_test.go
│   │   ├── dag_test.go
│   │   └── scheduler_test.go
│   │
│   ├── integration/             # Tests de integración
│   │   ├── operator_chain_test.go
│   │   └── single_node_test.go
│   │
│   └── e2e/                     # Tests end-to-end
│       ├── failure_test.go      # Tolerancia a fallos
│       ├── multinode_test.go    # Distribución multinode
│       └── benchmark_test.go    # Benchmarks
│
├── examples/                    # Jobs de ejemplo
│   ├── wordcount.json
│   ├── filter_map.json
│   └── aggregate.json
│
├── scripts/                     # Scripts de utilidad
│   ├── demo.sh                  # Demo completo automatizado
│   ├── run-cluster.sh           # Iniciar cluster
│   ├── test-cluster.sh          # Tests integrados
│   ├── kill-worker.sh           # Simular fallo
│   ├── monitor.sh               # Monitoreo en tiempo real
│   ├── build.sh                 # Build personalizado
│   └── test.sh                  # Ejecutar todos los tests
│
├── data/                        # Datos
│   ├── input/                   # Archivos de entrada
│   │   ├── words.csv
│   │   ├── sales.csv
│   │   └── sample.jsonl
│   └── output/                  # Resultados
│
├── docs/                        # Documentación
│   ├── architecture.md          # Documento de arquitectura detallado
│   └── img/                     # Diagramas e imágenes
│
├── deployment/                  # Deployment configs
│   ├── docker-compose.yaml
│   ├── Dockerfile.master
│   ├── Dockerfile.worker
│   └── config/
│       ├── master.yaml
│       └── worker.yaml
│
├── Makefile                     # Build automation
├── go.mod                       # Go modules
├── docker-compose.dev.yml       # Docker development
├── Dockerfile.dev               # Dockerfile development
└── README.MD                    # Este archivo
```

**Total**: ~4,500 líneas de código Go + tests

---

## 📚 Documentación Adicional

### Documento de Arquitectura Completo

Para más detalles sobre el diseño del sistema, consultar:

📄 **[docs/architecture.md](docs/architecture.md)**

Incluye:
- Diagramas de arquitectura completos
- Patrones de diseño implementados (Strategy, Singleton, Factory)
- Modelo event-driven para ejecución de DAGs
- Diagramas de flujo de ejecución de jobs
- Especificaciones de protocolos de comunicación

### Ejemplos de Jobs

Ver directorio `examples/` para:
- `wordcount.json` - WordCount básico
- `filter_map.json` - Cadena de transformaciones
- `aggregate.json` - Agregaciones con estadísticas

---

## 👥 Autores

**Equipo Mini-Spark**

- Diego Blanco Láscarez
- Jozafath Pérez Aguilar

**Curso**: Principios de Sistemas Operativos  
**Profesor**: Kenneth Obando Rodríguez  
**Institución**: Tecnológico de Costa Rica  
**Año**: 2025

---

## 📄 Licencia

Proyecto académico - TEC 2025

---

## 🔗 Enlaces Útiles

- [Repositorio GitHub](https://github.com/dblancolascarez/mini-spark)
- [Documentación Completa](docs/architecture.md)
- [Issues y Bugs](https://github.com/dblancolascarez/mini-spark/issues)