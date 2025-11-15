#!/bin/bash
set -euo pipefail

SCRIPT_DIR="./"
REPO_ROOT="./"

cd "$REPO_ROOT"

echo "Ensuring Go dependencies are downloaded..."
go mod download

# mapfile -t GO_FILES < <(git ls-files '*.go')
# if ((${#GO_FILES[@]} > 0)); then
# 	fmt_out=$(gofmt -l "${GO_FILES[@]}")
# 	if [[ -n "${fmt_out}" ]]; then
# 		echo "gofmt found unformatted files:"
# 		echo "${fmt_out}"
# 		exit 1
# 	fi
# fi

echo "Running go vet..."
VET_TARGETS=(
	./internal/...
)
go vet "${VET_TARGETS[@]}"

echo "Checking cyclomatic complexity (max 10)..."
cyclo_output=$(go run github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0 ./internal | grep -v '/tests/' | awk '$1+0 > 9')
if [[ -n "${cyclo_output}" ]]; then
	echo "Cyclomatic complexity violations:"
	echo "${cyclo_output}"
	exit 1
fi

