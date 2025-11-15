#!/bin/bash

echo "Running golangci-lint..."
chmod 777 run-static-analysis.sh
if ! ./run-static-analysis.sh; then
    echo "ERROR: golangci-lint found issues or had configuration problems"
    exit 1
fi

echo "All static analysis checks passed!"
exit 0