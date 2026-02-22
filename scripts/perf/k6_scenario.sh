#!/usr/bin/env bash

set -euo pipefail

SCENARIO_NAME="$1"
RESULTS_PREFIX="$2"
K6_SCRIPT_PATH="$3"
ANALYZE_MODE="${4:-copy}"

if [ -z "$SCENARIO_NAME" ] || [ -z "$RESULTS_PREFIX" ] || [ -z "$K6_SCRIPT_PATH" ]; then
  echo "Usage: $0 <scenario-name> <results-prefix> <k6-script-path> [analysis-mode]"
  echo "analysis-mode: copy | copy+degradation | none"
  exit 1
fi

echo "--- ${SCENARIO_NAME} ---"

RESULTS_DIR="${RESULTS_PREFIX}_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$RESULTS_DIR"

echo "Results will be saved to: ${RESULTS_DIR}"

if [ "${DRY_RUN:-0}" = "1" ]; then
  echo "[DRY-RUN] docker compose up -d --build"
  echo "[DRY-RUN] curl http://localhost:8080/api/v1/movies?limit=1"
  echo "[DRY-RUN] docker create grafana/k6:0.49.0 ...; docker cp ${K6_SCRIPT_PATH}; docker start"
  echo "[DRY-RUN] analysis mode: ${ANALYZE_MODE}"
  echo "[DRY-RUN] docker compose down"
  exit 0
fi

echo "Starting Docker environment..."
docker compose up -d --build

echo "Waiting for services to become ready... (30 seconds)"
sleep 30

echo "Checking API availability..."
if curl -f http://localhost:8080/api/v1/movies?limit=1 > /dev/null 2>&1; then
  echo "API is available"
else
  echo "API is unavailable"
  docker compose logs app
  exit 1
fi

echo "Starting resource monitoring..."
STATS_FILE="${RESULTS_DIR}/resource_usage.csv"
echo "timestamp,container,cpu_percent,mem_usage,mem_percent,net_io,block_io,pids" > "$STATS_FILE"

(
  while true; do
    TIMESTAMP=$(date +%s%3N)
    docker stats --no-stream --format \
      "$TIMESTAMP,{{.Name}},{{.CPUPerc}},{{.MemUsage}},{{.MemPerc}},{{.NetIO}},{{.BlockIO}},{{.PIDs}}" \
      go_app postgres_db >> "$STATS_FILE" 2>/dev/null
    sleep 2
  done
) &
STATS_PID=$!

echo "Starting k6 scenario: ${K6_SCRIPT_PATH}"
CONTAINER_ID=$(docker create \
  --network cinema-proj_app-net \
  --cpuset-cpus "10-11" \
  grafana/k6:0.49.0 \
  run /test.js \
  --out json=/tmp/k6_results.json \
  --out csv=/tmp/k6_results.csv)

docker cp "$K6_SCRIPT_PATH" "$CONTAINER_ID:/test.js"

echo "Starting test..."
docker start -a "$CONTAINER_ID"
K6_EXIT_CODE=$?

echo "Copying results..."
docker cp "$CONTAINER_ID:/tmp/k6_results.json" "${RESULTS_DIR}/" && echo "JSON results copied"
docker cp "$CONTAINER_ID:/tmp/k6_results.csv" "${RESULTS_DIR}/" && echo "CSV results copied"

docker rm "$CONTAINER_ID" > /dev/null 2>&1

echo "Stopping monitoring..."
kill "$STATS_PID" 2>/dev/null
wait "$STATS_PID" 2>/dev/null

if [ -f "${RESULTS_DIR}/k6_results.json" ] && [ "$K6_EXIT_CODE" -eq 0 ]; then
  echo ">>>>>>>>>>>>>> Test completed successfully"
  if [ "$ANALYZE_MODE" != "none" ]; then
    echo "Analyzing results..."
    if [ -f "./scripts/analysis/report.py" ]; then
      python3 ./scripts/analysis/report.py copy "$RESULTS_DIR"
    else
      echo ">>>>>>>>>>>>>>>> scripts/analysis/report.py not found"
    fi
  fi

  if [ "$ANALYZE_MODE" = "copy+degradation" ]; then
    if [ -f "./scripts/analysis/report.py" ]; then
      python3 ./scripts/analysis/report.py degradation "$RESULTS_DIR"
    else
      echo ">>>>>>>>>>>>>>>> scripts/analysis/report.py not found"
    fi
  fi
else
  echo ">>>>>>>>>>>>>>>>>>> Test finished with error (code: $K6_EXIT_CODE)"
  echo "--- Latest application logs ---"
  docker compose logs app --tail=20
fi

echo "Stop Docker environment..."
docker compose down

echo "--- ${SCENARIO_NAME} completed ---"
if [ -f "${RESULTS_DIR}/k6_results.json" ]; then
  echo ">>>>>>>>>>> Results in: ${RESULTS_DIR}"
  ls -la "$RESULTS_DIR/"
else
  echo ">>>>>>>>>>> Test did not complete."
fi
