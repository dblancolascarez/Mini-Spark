#!/bin/bash
# Script para simular fallos en un worker

if [ -z "$1" ]; then
    echo "Usage: $0 <worker-id>"
    echo "Example: $0 worker-1"
    exit 1
fi

WORKER_ID=$1

echo "🔍 Buscando proceso del worker: $WORKER_ID"

# Buscar el PID del worker por su WORKER_ID (variable de entorno)
PID=$(ps aux | grep "[b]in/worker" | grep -v grep | awk '{print $2}' | head -1)

if [ -z "$PID" ]; then
    echo "❌ No se encontró ningún worker en ejecución"
    exit 1
fi

echo "💀 Matando worker $WORKER_ID (PID: $PID)..."
kill -9 $PID

if [ $? -eq 0 ]; then
    echo "✅ Worker $WORKER_ID eliminado exitosamente"
    echo "📊 Puedes verificar con: curl http://localhost:8080/workers"
else
    echo "❌ Error al eliminar el worker"
    exit 1
fi