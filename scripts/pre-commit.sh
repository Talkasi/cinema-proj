#!/bin/bash

echo "Running golangci-lint..."
chmod 777 ./scripts/tools/static_analysis.sh
if ! ./scripts/tools/static_analysis.sh; then
    echo "ERROR: golangci-lint found issues or had configuration problems"
    exit 1
fi

echo "Running fast repository checks..."
if ! make verify-fast; then
    echo "ERROR: verify-fast checks failed"
    exit 1
fi

echo "All static analysis checks passed!"
exit 0
