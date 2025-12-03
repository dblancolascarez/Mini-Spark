# Mini-Spark - Proyecto Completo

## 🎯 Resumen Ejecutivo

**Motor de Procesamiento Distribuido Batch implementado desde cero en Go**

- **Lenguaje**: Go 1.21+
- **Arquitectura**: Master-Worker distribuida
- **Comunicación**: HTTP/JSON (stateless)
- **Tolerancia a Fallos**: Reintentos, detección de caídas, replanificación
- **Estado**: ✅ **COMPLETO** - Semanas 1-4 implementadas

---

## 📋 Funcionalidades Implementadas

### ✅ Semana 1-2: Arquitectura Base y Coordinación
- [x] Master/Coordinator con registro de workers
- [x] Sistema de heartbeats (3s intervalo, 10s timeout)
- [x] Planificador con políticas: Round-robin, Least-loaded
- [x] API REST completa (8 endpoints)
- [x] Cliente CLI para envío de jobs
- [x] DAG parser y validator
- [x] Operadores básicos: map, filter, flat_map, reader, writer
- [x] Sistema de archivos (CSV, JSONL)

### ✅ Semana 3: Operadores Avanzados y Tolerancia a Fallos
- [x] **Operadores avanzados implementados:**
  - `join` - Inner, left, right joins entre datasets
  - `aggregate` - count, sum, avg, max, min, group_by
  - `shuffle` - Redistribución por hash de particiones
  - `reduce_by_key` - count, sum, avg, max, min, collect

- [x] **Ejecución distribuida de tareas:**
  - Servidor HTTP en workers (`/api/v1/tasks/execute`)
  - `AssignTaskToWorker` para envío remoto de tareas
  - Fallback automático a ejecución local

- [x] **Tolerancia a fallos:**
  - Sistema de reintentos (máximo 3 intentos)
  - Detección de workers DOWN via heartbeat
  - Replanificación automática de tareas
  - Contador de reintentos por tarea

### ✅ Semana 4: Observabilidad, Pruebas y Documentación
- [x] **Sistema de logging estructurado:**
  - Niveles: DEBUG, INFO, WARN, ERROR, FATAL
  - Formatos: JSON y texto legible
  - Contexto y campos adicionales
  - Thread-safe con mutex

- [x] **Sistema de métricas completo:**
  - Métricas por nodo: CPU, memoria, tareas, latencia
  - Métricas por job: duración, throughput, registros
  - Métricas por stage: input/output, reintentos
  - API de snapshots en tiempo real

- [x] **Tests end-to-end:**
  - `failure_test.go` - Pruebas de recuperación ante fallos
  - `benchmark_test.go` - Benchmarks con 1M+ registros
  - Tests de escalabilidad (1, 2, 4 workers)
  - Tests de heartbeat timeout

- [x] **Documentación y demo:**
  - `demo.sh` - Script completo automatizado
  - README.md actualizado con guías completas
  - Ejemplos de jobs (wordcount, agregaciones)
  - Instrucciones de instalación y uso

---

## 🏗️ Arquitectura Técnica

### Componentes Principales

```
┌─────────────────────────────────────────────────────────┐
│                  MASTER NODE (Puerto 8080)              │
│                                                          │
│  ┌──────────────┐  ┌─────────────┐  ┌───────────────┐ │
│  │  API Server  │  │ Coordinator │  │   Scheduler   │ │
│  │  (HTTP/JSON) │  │ (Registry)  │  │  (RR/LL)      │ │
│  └──────┬───────┘  └──────┬──────┘  └───────┬───────┘ │
│         │                  │                  │         │
│         └──────────────────┴──────────────────┘         │
│                            │                            │
│  ┌─────────────────────────┴──────────────────────┐    │
│  │         Job Manager & Executor                 │    │
│  │  - DAG execution   - Task assignment           │    │
│  │  - Retry logic     - Failure recovery          │    │
│  └────────────────────────────────────────────────┘    │
└─────────────────────────┬────────────────────────────────┘
                          │
                          │ HTTP/JSON
         ┌────────────────┼────────────────┐
         │                │                │
    ┌────▼─────┐    ┌─────▼────┐    ┌────▼─────┐
    │ Worker 1 │    │ Worker 2 │    │ Worker N │
    │(8081)    │    │(8082)    │    │(808N)    │
    │          │    │          │    │          │
    │ Server   │    │ Server   │    │ Server   │
    │ Executor │    │ Executor │    │ Executor │
    │ Client   │    │ Client   │    │ Client   │
    └──────────┘    └──────────┘    └──────────┘
```

### Flujo de Ejecución de un Job

1. **Cliente envía job** → `POST /api/v1/jobs`
2. **Master valida DAG** → Parser + Validator
3. **Job Manager crea JobInfo** → Estado: ACCEPTED
4. **Job Executor ordena nodos** → Topological sort
5. **Para cada nodo del DAG:**
   - Scheduler selecciona worker disponible
   - Se intenta enviar tarea al worker remoto
   - Si falla, se reintenta hasta 3 veces
   - Si worker está DOWN, se replanifica
6. **Worker ejecuta tarea** → Crea operador y ejecuta
7. **Worker retorna resultado** → TaskResult con records
8. **Job Manager actualiza estado** → COMPLETED o FAILED

### Tolerancia a Fallos - Flujo

```
Worker envía heartbeat cada 3s
         │
         ▼
Master verifica timeout (10s)
         │
         ├─► Worker OK → Continúa
         │
         └─► Worker DOWN
                 │
                 ├─► Marca worker como DOWN
                 │
                 ├─► Obtiene tareas RUNNING del worker
                 │
                 ├─► Marca tareas como PENDING
                 │
                 └─► Scheduler reasigna a otros workers
```

---

## 📊 Métricas y Observabilidad

### Métricas Disponibles

**Por Nodo:**
- CPU usage (%)
- Memory usage (MB)
- Active tasks
- Completed/Failed tasks
- Average latency (ms)
- Total retries

**Por Job:**
- Duration (ms)
- Tasks total/completed/failed
- Records processed
- Throughput (records/sec)

**Por Stage:**
- Input/Output records
- Duration
- Retries count

### Endpoints de Métricas

```bash
# Snapshot general
curl http://localhost:8080/api/v1/metrics | jq

# Métricas de workers
curl http://localhost:8080/api/v1/workers | jq '.stats'

# Estado de salud
curl http://localhost:8080/health | jq
```

---

## 🧪 Tests y Benchmarks

### Tests Implementados

**Unitarios** (`test/unit/`):
- `dag_test.go` - Validación de DAGs
- `operators_test.go` - Pruebas de operadores
- `scheduler_test.go` - Algoritmos de planificación

**Integración** (`test/integration/`):
- `single_node_test.go` - Ejecución en un nodo
- `operator_chain_test.go` - Cadenas de operadores

**End-to-End** (`test/e2e/`):
- `failure_test.go` - Recuperación ante fallos
- `benchmark_test.go` - Benchmarks de rendimiento
- `multinode_test.go` - Cluster multi-nodo

### Ejecutar Tests

```bash
# Tests unitarios
go test ./test/unit/... -v

# Tests de integración
go test ./test/integration/... -v

# Tests E2E (requiere cluster)
go test ./test/e2e/... -v

# Benchmarks completos
go test ./test/e2e/benchmark_test.go -v -timeout 30m
```

### Resultados Esperados de Benchmarks

- **Throughput**: > 10,000 records/sec con 2 workers
- **Latencia**: < 100ms promedio por tarea
- **Escalabilidad**: ~1.8x mejora con 2x workers
- **Memory**: < 512MB por worker con 1M records

---

## 🚀 Guía de Uso Rápido

### 1. Compilar Proyecto

```bash
# Con Make
make build

# O manualmente
go build -o bin/master ./cmd/master
go build -o bin/worker ./cmd/worker
go build -o bin/client ./cmd/client
```

### 2. Iniciar Cluster (4 terminales)

```bash
# Terminal 1: Master
./bin/master

# Terminal 2: Worker 1
export WORKER_ID=worker-1
./bin/worker

# Terminal 3: Worker 2
export WORKER_ID=worker-2
./bin/worker

# Terminal 4: Cliente
./bin/client submit examples/wordcount.json
./bin/client list
```

### 3. Ejecutar Demo Completo

```bash
./scripts/demo.sh
```

El demo muestra:
1. ✅ Inicio del cluster
2. ✅ WordCount en CSV
3. ✅ Operadores avanzados (filter+map)
4. ✅ Simulación de fallo de worker
5. ✅ Recuperación automática
6. ✅ Estadísticas finales

### 4. Probar Tolerancia a Fallos

```bash
# Terminal 1: Master
./bin/master

# Terminales 2-3: Workers
export WORKER_ID=worker-1 && ./bin/worker
export WORKER_ID=worker-2 && ./bin/worker

# Terminal 4: Enviar job
./bin/client submit examples/wordcount.json

# Terminal 2: Matar Worker 1 (Ctrl+C)
# Observar en Terminal 1 (Master):
# - Detección de worker DOWN
# - Replanificación de tareas
# - Completación en Worker 2
```

---

## 📁 Estructura del Código

```
Mini-Spark/
├── cmd/
│   ├── master/main.go          # Entrypoint Master
│   ├── worker/main.go          # Entrypoint Worker
│   └── client/main.go          # CLI Cliente
│
├── internal/
│   ├── api/                    # API REST
│   │   ├── server.go          # HTTP Server
│   │   ├── handlers.go        # Request handlers
│   │   └── types.go           # DTOs
│   │
│   ├── master/                # Lógica Master
│   │   ├── coordinator.go     # Registry de workers
│   │   ├── scheduler.go       # Planificación
│   │   ├── job_manager.go     # Gestión de jobs
│   │   ├── job_executor.go    # Ejecución de jobs
│   │   └── heartbeat.go       # Monitoreo
│   │
│   ├── worker/                # Lógica Worker
│   │   ├── server.go          # HTTP Server
│   │   ├── client.go          # Cliente HTTP
│   │   ├── executor.go        # Ejecutor de tareas
│   │   └── task_pool.go       # Pool de workers
│   │
│   ├── operators/             # Operadores DAG
│   │   ├── operator.go        # Interface
│   │   ├── map.go
│   │   ├── filter.go
│   │   ├── reduce.go
│   │   ├── join.go
│   │   ├── aggregate.go
│   │   ├── shuffle.go
│   │   ├── reader.go
│   │   └── writer.go
│   │
│   ├── monitoring/            # Observabilidad
│   │   ├── logger.go          # Logging estructurado
│   │   └── metrics.go         # Sistema de métricas
│   │
│   ├── dag/                   # DAG Processing
│   │   ├── graph.go
│   │   ├── parser.go
│   │   └── validator.go
│   │
│   ├── protocol/              # Protocolos
│   │   ├── messages.go        # Tipos de mensajes
│   │   └── serialization.go  # Serialización
│   │
│   └── storage/               # Almacenamiento
│       ├── fs.go              # Sistema de archivos
│       └── partition.go       # Particionamiento
│
├── test/
│   ├── unit/                  # Tests unitarios
│   ├── integration/           # Tests integración
│   └── e2e/                   # Tests end-to-end
│
├── examples/                  # Jobs de ejemplo
│   ├── wordcount.json
│   └── filter_map.json
│
├── scripts/                   # Scripts utilidad
│   ├── demo.sh               # Demo completo ⭐
│   ├── test-cluster.sh       # Tests cluster
│   └── run-cluster.sh        # Iniciar cluster
│
└── data/
    ├── input/                # Datos entrada
    └── output/               # Resultados
```

---

## 🎓 Conceptos de SO Aplicados

| Concepto SO | Implementación en Mini-Spark |
|-------------|------------------------------|
| **Procesos/Hilos** | Goroutines para heartbeats, task pools, API server |
| **IPC/Redes** | HTTP/JSON entre master-workers, sockets TCP |
| **Planificación** | Scheduler con Round-robin y Least-loaded |
| **Memoria** | Cache de particiones, spill to disk cuando excede |
| **Sistema Archivos** | Lectura/escritura CSV/JSONL, particionamiento |
| **Coordinación Dist.** | Heartbeats, registro de workers, consensus simple |
| **Tolerancia Fallos** | Reintentos, detección DOWN, replanificación |
| **Sincronización** | Mutex (sync.RWMutex) para estructuras compartidas |

---

## 📝 Entregables Completados

- [x] **Código fuente completo** en repositorio
- [x] **README.md** con instrucciones detalladas
- [x] **Makefile y scripts** de build/ejecución
- [x] **Documento de arquitectura** (docs/architecture.md)
- [x] **Suite de pruebas** (unit/integration/e2e)
- [x] **Script demo.sh** con casos de uso
- [ ] **Video demostrativo** (pendiente - grabar siguiendo demo.sh)
- [x] **Reporte de benchmarks** (incluido en tests)

---

## 🎬 Guión para Video Demostrativo

### Parte 1: Introducción (1 min)
- Mostrar estructura del proyecto
- Explicar arquitectura master-worker
- Mencionar funcionalidades implementadas

### Parte 2: Instalación y Build (30s)
```bash
make build
ls bin/
```

### Parte 3: Demo en Vivo (3 min)
```bash
# Ejecutar demo.sh y explicar cada paso
./scripts/demo.sh
```

**Mostrar:**
1. Inicio del cluster
2. Registro de workers
3. Envío de job WordCount
4. Estado en tiempo real
5. Matar un worker
6. Observar recuperación
7. Estadísticas finales

### Parte 4: Revisión de Código (2 min)
- Abrir `internal/operators/join.go` - Operador join
- Abrir `internal/master/job_executor.go` - Sistema de reintentos
- Abrir `internal/monitoring/metrics.go` - Sistema de métricas

### Parte 5: Tests y Benchmarks (1 min)
```bash
go test ./test/e2e/failure_test.go -v
```

### Parte 6: Conclusión (30s)
- Resumen de funcionalidades
- Conceptos de SO aplicados
- Trabajo futuro (opcional)

**Duración total: ~8 minutos**

---

## 🔧 Trabajo Futuro (Opcional - Mejoras)

Si se quisiera extender el proyecto:

1. **Persistencia real**: SQLite/PostgreSQL en lugar de in-memory
2. **Checkpointing**: Snapshots periódicos del estado
3. **Compresión**: Comprimir datos intermedios
4. **Métricas exportadas**: Prometheus/Grafana integration
5. **Web UI**: Dashboard para monitoreo visual
6. **Autoscaling**: Agregar/remover workers dinámicamente
7. **Security**: Autenticación y encriptación
8. **Streaming**: Ventanas temporales para procesamiento continuo

---

## 📞 Soporte

Para preguntas sobre el proyecto:
- Ver documentación en `docs/architecture.md`
- Revisar ejemplos en `examples/`
- Ejecutar `./scripts/demo.sh` para guía interactiva

---

**Proyecto académico - Principios de Sistemas Operativos**  
**Tecnológico de Costa Rica - 2025**

✅ **PROYECTO COMPLETO Y FUNCIONAL**
