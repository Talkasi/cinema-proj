#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

cd "$REPO_ROOT"

echo "Ensuring Go dependencies are downloaded..."
go mod download

echo "Running go vet..."
VET_TARGETS=(
	./internal/...
)
go vet "${VET_TARGETS[@]}"

echo "Checking cyclomatic complexity (max 10)..."
cyclo_output=$(go run github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0 ./internal | grep -v '/tests/' | awk '$1+0 > 10')
if [[ -n "${cyclo_output}" ]]; then
	echo "Cyclomatic complexity violations:"
	echo "${cyclo_output}"
	exit 1
fi
