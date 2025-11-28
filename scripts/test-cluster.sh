#!/bin/bash

echo "====================================="
echo "  Mini-Spark Cluster Tests - Week 1"
echo "====================================="

BASE_URL="http://localhost:8080"

# Test 1: Health Check
echo -e "\n[TEST 1] Health Check"
curl -s ${BASE_URL}/health | jq -r '"Status: \(.status) | Active: \(.active_workers)/\(.total_workers)"'

# Test 2: List Workers 
echo -e "\n[TEST 2] Workers Details"
curl -s ${BASE_URL}/api/v1/workers | jq -r '.workers[] | "ID: \(.ID)\n  Status: \(.Status)\n  Host: \(.Host):\(.Port)\n  Tasks: \(.ActiveTasks)/\(.Capacity)\n  Memory: \(.MemoryUsageMB)MB\n  Last Heartbeat: \(.LastHeartbeat)\n"'

# Test 3: Worker Statistics
echo -e "\n[TEST 3] Statistics"
curl -s ${BASE_URL}/api/v1/workers | jq '.stats'

# Test 4: Active Workers Only
echo -e "\n[TEST 4] Active Workers"
curl -s ${BASE_URL}/api/v1/workers | jq -r '[.workers[] | select(.Status == "ACTIVE")] | length | "Active workers: \(.)"'

# Test 5: Worker Status Summary
echo -e "\n[TEST 5] Status Summary"
curl -s ${BASE_URL}/api/v1/workers | jq '[.workers[] | .Status] | group_by(.) | map({status: .[0], count: length})'

echo -e "\n====================================="
echo "  Tests Completed!"
echo "====================================="
