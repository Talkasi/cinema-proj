#!/usr/bin/env bash
set -euo pipefail

OUT_DIR="${OUT_DIR:-artifacts/observability}"
LABEL="${LABEL:-observability}"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"

mkdir -p "${OUT_DIR}"
OUT_FILE="${OUT_DIR}/${LABEL}-${TIMESTAMP}.txt"

go test ./tests/observability -bench . -benchmem -run Benchmark > "${OUT_FILE}"

echo "Reports saved to ${OUT_FILE}"
