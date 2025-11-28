#!/bin/bash

# Monitor en tiempo real del cluster
BASE_URL="http://localhost:8080"

while true; do
  clear
  echo "╔════════════════════════════════════════════════════════╗"
  echo "║        Mini-Spark Cluster Monitor                     ║"
  echo "╚════════════════════════════════════════════════════════╝"
  echo ""
  
  # Health
  echo "📊 SYSTEM HEALTH"
  curl -s ${BASE_URL}/health | jq -r '"  Status: \(.status) | Workers: \(.active_workers)/\(.total_workers)"'
  echo ""
  
  # Workers
  echo "👷 WORKERS"
  curl -s ${BASE_URL}/api/v1/workers | jq -r '.workers[] | "  [\(if .Status == "ACTIVE" then "✓" else "✗" end)] \(.ID) - \(.Status) (Tasks: \(.ActiveTasks), Mem: \(.MemoryUsageMB)MB)"'
  echo ""
  
  # Stats
  echo "📈 STATISTICS"
  curl -s ${BASE_URL}/api/v1/workers | jq -r '.stats | "  Active Tasks: \(.total_active_tasks)\n  Avg CPU: \(.avg_cpu_usage)%\n  Total Memory: \(.total_memory_mb)MB"'
  echo ""
  
  echo "Press Ctrl+C to exit | Refreshing in 3s..."
  sleep 3
done
