# Arquitectura Mini-Spark - Semana 1

## Componentes Implementados

### 1. Master/Coordinator
- Puerto: 8080
- Responsabilidades:
  - Registro de workers
  - Monitoreo de heartbeats
  - Detección de fallos (timeout: 10s)
  - API REST para gestión

### 2. Workers
- Responsabilidades:
  - Auto-registro con el master al iniciar
  - Envío de heartbeats cada 3 segundos
  - Reporte de métricas (memoria)

### 3. Comunicación
- Protocolo: HTTP/JSON
- Endpoints implementados:
  - `POST /api/v1/workers/register`
  - `POST /api/v1/workers/heartbeat`
  - `GET /api/v1/workers`
  - `GET /health`

## Flujo de Operación

1. Master inicia y escucha en :8080
2. Worker inicia y se registra con master
3. Worker envía heartbeats cada 3s
4. Master monitorea cada 5s y marca DOWN si no hay heartbeat por >10s

## Tolerancia a Fallos

- Detección automática de workers caídos
- Estado persistente de workers (ACTIVE/DOWN)
- Preparado para replanificación de tareas (Semana 3)

## Próximos Pasos (Semana 2)

- Implementar envío y parseo de DAGs
- Planificador de tareas
- Ejecución de operadores básicos (map, filter)
- Sistema de métricas de progreso