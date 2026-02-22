#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ "${DRY_RUN:-0}" = "1" ]; then
  echo "[DRY-RUN] docker compose up -d --build"
  echo "[DRY-RUN] repeat 100x: ${SCRIPT_DIR}/benchmark_single.sh <n>"
  echo "[DRY-RUN] docker compose down -v"
  exit 0
fi

echo "Starting Docker environment..."
docker compose up -d --build

echo "Waiting for services to become ready... (30 seconds)"
sleep 30

for i in $(seq 1 100); do
  bash "${SCRIPT_DIR}/benchmark_single.sh" "$i"
  echo "Pause 10 seconds before the next run for system stabilization..."
  sleep 10
done

echo "All test runs completed. Stopping environment..."
docker compose down -v

echo "Benchmarking completed."
