#!/usr/bin/env bash
set -euo pipefail

BENCH_NAME="BenchmarkHTTPTracingDisabledLoggingDefault"
LABEL="http-tracing-disabled-log-default"
BENCH_COUNT="${BENCH_COUNT:-20}"
OUT_DIR="${OUT_DIR:-artifacts/observability}"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
OUT_FILE="${OUT_DIR}/${LABEL}-${TIMESTAMP}.txt"

echo "Запуск ${BENCH_NAME} (${BENCH_COUNT} прогонов)..."
mkdir -p "${OUT_DIR}"

go test -count="${BENCH_COUNT}" ./tests/observability -bench "^${BENCH_NAME}$" -benchmem -run Benchmark | tee "${OUT_FILE}"

echo "Результаты сохранены в ${OUT_FILE}"
