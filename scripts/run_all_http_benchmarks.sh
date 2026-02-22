#!/usr/bin/env bash
set -euo pipefail

BENCH_COUNT="${BENCH_COUNT:-20}"
OUT_DIR="${OUT_DIR:-artifacts/observability}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

BENCH_SCRIPTS=(
  "run_bench_http_tracing_disabled_logging_default.sh"
  "run_bench_http_tracing_disabled_logging_extended.sh"
  "run_bench_http_tracing_enabled_logging_default.sh"
  "run_bench_http_tracing_enabled_logging_extended.sh"
)

echo "Запуск всех бенчмарков (BENCH_COUNT=${BENCH_COUNT}, OUT_DIR=${OUT_DIR})..."
for script in "${BENCH_SCRIPTS[@]}"; do
  echo ""
  echo ">>> ${script}"
  BENCH_COUNT="${BENCH_COUNT}" OUT_DIR="${OUT_DIR}" "${SCRIPT_DIR}/${script}"
done

echo ""
echo "Все бенчмарки выполнены. Итоги в каталоге ${OUT_DIR}"
