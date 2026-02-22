#!/usr/bin/env bash

set -euo pipefail

# Check that the run number was provided
if [ -z "$1" ]; then
  echo "Error: provide a run number. Example: ./benchmark_single.sh 1"
  exit 1
fi

RUN_ID="$1"
RESULTS_DIR="run_${RUN_ID}"
K6_SCRIPT_PATH="/scripts/test.js" # Path inside k6 container

# Create output directory for this run
mkdir -p "$RESULTS_DIR"

echo "--- Starting run #${RUN_ID} ---"

if [ "${DRY_RUN:-0}" = "1" ]; then
  echo "[DRY-RUN] docker stats collector -> ${RESULTS_DIR}/resource_usage.csv"
  echo "[DRY-RUN] docker compose run --rm k6 run ${K6_SCRIPT_PATH} --out json=/scripts/performance_results.json"
  echo "[DRY-RUN] docker cp k6_runner:/scripts/performance_results.json ${RESULTS_DIR}/performance_results.json"
  exit 0
fi

# --- STEP 1: Start background resource stats collection ---
echo "Starting docker stats collection..."
STATS_FILE="${RESULTS_DIR}/resource_usage.csv"
echo "timestamp,container_name,cpu_percent,mem_usage" > "$STATS_FILE"

(while true; do
  TIMESTAMP=$(date +%s)
  # Collect stats only for go_app and postgres_db containers
  docker stats --no-stream --format "$TIMESTAMP,{{.Name}},{{.CPUPerc}},{{.MemUsage}}" go_app postgres_db >> "$STATS_FILE"
  sleep 1
done) &
STATS_PID=$!

# --- STEP 2: Run the k6 test and wait for completion ---
echo "Starting k6 test..."
# Use 'docker compose run --rm' to create a temporary container for the test
docker compose run --rm k6 run "$K6_SCRIPT_PATH" --out json="/scripts/performance_results.json"
# Copy result from container to host
docker cp k6_runner:/scripts/performance_results.json "${RESULTS_DIR}/performance_results.json"


# --- STEP 3: Stop stats collection ---
echo "Stopping docker stats collection..."
kill "$STATS_PID"

echo "--- Run #${RUN_ID} completed. Results are in ${RESULTS_DIR} ---"
