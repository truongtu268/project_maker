#!/bin/bash
set -e

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
  echo "Error: Docker is not running or not installed. Please start Docker and try again."
  exit 1
fi

echo "Running integration tests..."
go test -v ./test/integration/...

echo "Integration tests completed successfully!" 