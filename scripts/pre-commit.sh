#!/bin/bash

echo "Running golangci-lint..."
if ! golangci-lint run --config .golangci.yml --timeout=5m; then
    echo "ERROR: golangci-lint found issues or had configuration problems"
    exit 1
fi

echo "All static analysis checks passed!"
exit 0