#!/bin/bash

# Script de demostración completo de Mini-Spark
# Muestra: WordCount, operadores avanzados, fallo simulado y recuperación

set -e

echo "=========================================="
echo "  Mini-Spark Demo - Sistema Completo"
echo "=========================================="

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"
MASTER_PID=""
WORKER1_PID=""
WORKER2_PID=""

# Función para limpiar procesos al salir
cleanup() {
    echo -e "\n${YELLOW}[CLEANUP]${NC} Stopping all processes..."
    [ ! -z "$WORKER1_PID" ] && kill $WORKER1_PID 2>/dev/null || true
    [ ! -z "$WORKER2_PID" ] && kill $WORKER2_PID 2>/dev/null || true
    [ ! -z "$MASTER_PID" ] && kill $MASTER_PID 2>/dev/null || true
    sleep 1
    echo -e "${GREEN}✓ Cleanup complete${NC}"
}

trap cleanup EXIT

# Verificar que los binarios existen
if [ ! -f "bin/master" ] || [ ! -f "bin/worker" ] || [ ! -f "bin/client" ]; then
    echo -e "${RED}ERROR: Binaries not found. Run 'make build' first${NC}"
    exit 1
fi

# ====================
# 1. Iniciar Cluster
# ====================
echo -e "\n${GREEN}[STEP 1]${NC} Starting Mini-Spark cluster..."

echo "  → Starting Master..."
./bin/master > /tmp/master.log 2>&1 &
MASTER_PID=$!
sleep 3

echo "  → Starting Worker 1..."
WORKER_ID=demo-worker-1 ./bin/worker > /tmp/worker1.log 2>&1 &
WORKER1_PID=$!
sleep 2

echo "  → Starting Worker 2..."
WORKER_ID=demo-worker-2 ./bin/worker > /tmp/worker2.log 2>&1 &
WORKER2_PID=$!
sleep 2

# Verificar cluster
echo -e "\n  ${YELLOW}Cluster Status:${NC}"
curl -s ${BASE_URL}/health | jq -r '"  Status: \(.status) | Workers: \(.active_workers)/\(.total_workers)"'

# ====================
# 2. Demo: WordCount
# ====================
echo -e "\n${GREEN}[STEP 2]${NC} Demo: WordCount on CSV data"
echo "  → Submitting WordCount job..."

JOB_OUTPUT=$(./bin/client submit examples/wordcount.json)
JOB_ID=$(echo "$JOB_OUTPUT" | grep "Job ID" | awk '{print $4}')

echo "  Job submitted: $JOB_ID"
sleep 2

echo "  → Job status:"
curl -s ${BASE_URL}/api/v1/jobs/${JOB_ID} | jq -r '"  Status: \(.status) | Progress: \(.progress)% | Tasks: \(.completed_tasks)/\(.total_tasks)"'

sleep 3

# ====================
# 3. Demo: Operadores Avanzados
# ====================
echo -e "\n${GREEN}[STEP 3]${NC} Demo: Advanced operators (Filter + Map)"

if [ -f "examples/filter_map.json" ]; then
    echo "  → Submitting filter+map pipeline..."
    JOB2_OUTPUT=$(./bin/client submit examples/filter_map.json)
    JOB2_ID=$(echo "$JOB2_OUTPUT" | grep "Job ID" | awk '{print $4}')
    echo "  Job submitted: $JOB2_ID"
    sleep 3
fi

# ====================
# 4. Demo: Fallo Simulado
# ====================
echo -e "\n${GREEN}[STEP 4]${NC} Demo: Simulated Worker Failure & Recovery"

echo "  → Current workers:"
curl -s ${BASE_URL}/api/v1/workers | jq -r '.workers[] | "  - \(.ID): \(.Status)"'

echo -e "\n  ${RED}→ Killing Worker 1 to simulate failure...${NC}"
kill $WORKER1_PID 2>/dev/null || true
WORKER1_PID=""
sleep 2

echo "  → Waiting for master to detect failure (10s timeout)..."
sleep 12

echo "  → Workers after failure:"
curl -s ${BASE_URL}/api/v1/workers | jq -r '.workers[] | "  - \(.ID): \(.Status)"'

echo -e "\n  ${YELLOW}→ Submitting new job (should use only Worker 2)...${NC}"
JOB3_OUTPUT=$(./bin/client submit examples/wordcount.json)
JOB3_ID=$(echo "$JOB3_OUTPUT" | grep "Job ID" | awk '{print $4}')
echo "  Job submitted: $JOB3_ID"
sleep 3

echo "  → Job status (running on remaining worker):"
curl -s ${BASE_URL}/api/v1/jobs/${JOB3_ID} | jq -r '"  Status: \(.status) | Progress: \(.progress)%"'

# ====================
# 5. Demo: Recuperación
# ====================
echo -e "\n${GREEN}[STEP 5]${NC} Demo: Worker Recovery"

echo "  → Restarting Worker 1..."
WORKER_ID=demo-worker-1 ./bin/worker > /tmp/worker1_recovered.log 2>&1 &
WORKER1_PID=$!
sleep 3

echo "  → Workers after recovery:"
curl -s ${BASE_URL}/api/v1/workers | jq -r '.workers[] | "  - \(.ID): \(.Status)"'

# ====================
# 6. Estadísticas Finales
# ====================
echo -e "\n${GREEN}[STEP 6]${NC} Final Statistics"

echo -e "\n  ${YELLOW}All Jobs:${NC}"
curl -s ${BASE_URL}/api/v1/jobs | jq -r '.jobs[] | "  Job \(.job_id): \(.status) - \(.name)"'

echo -e "\n  ${YELLOW}Worker Statistics:${NC}"
curl -s ${BASE_URL}/api/v1/workers | jq '.stats'

echo -e "\n  ${YELLOW}System Health:${NC}"
curl -s ${BASE_URL}/health | jq '.'

# ====================
# Fin del Demo
# ====================
echo -e "\n=========================================="
echo -e "${GREEN}✓ Demo Complete!${NC}"
echo "=========================================="
echo ""
echo "Logs available at:"
echo "  - Master:   /tmp/master.log"
echo "  - Worker 1: /tmp/worker1.log"
echo "  - Worker 2: /tmp/worker2.log"
echo ""
echo "Press Enter to stop cluster..."
read

# Cleanup se ejecuta automáticamente por trap EXIT