# Mini-Spark-Go

Motor de Procesamiento Distribuido Batch (Ruta A) implementado en Go.

## Descripción

Sistema distribuido tipo Spark que ejecuta jobs definidos como DAGs (Directed Acyclic Graphs) 
sobre datos en archivos, con operadores como map, filter, reduce, join y aggregate.

## Arquitectura

- **Master/Coordinator**: Coordina workers, planifica tareas, gestiona estado
- **Workers**: Ejecutan tareas, manejan particiones, reportan métricas
- **Cliente**: CLI para enviar jobs y consultar estado


## Estado del Proyecto
**Semana 1 - Completada**
- Diseño y scaffolding
- Registro de workers
- Sistema de heartbeats
- API REST básica

**Semana 2 - En progreso**
- Planificación y ejecución básica


## Instalación y Uso

### Con Docker (Recomendado)

#### Levantar entorno de desarrollo**
`docker-compose -f docker-compose.dev.yml run --rm dev`

#### Compilar
`make build`

#### En diferentes terminales:
**Utilizar Docker en terminales**
`docker exec -it 8857fb9e697c /bin/bash`

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